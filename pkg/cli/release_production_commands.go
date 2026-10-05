package cli

import "github.com/spf13/cobra"

// Adapter operations remain callable, but everyday help presents the small
// operator surface. Each operation accepts only options it actually consumes.
func newProductionCommand() *cobra.Command {
	group := &cobra.Command{Use: "production", Short: "Review production releases; request recovery through repository Actions",
		Long: `Successful builds update one production release PR. Review and merge it to promote its exact artifact through your deployment workflow. Use repository Actions for hotfix or rollback; the commands below provide the same recovery requests.

Integration adapters remain available by name; see the production adapter contract.`}
	operations := []struct {
		name     string
		options  []string
		operator bool
	}{
		{"status", nil, true},
		{"hotfix", []string{"fix-pr", "issue", "reason"}, true},
		{"rollback", []string{"to", "reason"}, true},
		{"workflow", []string{"mint-ref"}, true},
		{"candidate", []string{"run-id"}, false},
		{"propose", []string{"kind", "event", "summary", "pin"}, false},
		{"validate", []string{"pr"}, false},
		{"intent", []string{"pr", "merge-sha"}, false},
		{"start", []string{"intent-id", "run-id"}, false},
		{"finish", []string{"intent-id", "run-id"}, false},
		{"published", []string{"intent-id"}, false},
		{"bootstrap", []string{"run-id"}, false},
		{"prepare-hotfix", []string{"input"}, false},
		{"publish", []string{"intent-id"}, false},
		{"version-hotfix", []string{"pr"}, false},
		{"propose-rollback", []string{"pin", "summary"}, false},
		{"status-pr", []string{"intent-id"}, false},
		{"validate-review", []string{"pr"}, false},
		{"scan", nil, false},
		{"control", []string{"input", "event", "run-id"}, false},
		{"report", []string{"intent-id"}, false},
	}
	for _, operation := range operations {
		f := productionFlags{Kind: "normal", Event: "candidate"}
		name := operation.name
		cmd := &cobra.Command{Use: name, Short: productionHelp(name), Args: cobra.NoArgs, Hidden: !operation.operator,
			RunE: func(cmd *cobra.Command, args []string) error { return runProduction(cmd, name, f) }}
		flags := cmd.Flags()
		flags.StringVar(&f.Config, "config", ".mint.yaml", "repository-owned release policy")
		outputHelp := "write resulting JSON manifest"
		if name == "workflow" {
			outputHelp = "write YAML to the configured control workflow path"
		}
		flags.StringVar(&f.Output, "output", "", outputHelp)
		if name != "workflow" {
			flags.StringVar(&f.APIURL, "api-url", "https://api.github.com", "GitHub API URL")
			flags.StringVar(&f.TokenEnv, "token-env", "GH_TOKEN", "job-scoped GitHub credential environment variable")
		}
		for _, option := range operation.options {
			switch option {
			case "input":
				flags.StringVar(&f.Input, option, "", "typed request JSON file")
			case "kind":
				flags.StringVar(&f.Kind, option, "normal", "normal or hotfix proposal")
			case "event":
				flags.StringVar(&f.Event, option, "candidate", "candidate/close/reopen, or authenticated Actions event for control")
			case "intent-id":
				flags.StringVar(&f.IntentID, option, "", "frozen intent identity")
			case "merge-sha":
				flags.StringVar(&f.MergeSHA, option, "", "exact reviewed proposal merge SHA")
			case "summary":
				flags.StringVar(&f.Summary, option, "", "manual summary override")
			case "pin":
				flags.StringVar(&f.Pin, option, "", "hotfix candidate SHA or rollback deployed source SHA; normal releases always advance")
			case "run-id":
				flags.Int64Var(&f.RunID, option, 0, "server-verified workflow run")
			case "to":
				flags.StringVar(&f.Version, option, "", "previously deployed version; default previous verified deployment")
			case "reason":
				flags.StringVar(&f.Reason, option, "", "reason for the production recovery request")
			case "mint-ref":
				flags.StringVar(&f.MintRef, option, "", "published immutable Mint commit for generated Actions workflow")
			case "issue":
				flags.IntVar(&f.Issue, option, 0, "issue for a newly authored production fix")
			case "fix-pr":
				flags.IntVar(&f.FixPR, option, 0, "merged reviewed fix PR to isolate from queued default-branch changes")
			case "pr":
				flags.IntVar(&f.PR, option, 0, "exact trusted proposal/source PR number")
			}
		}
		if name == "status-pr" {
			cmd.Deprecated = "use report; outcomes are attached to the original release PR"
		}
		group.AddCommand(cmd)
	}
	return group
}
