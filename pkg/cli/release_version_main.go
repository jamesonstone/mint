package cli

import (
	"encoding/json"
	"fmt"
	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/jamesonstone/mint/pkg/release"
	"github.com/spf13/cobra"
	"os"
	"slices"
)

func init() {
	var mainRef, remote, output, target, config string
	var push bool
	var control []string
	command := &cobra.Command{Use: "version-main", Short: "Version every untagged first-parent main commit without publishing Releases", Args: cobra.NoArgs}
	command.Flags().StringVar(&mainRef, "main-ref", "origin/main", "authoritative current default branch ref (not triggering event SHA)")
	command.Flags().StringVar(&remote, "remote", "origin", "remote receiving immutable tags")
	command.Flags().BoolVar(&push, "push", true, "push immutable tags; callers must serialize allocations")
	command.Flags().StringSliceVar(&control, "control-paths", []string{"CHANGELOG.md", ".mint/proposal.json", ".mint/summary.md", ".mint/status.json"}, "exact trusted release-control paths")
	command.Flags().StringVar(&target, "target-commitish", "", "specific main source whose identity is exported after batch versioning")
	command.Flags().StringVar(&output, "github-output", "", "append final version identity to Actions output")
	command.Flags().StringVar(&config, "config", "", "optional schema 2 policy; a component policy versions only its paths with its tag prefix and uses its release-control paths")
	command.RunE = func(cmd *cobra.Command, args []string) error {
		var scope release.Scope
		if config != "" {
			if cmd.Flags().Changed("control-paths") {
				return fmt.Errorf("--config supplies release-control paths; do not also pass --control-paths")
			}
			policy, err := promotion.LoadPolicy(config)
			if err != nil {
				return err
			}
			if policy.Schema != 2 || !promotion.SafeRepositoryPath(config) {
				return fmt.Errorf("--config must be a schema 2 policy at a safe repository-relative path")
			}
			component := policy.ComponentScope()
			scope = component.ReleaseScope()
			control = append(append([]string(nil), policy.ControlPaths...), config)
			// The optional status file stays control-only, as in the default list.
			if !slices.Contains(control, component.StatusPath()) {
				control = append(control, component.StatusPath())
			}
		}
		versions, err := release.VersionMain(cmd.Context(), release.MainVersionOptions{MainRef: mainRef, Remote: remote, Push: push, ControlPaths: control, Scope: scope})
		if err != nil {
			return err
		}
		if output != "" && scope.TagPrefix != "" {
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
