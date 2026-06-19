package cli

import (
	"fmt"

	"github.com/jamesonstone/mint/pkg/changelog"
	"github.com/spf13/cobra"
)

type changelogFlags struct {
	prevTag    string
	currentTag string
	owner      string
	repo       string
	output     string
}

var rootChangelogFlags changelogFlags
var changelogCommandFlags changelogFlags

var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "Generate CHANGELOG.md from conventional commits",
	Long: `Generate a CHANGELOG.md release entry from conventional commits between
Git refs.

Mint validates the tag range, parses conventional commit subjects, links GitHub
issues and commits, groups entries deterministically, and prepends a new release
block to the changelog without overwriting an existing version.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChangelog(cmd, changelogCommandFlags)
	},
}

func init() {
	bindChangelogFlags(rootCmd, &rootChangelogFlags)
	rootCmd.RunE = runRoot

	bindChangelogFlags(changelogCmd, &changelogCommandFlags)
	rootCmd.AddCommand(changelogCmd)
}

func bindChangelogFlags(cmd *cobra.Command, values *changelogFlags) {
	flags := cmd.Flags()
	flags.StringVar(&values.prevTag, "prev-tag", "", "previous Git tag or ref; empty for first release")
	flags.StringVar(&values.currentTag, "current-tag", "", "current Git tag or ref to release")
	flags.StringVar(&values.owner, "owner", "", "GitHub repository owner")
	flags.StringVar(&values.repo, "repo", "", "GitHub repository name")
	flags.StringVar(&values.output, "output", changelog.DefaultOutputFile, "path to CHANGELOG.md")
}

func runRoot(cmd *cobra.Command, args []string) error {
	if rootChangelogRequested(cmd) {
		return runChangelog(cmd, rootChangelogFlags)
	}
	return cmd.Help()
}

func rootChangelogRequested(cmd *cobra.Command) bool {
	for _, name := range []string{"prev-tag", "current-tag", "owner", "repo", "output"} {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

func runChangelog(cmd *cobra.Command, values changelogFlags) error {
	result, err := changelog.Generate(cmd.Context(), changelog.Options{
		PrevTag:       values.prevTag,
		CurrentTag:    values.currentTag,
		RepoOwner:     values.owner,
		RepoName:      values.repo,
		OutputFile:    values.output,
		WarningWriter: cmd.ErrOrStderr(),
	})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "added %s with %d commits, %d breaking\n",
		result.Tag, result.CommitCount, result.BreakingCount)
	return err
}
