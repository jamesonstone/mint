package cli

import (
	"fmt"
	"os"
	"strings"

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

type releaseGitHubFlags struct {
	owner        string
	repo         string
	tag          string
	target       string
	title        string
	notesFile    string
	tokenEnv     string
	apiURL       string
	githubOutput string
}

var releaseResolveCommandFlags releaseResolveFlags
var releaseWorkflowCommandFlags releaseWorkflowFlags
var releaseGitHubCommandFlags releaseGitHubFlags

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Resolve versions, publish GitHub releases, and render workflows",
	Long: `Resolve release metadata, publish GitHub Releases, and render publish workflows.

Mint release commands compute read-only SemVer release metadata from Git
history, publish GitHub Releases for resolved tags, and generate GHCR or ECR
publish workflow YAML. They do not deploy services, publish package-manager
artifacts, or make the resolver mutate Git state directly.`,
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

var releaseGitHubCmd = &cobra.Command{
	Use:   "github",
	Short: "Create or reuse a GitHub Release for a SemVer tag",
	Long: `Create or reuse a GitHub Release for a SemVer tag.

The command is idempotent for an existing release with the same tag. The token
is read from the environment so scripts do not need to pass secrets as command
arguments.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReleaseGitHub(cmd, releaseGitHubCommandFlags)
	},
}

func init() {
	bindReleaseResolveFlags(releaseResolveCmd, &releaseResolveCommandFlags)
	bindReleaseWorkflowFlags(releaseWorkflowCmd, &releaseWorkflowCommandFlags)
	bindReleaseGitHubFlags(releaseGitHubCmd, &releaseGitHubCommandFlags)

	releaseCmd.AddCommand(releaseResolveCmd)
	releaseCmd.AddCommand(releaseWorkflowCmd)
	releaseCmd.AddCommand(releaseGitHubCmd)
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

func bindReleaseGitHubFlags(cmd *cobra.Command, values *releaseGitHubFlags) {
	flags := cmd.Flags()
	flags.StringVar(&values.owner, "owner", "", "GitHub repository owner")
	flags.StringVar(&values.repo, "repo", "", "GitHub repository name")
	flags.StringVar(&values.tag, "tag", "", "strict vX.Y.Z SemVer tag to release")
	flags.StringVar(&values.target, "target", "", "commitish where GitHub should create the tag when missing")
	flags.StringVar(&values.title, "title", "", "release title; defaults to --tag")
	flags.StringVar(&values.notesFile, "notes-file", "", "optional file containing release notes")
	flags.StringVar(&values.tokenEnv, "token-env", "", "environment variable containing a GitHub token; falls back to MINT_GITHUB_TOKEN, GITHUB_TOKEN, then GH_TOKEN")
	flags.StringVar(&values.apiURL, "api-url", release.DefaultGitHubAPIBaseURL, "GitHub API base URL")
	flags.StringVar(&values.githubOutput, "github-output", "", "optional path to a GitHub Actions output file")
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

func runReleaseGitHub(cmd *cobra.Command, values releaseGitHubFlags) error {
	notes := ""
	if values.notesFile != "" {
		data, err := os.ReadFile(values.notesFile)
		if err != nil {
			return err
		}
		notes = string(data)
	}

	result, err := release.PublishGitHubRelease(cmd.Context(), release.GitHubReleaseOptions{
		Owner:      values.owner,
		Repo:       values.repo,
		Tag:        values.tag,
		Target:     values.target,
		Title:      values.title,
		Notes:      notes,
		Token:      githubToken(values.tokenEnv),
		APIBaseURL: values.apiURL,
	})
	if err != nil {
		return err
	}

	if values.githubOutput != "" {
		if err := release.WriteGitHubReleaseOutputFile(values.githubOutput, result); err != nil {
			return err
		}
	}

	if result.Created {
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "created GitHub release %s %s\n", result.TagName, result.URL)
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "GitHub release %s already exists %s\n", result.TagName, result.URL)
	return err
}

func githubToken(tokenEnv string) string {
	names := make([]string, 0, 4)
	if strings.TrimSpace(tokenEnv) != "" {
		names = append(names, strings.TrimSpace(tokenEnv))
	}
	names = append(names, "MINT_GITHUB_TOKEN", "GITHUB_TOKEN", "GH_TOKEN")

	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	return ""
}
