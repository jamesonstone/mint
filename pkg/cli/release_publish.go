package cli

import (
	"fmt"

	"github.com/jamesonstone/mint/pkg/release"
	"github.com/spf13/cobra"
)

type releasePublishFlags struct {
	commitish    string
	owner        string
	repo         string
	title        string
	remote       string
	push         bool
	tokenEnv     string
	apiURL       string
	githubOutput string
	versionFile  string
}

var releasePublishCommandFlags releasePublishFlags

var releasePublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Resolve, tag, and publish a GitHub Release",
	Long: `Resolve, tag, and publish a GitHub Release.

The command owns release-state operations only: it resolves the SemVer version,
creates or reuses the Git tag, and creates or reuses the GitHub Release. It does
not build Docker images, publish to registries, or deploy services.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReleasePublish(cmd, releasePublishCommandFlags)
	},
}

func init() {
	bindReleasePublishFlags(releasePublishCmd, &releasePublishCommandFlags)
	releaseCmd.AddCommand(releasePublishCmd)
}

func bindReleasePublishFlags(cmd *cobra.Command, values *releasePublishFlags) {
	flags := cmd.Flags()
	flags.StringVar(&values.commitish, "commitish", release.DefaultCommitish, "Git ref to resolve")
	flags.StringVar(&values.owner, "owner", "", "GitHub repository owner")
	flags.StringVar(&values.repo, "repo", "", "GitHub repository name")
	flags.StringVar(&values.title, "title", "", "GitHub Release title; defaults to the resolved tag")
	flags.StringVar(&values.remote, "remote", release.DefaultTagRemote, "Git remote used when pushing a newly created tag")
	flags.BoolVar(&values.push, "push", true, "push refs/tags/<tag> to the configured remote when a tag is created")
	flags.StringVar(&values.tokenEnv, "token-env", "", "environment variable containing a GitHub token; falls back to MINT_GITHUB_TOKEN, GITHUB_TOKEN, then GH_TOKEN")
	flags.StringVar(&values.apiURL, "api-url", release.DefaultGitHubAPIBaseURL, "GitHub API base URL")
	flags.StringVar(&values.githubOutput, "github-output", "", "optional path to a GitHub Actions output file")
	flags.StringVar(&values.versionFile, "version-file", release.DefaultVersionFile, "path to write the resolved runtime version without the v prefix")
}

func runReleasePublish(cmd *cobra.Command, values releasePublishFlags) error {
	result, err := release.PublishRelease(cmd.Context(), release.PublishOptions{
		Commitish:   values.commitish,
		Owner:       values.owner,
		Repo:        values.repo,
		Title:       values.title,
		Token:       githubToken(values.tokenEnv),
		APIBaseURL:  values.apiURL,
		Remote:      values.remote,
		PushTag:     values.push,
		VersionFile: values.versionFile,
	})
	if err != nil {
		return err
	}

	if values.githubOutput != "" {
		if err := release.WritePublishOutputFile(values.githubOutput, result); err != nil {
			return err
		}
	}

	if result.GitHubRelease.Created {
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "published GitHub release %s %s\n", result.GitHubRelease.TagName, result.GitHubRelease.URL)
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "GitHub release %s already exists %s\n", result.GitHubRelease.TagName, result.GitHubRelease.URL)
	return err
}
