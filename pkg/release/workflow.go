package release

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var mintRefPattern = regexp.MustCompile(`^[A-Za-z0-9._/-]+$`)

// GenerateWorkflow renders a deterministic Git-tag-first GHCR or ECR workflow.
func GenerateWorkflow(opts WorkflowOptions) (string, error) {
	mintRef := opts.MintRef
	if mintRef == "" {
		mintRef = DefaultMintRef
	}
	if !mintRefPattern.MatchString(mintRef) {
		return "", validationError("mint-ref", "mint ref contains unsupported characters")
	}

	images, registry, err := validateWorkflowImages(opts.Images)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString("name: Release Publish\n\n")
	builder.WriteString("on:\n")
	builder.WriteString("  push:\n\n")
	builder.WriteString("concurrency:\n")
	builder.WriteString("  group: release-publish\n")
	builder.WriteString("  cancel-in-progress: false\n\n")
	builder.WriteString("permissions:\n")
	builder.WriteString("  contents: write\n")
	if registry == RegistryGHCR {
		builder.WriteString("  packages: write\n")
	}
	if registry == RegistryECR {
		builder.WriteString("  id-token: write\n")
	}
	builder.WriteString("\n")
	builder.WriteString("jobs:\n")
	builder.WriteString("  publish:\n")
	builder.WriteString("    if: github.ref_name == github.event.repository.default_branch\n")
	builder.WriteString("    runs-on: ubuntu-latest\n")
	builder.WriteString("    steps:\n")
	builder.WriteString("      - name: Check out repository\n")
	builder.WriteString("        uses: actions/checkout@v4\n")
	builder.WriteString("        with:\n")
	builder.WriteString("          fetch-depth: 0\n")
	builder.WriteString("      - name: Refresh Git tags\n")
	builder.WriteString("        shell: bash\n")
	builder.WriteString("        run: git fetch --force --tags\n")
	builder.WriteString("      - name: Resolve release\n")
	builder.WriteString("        id: release\n")
	builder.WriteString("        uses: jamesonstone/mint@" + mintRef + "\n")
	builder.WriteString("        with:\n")
	builder.WriteString("          command: release-resolve\n")
	builder.WriteString("          commitish: ${{ github.sha }}\n")
	builder.WriteString("      - name: Write release notes\n")
	builder.WriteString("        id: release-notes\n")
	builder.WriteString("        env:\n")
	builder.WriteString("          RELEASE_NOTES: ${{ steps.release.outputs.release_notes }}\n")
	builder.WriteString("        shell: bash\n")
	builder.WriteString("        run: |\n")
	builder.WriteString(indentBlock(releaseNotesFileScript(), 10))
	builder.WriteString("      - name: Create release tag\n")
	builder.WriteString("        uses: jamesonstone/mint@" + mintRef + "\n")
	builder.WriteString("        with:\n")
	builder.WriteString("          command: release-tag\n")
	builder.WriteString("          release-tag: ${{ steps.release.outputs.version_tag }}\n")
	builder.WriteString("          target-sha: ${{ steps.release.outputs.target_sha }}\n")
	builder.WriteString("          release-notes-file: ${{ steps.release-notes.outputs.path }}\n")
	builder.WriteString("          release-remote: origin\n")
	builder.WriteString("          release-push: ${{ steps.release.outputs.needs_git_tag }}\n")
	builder.WriteString("      - name: Set up Docker Buildx\n")
	builder.WriteString("        uses: docker/setup-buildx-action@v3\n")

	switch registry {
	case RegistryGHCR:
		writeGHCRLogin(&builder)
	case RegistryECR:
		writeECRLogin(&builder, images)
	}

	for _, image := range images {
		writeBuildStep(&builder, image)
	}

	return builder.String(), nil
}

func releaseNotesFileScript() string {
	return `set -euo pipefail

notes_file="$RUNNER_TEMP/mint-release-notes.txt"
printf '%s\n' "$RELEASE_NOTES" > "$notes_file"
echo "path=$notes_file" >> "$GITHUB_OUTPUT"
`
}

func writeGHCRLogin(builder *strings.Builder) {
	builder.WriteString("      - name: Log in to GHCR\n")
	builder.WriteString("        uses: docker/login-action@v3\n")
	builder.WriteString("        with:\n")
	builder.WriteString("          registry: ghcr.io\n")
	builder.WriteString("          username: ${{ github.actor }}\n")
	builder.WriteString("          password: ${{ secrets.GITHUB_TOKEN }}\n")
}

func writeECRLogin(builder *strings.Builder, images []ImageSpec) {
	registries := uniqueRegistryHosts(images)
	builder.WriteString("      - name: Configure AWS credentials\n")
	builder.WriteString("        uses: aws-actions/configure-aws-credentials@v5\n")
	builder.WriteString("        with:\n")
	builder.WriteString("          role-to-assume: ${{ secrets.AWS_PUBLISH_ROLE_TO_ASSUME }}\n")
	builder.WriteString("          aws-region: " + ecrRegion(images[0].RegistryHost) + "\n")
	builder.WriteString("      - name: Log in to Amazon ECR\n")
	builder.WriteString("        uses: aws-actions/amazon-ecr-login@v2\n")
	if len(registries) > 0 {
		builder.WriteString("        with:\n")
		builder.WriteString("          registries: " + strings.Join(registries, ",") + "\n")
	}
}

func writeBuildStep(builder *strings.Builder, image ImageSpec) {
	builder.WriteString("      - name: Publish " + image.Name + "\n")
	builder.WriteString("        shell: bash\n")
	builder.WriteString("        run: |\n")
	script := fmt.Sprintf(`set -euo pipefail
docker buildx build \
  --file %s \
  --tag %s:"${{ steps.release.outputs.version_tag }}" \
  --tag %s:latest \
  --push \
  %s
`,
		shellQuote(image.Dockerfile),
		shellQuote(image.URI),
		shellQuote(image.URI),
		shellQuote(image.Context),
	)
	builder.WriteString(indentBlock(script, 10))
}

func uniqueRegistryHosts(images []ImageSpec) []string {
	seen := make(map[string]struct{}, len(images))
	for _, image := range images {
		seen[image.RegistryHost] = struct{}{}
	}
	hosts := make([]string, 0, len(seen))
	for host := range seen {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)
	return hosts
}

func indentBlock(value string, spaces int) string {
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(value, "\n")
	var builder strings.Builder
	for _, line := range lines {
		if line == "" {
			builder.WriteString(prefix)
			builder.WriteString("\n")
			continue
		}
		builder.WriteString(prefix)
		builder.WriteString(line)
		builder.WriteString("\n")
	}
	return builder.String()
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}
