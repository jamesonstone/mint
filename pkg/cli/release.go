package cli

import (
	"fmt"
	"os"

	"github.com/jamesonstone/mint/pkg/release"
	"github.com/spf13/cobra"
)

type releaseResolveFlags struct {
	commitish    string
	githubOutput string
}

type releaseWorkflowFlags struct {
	images  []string
	output  string
	mintRef string
}

var releaseResolveCommandFlags releaseResolveFlags
var releaseWorkflowCommandFlags releaseWorkflowFlags

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Resolve release versions and render publish workflows",
	Long: `Resolve release metadata and render publish workflows.

Mint release commands compute read-only SemVer release metadata from Git history
and generate GHCR or ECR publish workflow YAML. They do not create GitHub
Releases, deploy services, publish package-manager artifacts, or mutate Git
state directly.`,
	Args: cobra.NoArgs,
}

var releaseResolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Resolve the release tag for a Git commit",
	Long: `Resolve the release tag for a Git commit.

The resolver reads reachable Git history, selects a strict vX.Y.Z SemVer tag,
and prints the resolved version tag. It can also write all release fields in
GitHub Actions output format for workflow steps.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReleaseResolve(cmd, releaseResolveCommandFlags)
	},
}

var releaseWorkflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Render a GHCR or ECR publish workflow",
	Long: `Render a GHCR or ECR publish workflow.

Each --image flag must use name=<name>,uri=<image-uri>,dockerfile=<path> with
an optional context=<path>. Image URIs must be repository URIs without tags, and
all images in one workflow must use the same supported registry kind.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReleaseWorkflow(cmd, releaseWorkflowCommandFlags)
	},
}

func init() {
	bindReleaseResolveFlags(releaseResolveCmd, &releaseResolveCommandFlags)
	bindReleaseWorkflowFlags(releaseWorkflowCmd, &releaseWorkflowCommandFlags)

	releaseCmd.AddCommand(releaseResolveCmd)
	releaseCmd.AddCommand(releaseWorkflowCmd)
	rootCmd.AddCommand(releaseCmd)
}

func bindReleaseResolveFlags(cmd *cobra.Command, values *releaseResolveFlags) {
	flags := cmd.Flags()
	flags.StringVar(&values.commitish, "commitish", release.DefaultCommitish, "Git ref to resolve")
	flags.StringVar(&values.githubOutput, "github-output", "", "optional path to a GitHub Actions output file")
}

func bindReleaseWorkflowFlags(cmd *cobra.Command, values *releaseWorkflowFlags) {
	flags := cmd.Flags()
	flags.StringArrayVar(&values.images, "image", nil, "repeatable image spec: name=<name>,uri=<image-uri>,dockerfile=<path>,context=<path>")
	flags.StringVar(&values.output, "output", "", "optional workflow output path; stdout when omitted")
	flags.StringVar(&values.mintRef, "mint-ref", release.DefaultMintRef, "Mint action ref used by generated workflows")
}

func runReleaseResolve(cmd *cobra.Command, values releaseResolveFlags) error {
	result, err := release.Resolve(cmd.Context(), release.Options{
		Commitish: values.commitish,
	})
	if err != nil {
		return err
	}

	if values.githubOutput != "" {
		if err := release.WriteGitHubOutputFile(values.githubOutput, result); err != nil {
			return err
		}
	}

	_, err = fmt.Fprintln(cmd.OutOrStdout(), result.VersionTag)
	return err
}

func runReleaseWorkflow(cmd *cobra.Command, values releaseWorkflowFlags) error {
	images := make([]release.ImageSpec, 0, len(values.images))
	for _, value := range values.images {
		image, err := release.ParseImageSpec(value)
		if err != nil {
			return err
		}
		images = append(images, image)
	}

	workflow, err := release.GenerateWorkflow(release.WorkflowOptions{
		Images:  images,
		MintRef: values.mintRef,
	})
	if err != nil {
		return err
	}

	if values.output != "" {
		return os.WriteFile(values.output, []byte(workflow), 0o644)
	}

	_, err = fmt.Fprint(cmd.OutOrStdout(), workflow)
	return err
}
