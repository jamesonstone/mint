"""Adapter tests execute the exact production helper code without cloud calls."""
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

ROOT = Path(__file__).resolve().parents[2]


def module(name, filename):
    spec = importlib.util.spec_from_file_location(name, ROOT / 'adapters' / filename)
    result = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(result)
    return result


MANIFEST = module('manifest', 'mint-manifest.py')
CONTROL = module('control', 'mint-control.py')


class AdapterTests(unittest.TestCase):
    def config(self):
        return {'Repository':'owner/repo', 'DefaultBranch':'main', 'BuildWorkflow':'.github/workflows/build.yaml', 'PromotionWorkflow':'.github/workflows/promote.yaml', 'ControlPaths':['CHANGELOG.md']}

    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.previous = os.getcwd()
        os.chdir(self.temporary.name)
        self.addCleanup(os.chdir, self.previous)
        Path('template.json').write_text('{"image":"immutable"}')
        Path('output').touch()
        self.environment = patch.dict(os.environ, {
            'RELEASE_VARIABLES': '{"FEATURE":true,"REGION":"verified"}',
            'RELEASE_CONFIG_FILES': '["template.json"]', 'SOURCE_SHA': 'a' * 40,
            'VERSION_TAG': 'v1.2.3', 'ARTIFACT_DIGEST': 'sha256:' + 'b' * 64,
            'ARTIFACT_REFERENCE': 'immutable/object', 'GITHUB_REPOSITORY': 'owner/repo',
            'GITHUB_RUN_ID': '7', 'GITHUB_ACTOR': 'owner', 'GITHUB_OUTPUT': 'output',
            'RUNNER_TEMP': self.temporary.name, 'GITHUB_EVENT_NAME': 'workflow_run',
            'RELEASE_ENVIRONMENT': 'production', 'DEPLOYMENT_VERIFIED': 'true', 'CONTROL_OPERATION': 'reconcile', 'INTENT_ID': 'reviewed:merge',
        }, clear=True)
        self.environment.start()
        self.addCleanup(self.environment.stop)

    def test_configuration_is_canonical_and_fences_both_variables_and_template(self):
        expected = MANIFEST.configuration()
        os.environ['RELEASE_VARIABLES'] = '{"REGION":"verified","FEATURE":true}'
        self.assertEqual(MANIFEST.configuration(), expected)
        Path('template.json').write_text('{"image":"different"}')
        self.assertNotEqual(MANIFEST.configuration(), expected)
        os.environ['EXPECTED_CONFIGURATION'] = expected
        with patch('sys.argv', ['adapter', 'verify-configuration']), self.assertRaises(ValueError):
            MANIFEST.main()

    def test_candidate_has_exact_source_artifact_and_config(self):
        candidate = MANIFEST.candidate()
        self.assertEqual(candidate['source_sha'], 'a' * 40)
        self.assertEqual(candidate['artifact']['digest'], 'sha256:' + 'b' * 64)
        self.assertEqual(candidate['artifact']['configuration_sha256'], MANIFEST.configuration())
        os.environ['SOURCE_SHA'] = 'main'
        with self.assertRaises(ValueError):
            MANIFEST.candidate()

    def test_candidate_cli_writes_exact_manifest(self):
        subprocess.run(
            [sys.executable, str(ROOT / 'adapters/mint-manifest.py'), 'candidate', 'candidate.json'],
            check=True, capture_output=True, text=True,
        )
        candidate = json.loads(Path('candidate.json').read_text())
        self.assertEqual(candidate['repository'], 'owner/repo')
        self.assertEqual(candidate['source_sha'], 'a' * 40)
        self.assertEqual(candidate['version'], 'v1.2.3')
        self.assertEqual(candidate['run_id'], 7)
        self.assertEqual(candidate['artifact']['reference'], 'immutable/object')
        self.assertEqual(candidate['artifact']['digest'], 'sha256:' + 'b' * 64)
        self.assertEqual(candidate['artifact']['configuration_sha256'], MANIFEST.configuration())

    def test_deployment_cannot_attest_changed_configuration(self):
        os.environ['EXPECTED_CONFIGURATION'] = 'sha256:' + '0' * 64
        with patch('sys.argv', ['adapter', 'deployment', 'result.json']), self.assertRaises(ValueError):
            MANIFEST.main()
        self.assertFalse(Path('result.json').exists())

    def test_completed_failure_never_publishes_or_advances_to_latest(self):
        self.completed('failure', ['finish', 'report'])

    def test_completed_success_records_before_publication_and_status_review(self):
        self.completed('success', ['finish', 'publish', 'report'])

    def test_rollback_completion_never_publishes_another_release(self):
        self.completed('success', ['finish', 'report'], 'rollback')

    def completed(self, conclusion, expected, kind='normal'):
        sha = 'a' * 40
        journal = {'intents': {'intent': {'id': 'reviewed:merge', 'merge_sha': sha, 'deployment_run_id': 7, 'kind': kind}}}
        event = {'workflow_run': {'name': 'Mint Production', 'head_sha': sha, 'id': 7, 'conclusion': conclusion}}
        commands = []
        event['workflow_run'].update(path=self.config()['PromotionWorkflow'], repository={'full_name':'owner/repo'}, head_repository={'full_name':'owner/repo'}, status='completed')
        with patch.object(CONTROL, 'gh', return_value=event['workflow_run']), patch.object(CONTROL.subprocess, 'check_output', return_value=json.dumps(journal).encode()), patch.object(CONTROL, 'mint', side_effect=lambda cmd, *args: commands.append((cmd, args))):
            CONTROL.reconcile(event, self.config())
        self.assertEqual([cmd for cmd, _ in commands], expected)
        self.assertTrue(all('reviewed:merge' in args for _, args in commands))
        self.assertNotIn('propose', [cmd for cmd, _ in commands])

    def test_foreign_completion_does_not_select_any_latest_intent(self):
        event = {'workflow_run': {'name': 'Mint Production', 'head_sha': 'b' * 40, 'id': 8, 'conclusion': 'success'}}
        event['workflow_run'].update(path=self.config()['PromotionWorkflow'], repository={'full_name':'owner/repo'}, head_repository={'full_name':'owner/repo'}, status='completed')
        with patch.object(CONTROL, 'gh', return_value=event['workflow_run']), patch.object(CONTROL.subprocess, 'check_output', return_value=b'{"intents":{}}'), patch.object(CONTROL, 'mint') as mint:
            CONTROL.reconcile(event, self.config())
            mint.assert_not_called()

    def test_reviewed_hotfix_merge_dispatches_trusted_main_build(self):
        os.environ['GITHUB_EVENT_NAME'] = 'pull_request_target'
        event = {'pull_request': {'number': 42, 'merged': True, 'body': '<!-- mint:hotfix-source:GH-1:base -->'}, 'repository': {'default_branch': 'main'}}
        Path('.mint.yaml').write_text('build_workflow: .github/workflows/build.yaml\n')
        event['pull_request'].update(head={'repo':{'full_name':'owner/repo'}}, base={'repo':{'full_name':'owner/repo'}})
        with patch.object(CONTROL, 'gh', return_value=event['pull_request']), patch.object(CONTROL.subprocess, 'run') as run:
            CONTROL.reconcile(event, self.config())
            run.assert_called_once_with(['gh', 'workflow', 'run', 'build.yaml', '--ref', 'main', '-f', 'hotfix_pr=42'], check=True)

    def test_manifest_requires_provider_verification(self):
        os.environ.pop('DEPLOYMENT_VERIFIED')
        with patch('sys.argv', ['adapter', 'deployment', 'result.json']), self.assertRaises(ValueError):
            MANIFEST.main()
        self.assertFalse(Path('result.json').exists())

    def test_configuration_exclusions_are_caller_owned(self):
        os.environ['RELEASE_VARIABLES'] = '{"TEAM_FLAG":true}'
        before = MANIFEST.configuration()
        os.environ['RELEASE_CONFIG_EXCLUDE_VARIABLES'] = '["TEAM_FLAG"]'
        self.assertNotEqual(before, MANIFEST.configuration())
        os.environ['RELEASE_VARIABLES'] = '{"TEAM_FLAG":false}'
        excluded = MANIFEST.configuration()
        os.environ['RELEASE_VARIABLES'] = '{"TEAM_FLAG":true,"MINT_RELEASE_ENABLED":true}'
        self.assertEqual(excluded, MANIFEST.configuration())

    def test_output_injection_writes_nothing(self):
        with self.assertRaises(ValueError):
            CONTROL.outputs({'safe':'value', 'unsafe':'value\nnew_output=true'})
        self.assertEqual(Path('output').read_text(), '')

    def test_callback_uses_live_workflow_identity(self):
        event = {'workflow_run': {'id':7, 'name':'Mint Production'}}
        live = {'id':7, 'path':'.github/workflows/foreign.yaml', 'repository':{'full_name':'owner/repo'}, 'head_repository':{'full_name':'owner/repo'}, 'status':'completed', 'conclusion':'success'}
        with patch.object(CONTROL, 'gh', return_value=live), patch.object(CONTROL, 'mint') as mint:
            CONTROL.reconcile(event, self.config())
            mint.assert_not_called()
        live['repository']['full_name'] = 'foreign/repo'
        with patch.object(CONTROL, 'gh', return_value=live), patch.object(CONTROL, 'mint') as mint:
            with self.assertRaises(ValueError):
                CONTROL.reconcile(event, self.config())
            mint.assert_not_called()

    def test_configuration_does_not_follow_source_links(self):
        Path('template.json').unlink()
        Path('template.json').symlink_to('/etc/hosts')
        with self.assertRaises(ValueError):
            MANIFEST.configuration()
        os.environ['RELEASE_CONFIG_FILES'] = '["../outside.json"]'
        with self.assertRaises(ValueError):
            MANIFEST.configuration()

    def test_hotfix_form_rejects_non_exact_commit_input(self):
        body = '### Verified production baseline ID\nbase\n### Reviewed fix SHAs\nmain; deploy\n### Reason and validation\nfix'
        with patch.object(CONTROL, 'gh', return_value={'number': 1, 'body': body}), patch.object(CONTROL, 'mint') as mint:
            with self.assertRaises(ValueError):
                CONTROL.request_hotfix('1')
            mint.assert_not_called()


if __name__ == '__main__':
    unittest.main()
