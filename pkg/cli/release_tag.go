package cli

import (
	"fmt"

	"github.com/jamesonstone/mint/pkg/release"
	"github.com/spf13/cobra"
)

type releaseTagFlags struct {
	tag          string
	target       string
	notesFile    string
	remote       string
	push         bool
	githubOutput string
}

var releaseTagCommandFlags releaseTagFlags

var releaseTagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Create or reuse an annotated SemVer Git tag",
	Long: `Create or reuse an annotated SemVer Git tag.

The command validates the target commit, creates an annotated release tag from
the provided release notes file, and never moves an existing tag. Existing tags
on the same target commit are treated as success.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReleaseTag(cmd, releaseTagCommandFlags)
	},
}

func init() {
	bindReleaseTagFlags(releaseTagCmd, &releaseTagCommandFlags)
	releaseCmd.AddCommand(releaseTagCmd)
}

func bindReleaseTagFlags(cmd *cobra.Command, values *releaseTagFlags) {
	flags := cmd.Flags()
	flags.StringVar(&values.tag, "tag", "", "strict vX.Y.Z SemVer tag to create")
	flags.StringVar(&values.target, "target", "", "target commitish for the tag")
	flags.StringVar(&values.notesFile, "notes-file", "", "file containing annotated tag notes")
	flags.StringVar(&values.remote, "remote", release.DefaultTagRemote, "Git remote used when pushing the tag")
	flags.BoolVar(&values.push, "push", true, "push refs/tags/<tag> to the configured remote")
	flags.StringVar(&values.githubOutput, "github-output", "", "optional path to a GitHub Actions output file")
}

func runReleaseTag(cmd *cobra.Command, values releaseTagFlags) error {
	result, err := release.CreateReleaseTag(cmd.Context(), release.TagOptions{
		Tag:       values.tag,
		Target:    values.target,
		NotesFile: values.notesFile,
		Remote:    values.remote,
		Push:      values.push,
	})
	if err != nil {
		return err
	}

	if values.githubOutput != "" {
		if err := release.WriteReleaseTagOutputFile(values.githubOutput, result); err != nil {
			return err
		}
	}

	if result.Created {
		if result.Pushed {
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "created and pushed Git tag %s %s\n", result.TagName, result.TargetSHA)
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "created Git tag %s %s\n", result.TagName, result.TargetSHA)
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Git tag %s already exists on target commit %s\n", result.TagName, result.TargetSHA)
	return err
}
