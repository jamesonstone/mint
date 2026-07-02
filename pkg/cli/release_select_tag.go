package cli

import (
	"fmt"

	"github.com/jamesonstone/mint/pkg/release"
	"github.com/spf13/cobra"
)

type releaseSelectTagFlags struct {
	commitish    string
	requestedTag string
	githubOutput string
}

var releaseSelectTagCommandFlags releaseSelectTagFlags

var releaseSelectTagCmd = &cobra.Command{
	Use:   "select-tag",
	Short: "Select an existing SemVer release tag",
	Long: `Select an existing SemVer release tag.

The command validates a caller-provided strict vX.Y.Z tag when one is supplied.
Otherwise, it selects the highest strict SemVer Git tag pointing at the target
commit. It never computes a new release version, creates tags, pushes images,
or deploys services.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReleaseSelectTag(cmd, releaseSelectTagCommandFlags)
	},
}

func init() {
	bindReleaseSelectTagFlags(releaseSelectTagCmd, &releaseSelectTagCommandFlags)
	releaseCmd.AddCommand(releaseSelectTagCmd)
}

func bindReleaseSelectTagFlags(cmd *cobra.Command, values *releaseSelectTagFlags) {
	flags := cmd.Flags()
	flags.StringVar(&values.commitish, "commitish", release.DefaultCommitish, "Git ref to inspect")
	flags.StringVar(&values.requestedTag, "requested-tag", "", "optional strict vX.Y.Z SemVer tag supplied by the caller")
	flags.StringVar(&values.githubOutput, "github-output", "", "optional path to a GitHub Actions output file")
}

func runReleaseSelectTag(cmd *cobra.Command, values releaseSelectTagFlags) error {
	result, err := release.SelectTag(cmd.Context(), release.SelectTagOptions{
		Commitish:    values.commitish,
		RequestedTag: values.requestedTag,
	})
	if err != nil {
		return err
	}

	if values.githubOutput != "" {
		if err := release.WriteSelectTagOutputFile(values.githubOutput, result); err != nil {
			return err
		}
	}

	_, err = fmt.Fprintln(cmd.OutOrStdout(), result.VersionTag)
	return err
}
