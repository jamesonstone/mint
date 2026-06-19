package release

import (
	"strings"
	"testing"
)

func TestParseImageSpecDefaultsContextAndDetectsGHCR(t *testing.T) {
	spec, err := ParseImageSpec("name=api,uri=ghcr.io/jamesonstone/mint-api,dockerfile=Dockerfile.api")
	if err != nil {
		t.Fatalf("ParseImageSpec() error = %v", err)
	}

	if spec.Name != "api" || spec.URI != "ghcr.io/jamesonstone/mint-api" || spec.Dockerfile != "Dockerfile.api" {
		t.Fatalf("ParseImageSpec() = %+v", spec)
	}
	if spec.Context != "." {
		t.Fatalf("Context = %q, want .", spec.Context)
	}
	if spec.RegistryHost != "ghcr.io" || spec.RegistryKind != RegistryGHCR {
		t.Fatalf("registry = %q/%q, want ghcr.io/%q", spec.RegistryHost, spec.RegistryKind, RegistryGHCR)
	}
}

func TestParseImageSpecDetectsECR(t *testing.T) {
	spec, err := ParseImageSpec("name=worker,uri=123456789012.dkr.ecr.us-east-1.amazonaws.com/mint-worker,dockerfile=Dockerfile.worker,context=.")
	if err != nil {
		t.Fatalf("ParseImageSpec() error = %v", err)
	}

	if spec.RegistryHost != "123456789012.dkr.ecr.us-east-1.amazonaws.com" {
		t.Fatalf("RegistryHost = %q", spec.RegistryHost)
	}
	if spec.RegistryKind != RegistryECR {
		t.Fatalf("RegistryKind = %q, want %q", spec.RegistryKind, RegistryECR)
	}
}

func TestParseImageSpecRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "missing name", input: "uri=ghcr.io/jamesonstone/mint,dockerfile=Dockerfile", want: "name"},
		{name: "missing uri", input: "name=api,dockerfile=Dockerfile", want: "uri"},
		{name: "missing dockerfile", input: "name=api,uri=ghcr.io/jamesonstone/mint", want: "dockerfile"},
		{name: "duplicate field", input: "name=api,name=worker,uri=ghcr.io/jamesonstone/mint,dockerfile=Dockerfile", want: "duplicate"},
		{name: "unknown field", input: "name=api,uri=ghcr.io/jamesonstone/mint,dockerfile=Dockerfile,tag=latest", want: "unsupported image field"},
		{name: "tagged uri", input: "name=api,uri=ghcr.io/jamesonstone/mint:latest,dockerfile=Dockerfile", want: "must not include a tag"},
		{name: "digest uri", input: "name=api,uri=ghcr.io/jamesonstone/mint@sha256:abcd,dockerfile=Dockerfile", want: "must not include a digest"},
		{name: "unsupported registry", input: "name=api,uri=docker.io/library/mint,dockerfile=Dockerfile", want: "unsupported-registry"},
		{name: "bad name", input: "name=bad/name,uri=ghcr.io/jamesonstone/mint,dockerfile=Dockerfile", want: "image name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseImageSpec(tt.input)
			if err == nil {
				t.Fatalf("ParseImageSpec() error = nil, want %q", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("ParseImageSpec() error = %q, want contains %q", err.Error(), tt.want)
			}
		})
	}
}

func TestValidateWorkflowImagesRejectsDuplicateAndMixedRegistries(t *testing.T) {
	ghcr, err := ParseImageSpec("name=api,uri=ghcr.io/jamesonstone/mint-api,dockerfile=Dockerfile.api")
	if err != nil {
		t.Fatalf("ParseImageSpec() error = %v", err)
	}
	duplicate, err := ParseImageSpec("name=api,uri=ghcr.io/jamesonstone/mint-worker,dockerfile=Dockerfile.worker")
	if err != nil {
		t.Fatalf("ParseImageSpec() error = %v", err)
	}
	ecr, err := ParseImageSpec("name=worker,uri=123456789012.dkr.ecr.us-east-1.amazonaws.com/mint-worker,dockerfile=Dockerfile.worker")
	if err != nil {
		t.Fatalf("ParseImageSpec() error = %v", err)
	}

	_, _, err = validateWorkflowImages([]ImageSpec{ghcr, duplicate})
	if err == nil || !strings.Contains(err.Error(), "duplicate image name") {
		t.Fatalf("duplicate validation error = %v", err)
	}

	_, _, err = validateWorkflowImages([]ImageSpec{ghcr, ecr})
	if err == nil || !strings.Contains(err.Error(), "mixed registry kinds") {
		t.Fatalf("mixed registry validation error = %v", err)
	}
}
