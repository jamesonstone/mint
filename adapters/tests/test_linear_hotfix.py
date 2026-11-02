"""Linear attribution never falls back to GitHub issues or weakens native authorization."""
import importlib.util
import json
import os
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

path = Path(__file__).resolve().parents[1] / "mint-control.py"
spec = importlib.util.spec_from_file_location("linear_control", path)
control = importlib.util.module_from_spec(spec)
spec.loader.exec_module(control)


class LinearHotfixTests(unittest.TestCase):
    def test_dispatch_serializes_ticket_without_any_github_issue_call(self):
        with tempfile.TemporaryDirectory() as temporary:
            env = {"CONTROL_OPERATION":"hotfix", "HOTFIX_TICKET_URL":"https://linear.app/acme/issue/ENG-42", "HOTFIX_BASELINE_ID":"verified", "HOTFIX_FIXES":"a" * 40, "HOTFIX_REASON":"Restore login", "GITHUB_RUN_ID":"7", "RUNNER_TEMP":temporary}
            with patch.dict(os.environ, env, clear=True), patch.object(control,"gh", side_effect=AssertionError("GitHub issue access")), patch.object(control,"mint") as mint:
                control.reconcile({}, {})
                request = Path(temporary) / "hotfix-request.json"
                self.assertEqual(json.loads(request.read_text()), {"Issue":0,"ticket_url":env["HOTFIX_TICKET_URL"],"BaselineID":"verified","Fixes":["a" * 40]})
                mint.assert_called_once_with("prepare-hotfix","--input",str(request),"--run-id","7")

    def test_bad_or_mixed_provider_input_never_calls_native_mutation(self):
        env = {"CONTROL_OPERATION":"hotfix", "HOTFIX_TICKET_URL":"https://linear.app/acme/issue/ENG-42", "REQUEST_ISSUE":"123"}
        with patch.dict(os.environ,env,clear=True), patch.object(control,"mint") as mint:
            with self.assertRaises(ValueError): control.reconcile({}, {})
            mint.assert_not_called()

    def test_external_source_branch_keeps_component_boundary(self):
        config = {"Component":{"name":"backend"}}
        self.assertTrue(control.hotfix_base_matches(config,"mint-hotfix-base/backend/ENG-42"))
        self.assertFalse(control.hotfix_base_matches(config,"mint-hotfix-base/frontend/ENG-42"))
        self.assertFalse(control.hotfix_base_matches(config,"mint-hotfix-base/backend/ENG-42/other"))

    def test_wrong_workspace_or_team_cannot_prepare(self):
        env = {"HOTFIX_TICKET_URL":"https://linear.app/other/issue/ENG-42", "HOTFIX_TICKET_WORKSPACE":"acme", "HOTFIX_TICKET_TEAM":"ENG"}
        with patch.dict(os.environ,env,clear=True), patch.object(control,"mint") as mint:
            with self.assertRaises(ValueError): control.request_linear_hotfix()
            mint.assert_not_called()
