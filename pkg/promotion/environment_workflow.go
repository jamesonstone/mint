package promotion

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// RenderEnvironmentWorkflow emits repository-native requests and ambient
// observation triggers. Provider-specific build, observe and deploy jobs remain
// separately configured adapters with their own permissions.
func RenderEnvironmentWorkflow(policy Policy, mintRef, configPath string, buildNames ...string) (string, error) {
	if policy.Schema != 2 || !shaPattern.MatchString(mintRef) || !SafeRepositoryPath(configPath) {
		return "", fmt.Errorf("schema 2 policy, immutable Mint commit and safe config path required")
	}
	names := make([]string, 0, len(policy.Environments))
	for name := range policy.Environments {
		names = append(names, name)
	}
	sort.Strings(names)
	options, _ := json.Marshal(names)
	if len(names) == 0 {
		return "", fmt.Errorf("configured environments required")
	}
	callback := ""
	if len(buildNames) > 0 {
		unique := []string{}
		seen := map[string]bool{}
		for _, name := range buildNames {
			if name == "" || strings.Contains(name, "${{") || strings.ContainsAny(name, "\r\n") {
				return "", fmt.Errorf("invalid callback workflow name")
			}
			if !seen[name] {
				literal := strings.NewReplacer("\\", "\\\\", "*", "\\*", "?", "\\?", "+", "\\+", "!", "\\!", "[", "\\[", "]", "\\]").Replace(name)
				unique = append(unique, literal)
				seen[name] = true
			}
		}
		callbacks, _ := json.Marshal(unique)
		callback = fmt.Sprintf("  workflow_run:\n    workflows: %s\n    types: [completed]\n", string(callbacks))
	}
	return fmt.Sprintf(`name: Mint environment control
on:
  push:
    branches: ['%s']
%s  pull_request_target:
    types: [closed]
  workflow_dispatch:
    inputs:
      environment:
        description: Environment to inspect or change
        type: choice
        options: %s
        required: true
      operation:
        description: Request action (prefer a roll-forward hotfix)
        type: choice
        options: [promote, hotfix, rollback, resume, observe, reconcile]
        default: promote
        required: true
      target_version:
        description: Exact version for promotion or rollback; blank selects eligible latest or previous verified
        type: string
        required: false
      intent_id:
        description: Locked intent to reconcile; blank selects the current lock
        type: string
        required: false
      observation_run_id:
        description: Successful fresh observer run; blank selects the latest recorded observation
        type: string
        required: false
      fix_pr:
        description: Merged reviewed fix PR for an isolated hotfix
        type: string
        required: false
      issue:
        description: Issue for a newly authored isolated hotfix
        type: string
        required: false
      reason:
        description: Reason for the request
        type: string
        required: false
permissions:
  contents: write
  actions: write
  pull-requests: write
  issues: write
  checks: read
concurrency:
  group: mint-environment-controller-${{ github.run_id }}
  cancel-in-progress: false
jobs:
  control:
    if: vars.MINT_RELEASE_ENABLED == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@11d5960a326750d5838078e36cf38b85af677262
        with:
          ref: ${{ github.event.repository.default_branch }}
          fetch-depth: 0
          persist-credentials: false
      - name: Configure authenticated Git fetch
        env:
          GH_TOKEN: ${{ github.token }}
        run: gh auth setup-git
      - uses: jamesonstone/mint@%s
        with:
          command: deployment-control
          production-config: '%s'
          github-token: ${{ github.token }}
`, policy.DefaultBranch, callback, string(options), mintRef, configPath), nil
}
