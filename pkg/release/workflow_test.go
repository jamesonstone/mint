package release

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGenerateWorkflowGHCR(t *testing.T) {
	api, err := ParseImageSpec("name=api,uri=ghcr.io/jamesonstone/mint-api,dockerfile=Dockerfile.api,context=.")
	if err != nil {
		t.Fatalf("ParseImageSpec() error = %v", err)
	}
	worker, err := ParseImageSpec("name=worker,uri=ghcr.io/jamesonstone/mint-worker,dockerfile=Dockerfile.worker,context=worker")
	if err != nil {
		t.Fatalf("ParseImageSpec() error = %v", err)
	}

	workflow, err := GenerateWorkflow(WorkflowOptions{Images: []ImageSpec{api, worker}})
	if err != nil {
		t.Fatalf("GenerateWorkflow() error = %v", err)
	}

	assertValidYAML(t, workflow)
	for _, want := range []string{
		"concurrency:\n  group: release-publish\n  cancel-in-progress: false",
		"packages: write",
		"uses: jamesonstone/mint@v1",
		"command: release-resolve",
		"git fetch --force --tags",
		"RELEASE_NOTES: ${{ steps.release.outputs.release_notes }}",
		"printf '%s\\n' \"$RELEASE_NOTES\" > \"$notes_file\"",
		"git tag -a \"$tag\" \"$target\" -F \"$notes_file\"",
		"git push origin \"refs/tags/$tag\"",
		"docker/login-action@v3",
		"registry: ghcr.io",
		"username: ${{ github.actor }}",
		"password: ${{ secrets.GITHUB_TOKEN }}",
		"ghcr.io/jamesonstone/mint-api",
		"ghcr.io/jamesonstone/mint-worker",
		"--tag 'ghcr.io/jamesonstone/mint-api':\"${{ steps.release.outputs.version_tag }}\"",
		"--tag 'ghcr.io/jamesonstone/mint-api':latest",
		"--tag 'ghcr.io/jamesonstone/mint-worker':\"${{ steps.release.outputs.version_tag }}\"",
		"--tag 'ghcr.io/jamesonstone/mint-worker':latest",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("workflow missing %q:\n%s", want, workflow)
		}
	}
	assertTagBeforePublish(t, workflow)
	assertNoDeployContent(t, workflow)
	if strings.Contains(workflow, "cat > \"$notes_file\" <<") {
		t.Fatalf("workflow should not use a fixed heredoc delimiter for release notes:\n%s", workflow)
	}
}

func TestGenerateWorkflowECR(t *testing.T) {
	image, err := ParseImageSpec("name=api,uri=123456789012.dkr.ecr.us-east-1.amazonaws.com/mint-api,dockerfile=Dockerfile.api")
	if err != nil {
		t.Fatalf("ParseImageSpec() error = %v", err)
	}

	workflow, err := GenerateWorkflow(WorkflowOptions{Images: []ImageSpec{image}, MintRef: "v2"})
	if err != nil {
		t.Fatalf("GenerateWorkflow() error = %v", err)
	}

	assertValidYAML(t, workflow)
	for _, want := range []string{
		"id-token: write",
		"uses: jamesonstone/mint@v2",
		"aws-actions/configure-aws-credentials@v5",
		"role-to-assume: ${{ secrets.AWS_PUBLISH_ROLE_TO_ASSUME }}",
		"aws-region: us-east-1",
		"aws-actions/amazon-ecr-login@v2",
		"registries: 123456789012.dkr.ecr.us-east-1.amazonaws.com",
		"--tag '123456789012.dkr.ecr.us-east-1.amazonaws.com/mint-api':\"${{ steps.release.outputs.version_tag }}\"",
		"--tag '123456789012.dkr.ecr.us-east-1.amazonaws.com/mint-api':latest",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("workflow missing %q:\n%s", want, workflow)
		}
	}
	for _, reject := range []string{"registry: ghcr.io", "docker/login-action@v3"} {
		if strings.Contains(workflow, reject) {
			t.Fatalf("ECR workflow contains GHCR content %q:\n%s", reject, workflow)
		}
	}
	assertTagBeforePublish(t, workflow)
	assertNoDeployContent(t, workflow)
}

func TestGenerateWorkflowRejectsInvalidInputs(t *testing.T) {
	ghcr, err := ParseImageSpec("name=api,uri=ghcr.io/jamesonstone/mint-api,dockerfile=Dockerfile.api")
	if err != nil {
		t.Fatalf("ParseImageSpec() error = %v", err)
	}
	ecr, err := ParseImageSpec("name=worker,uri=123456789012.dkr.ecr.us-east-1.amazonaws.com/mint-worker,dockerfile=Dockerfile.worker")
	if err != nil {
		t.Fatalf("ParseImageSpec() error = %v", err)
	}

	tests := []struct {
		name string
		opts WorkflowOptions
		want string
	}{
		{name: "no images", opts: WorkflowOptions{}, want: "at least one image"},
		{name: "mixed registries", opts: WorkflowOptions{Images: []ImageSpec{ghcr, ecr}}, want: "mixed registry kinds"},
		{name: "bad mint ref", opts: WorkflowOptions{Images: []ImageSpec{ghcr}, MintRef: "bad ref"}, want: "mint ref"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GenerateWorkflow(tt.opts)
			if err == nil {
				t.Fatalf("GenerateWorkflow() error = nil, want %q", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("GenerateWorkflow() error = %q, want contains %q", err.Error(), tt.want)
			}
		})
	}
}

func assertValidYAML(t *testing.T, value string) {
	t.Helper()

	var decoded map[string]any
	if err := yaml.Unmarshal([]byte(value), &decoded); err != nil {
		t.Fatalf("workflow does not parse as YAML: %v\n%s", err, value)
	}
}

func assertTagBeforePublish(t *testing.T, workflow string) {
	t.Helper()

	tagIndex := strings.Index(workflow, "git tag -a")
	publishIndex := strings.Index(workflow, "docker buildx build")
	if tagIndex < 0 || publishIndex < 0 {
		t.Fatalf("workflow missing tag or publish step:\n%s", workflow)
	}
	if tagIndex > publishIndex {
		t.Fatalf("tag creation appears after publish:\n%s", workflow)
	}
}

func assertNoDeployContent(t *testing.T, workflow string) {
	t.Helper()

	rejected := []string{
		"aws ecs",
		"task-definition",
		"workflow_dispatch",
		"environment:",
		"ECS_SERVICE",
		"ECS_CLUSTER",
		"container-name",
		"github-release",
	}
	for _, value := range rejected {
		if strings.Contains(workflow, value) {
			t.Fatalf("workflow contains deploy content %q:\n%s", value, workflow)
		}
	}
}
