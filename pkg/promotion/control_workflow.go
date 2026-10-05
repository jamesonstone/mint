package promotion

import (
	"fmt"
	"regexp"
)

// RenderControlWorkflow generates the repository-native request entry point.
// It uses trusted default-branch code and an immutable feature-bearing Mint pin.
func RenderControlWorkflow(cfg Config, mintRef string) (string, error) {
	if !shaPattern.MatchString(mintRef) {
		return "", fmt.Errorf("use a published feature-bearing Mint commit SHA")
	}
	if !regexp.MustCompile(`^[A-Za-z0-9-]+$`).MatchString(cfg.HumanLogin) {
		return "", fmt.Errorf("invalid control operator login")
	}
	return fmt.Sprintf(`name: Production release control
on:
  pull_request_target:
    types: [closed]
  workflow_dispatch:
    inputs:
      operation:
        description: Recovery action (prefer a roll-forward hotfix)
        type: choice
        options: [hotfix, rollback]
        default: hotfix
        required: true
      fix_pr:
        description: Merged fix PR number for a hotfix
        type: string
        required: false
      issue:
        description: Issue number for a newly authored hotfix (instead of fix_pr)
        type: string
        required: false
      reason:
        description: Reason for recovery
        type: string
        required: true
      target_version:
        description: Rollback version; blank selects the previous verified deployed version
        type: string
        required: false
permissions:
  contents: write
  actions: write
  pull-requests: write
  issues: write
  checks: read
concurrency:
  group: mint-release-journal
  cancel-in-progress: false
jobs:
  request:
    if: vars.MINT_RELEASE_ENABLED == 'true' && github.actor == '%s'
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
          command: production-control
          github-token: ${{ github.token }}
`, cfg.HumanLogin, mintRef), nil
}
