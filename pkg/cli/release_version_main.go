package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jamesonstone/mint/v2/pkg/release"
	"github.com/spf13/cobra"
)

func init() {
	releaseCmd.AddCommand(newVersionMainCommand())
}

func newVersionMainCommand() *cobra.Command {
	var mainRef, remote, output, target string
	var push bool
	var control []string
	var scoped scopeFlags
	command := &cobra.Command{Use: "version-main", Short: "Version every untagged first-parent main commit without publishing Releases", Args: cobra.NoArgs}
	command.Flags().StringVar(&mainRef, "main-ref", "origin/main", "authoritative current default branch ref (not triggering event SHA)")
	command.Flags().StringVar(&remote, "remote", "origin", "remote receiving immutable tags")
	command.Flags().BoolVar(&push, "push", true, "push immutable tags; callers must serialize allocations")
	command.Flags().StringSliceVar(&control, "control-paths", []string{"CHANGELOG.md"}, "exact paths for release-only changes")
	command.Flags().StringVar(&target, "target-commitish", "", "specific main source whose identity is exported after batch versioning")
	command.Flags().StringVar(&output, "github-output", "", "append final version identity to Actions output")
	bindScopeFlags(command.Flags(), &scoped, true)
	command.RunE = func(cmd *cobra.Command, args []string) error {
		scope := scoped.scope()
		versions, err := release.VersionMain(cmd.Context(), release.MainVersionOptions{MainRef: mainRef, Remote: remote, Push: push, ControlPaths: control, Scope: scope})
		if err != nil {
			return err
		}
		if output != "" && (scope.TagPrefix != "" || len(scope.Paths) > 0) {
			// A component reports the identity of the requested commit (or main)
			// and whether that commit changed the component; it always writes
			// outputs so callers can skip unchanged components.
			commitish := target
			if commitish == "" {
				commitish = mainRef
			}
			identity, changed, err := release.ComponentVersionIdentity(cmd.Context(), "", commitish, control, scope)
			if err != nil {
				return err
			}
			if err := appendOutput(output, fmt.Sprintf("version_tag=%s\ntarget_sha=%s\ncontrol_only=%t\ncomponent_changed=%t\n", identity.Version, identity.SHA, identity.ControlOnly, changed)); err != nil {
				return err
			}
		} else if output != "" && len(versions) > 0 {
			last := versions[len(versions)-1]
			if target != "" {
				last, err = release.VersionIdentity(cmd.Context(), "", target, control, scope)
				if err != nil {
					return err
				}
			}
			if err := appendOutput(output, fmt.Sprintf("version_tag=%s\ntarget_sha=%s\ncontrol_only=%t\n", last.Version, last.SHA, last.ControlOnly)); err != nil {
				return err
			}
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(versions)
	}
	return command
}

func appendOutput(path, text string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, writeErr := file.WriteString(text)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}
