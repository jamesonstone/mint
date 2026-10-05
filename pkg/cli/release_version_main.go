package cli

import (
	"encoding/json"
	"fmt"
	"github.com/jamesonstone/mint/pkg/release"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	var mainRef, remote, output, target string
	var push bool
	var control []string
	command := &cobra.Command{Use: "version-main", Short: "Version every untagged first-parent main commit without publishing Releases", Args: cobra.NoArgs}
	command.Flags().StringVar(&mainRef, "main-ref", "origin/main", "authoritative current default branch ref (not triggering event SHA)")
	command.Flags().StringVar(&remote, "remote", "origin", "remote receiving immutable tags")
	command.Flags().BoolVar(&push, "push", true, "push immutable tags; callers must serialize allocations")
	command.Flags().StringSliceVar(&control, "control-paths", []string{"CHANGELOG.md", ".mint/proposal.json", ".mint/summary.md", ".mint/status.json"}, "exact trusted release-control paths")
	command.Flags().StringVar(&target, "target-commitish", "", "specific main source whose identity is exported after batch versioning")
	command.Flags().StringVar(&output, "github-output", "", "append final version identity to Actions output")
	command.RunE = func(cmd *cobra.Command, args []string) error {
		versions, err := release.VersionMain(cmd.Context(), release.MainVersionOptions{MainRef: mainRef, Remote: remote, Push: push, ControlPaths: control})
		if err != nil {
			return err
		}
		if output != "" && len(versions) > 0 {
			last := versions[len(versions)-1]
			if target != "" {
				last, err = release.VersionIdentity(cmd.Context(), "", target, control)
				if err != nil {
					return err
				}
			}
			file, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				return err
			}
			_, writeErr := fmt.Fprintf(file, "version_tag=%s\ntarget_sha=%s\ncontrol_only=%t\n", last.Version, last.SHA, last.ControlOnly)
			closeErr := file.Close()
			if writeErr != nil {
				return writeErr
			}
			if closeErr != nil {
				return closeErr
			}
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(versions)
	}
	releaseCmd.AddCommand(command)
}
