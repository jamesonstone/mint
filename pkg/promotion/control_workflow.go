package promotion

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

// RenderControlWorkflow generates the repository-native request entry point.
// It uses trusted default-branch code and an immutable feature-bearing Mint pin.
func RenderControlWorkflow(cfg Config, mintRef string, configPaths ...string) (string, error) {
	if !shaPattern.MatchString(mintRef) {
		return "", fmt.Errorf("use a published feature-bearing Mint commit SHA")
	}
	if err := validateAuthorization(cfg.Authorization, cfg.HumanLogin, cfg.Assignees); err != nil {
		return "", fmt.Errorf("invalid control operator login")
	}
	if !ValidWorkflowPath(cfg.ControlWorkflowPath()) {
		return "", fmt.Errorf("control_workflow must be a YAML file directly under .github/workflows")
	}
	configPath := ".mint.yaml"
	if len(configPaths) > 1 {
		return "", fmt.Errorf("provide one repository-relative configuration path")
	}
	if len(configPaths) == 1 {
		configPath = configPaths[0]
	}
	if !SafeRepositoryPath(configPath) {
		return "", fmt.Errorf("configuration path must be a safe repository-relative file path")
	}
	actorGate := ""
	if cfg.Authorization == "" {
		actorGate = " && github.actor == '" + cfg.HumanLogin + "'"
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
    if: vars.MINT_RELEASE_ENABLED == 'true'%s
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
          production-config: '%s'
          github-token: ${{ github.token }}
`, actorGate, mintRef, configPath), nil
}

// SafeRepositoryPath excludes traversal, absolute paths and Actions expressions
// before a repository-owned filename is interpolated into generated YAML.
func SafeRepositoryPath(value string) bool {
	if !regexp.MustCompile(`^[A-Za-z0-9_.-]+(/[A-Za-z0-9_.-]+)*$`).MatchString(value) || path.Clean(value) != value {
		return false
	}
	for _, component := range strings.Split(value, "/") {
		if component == "." || component == ".." {
			return false
		}
	}
	return true
}

// ValidWorkflowPath matches the location GitHub Actions discovers for workflows.
func ValidWorkflowPath(value string) bool {
	return SafeRepositoryPath(value) && path.Dir(value) == ".github/workflows" && (path.Ext(value) == ".yml" || path.Ext(value) == ".yaml")
}
