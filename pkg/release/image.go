package release

import (
	"regexp"
	"strings"
)

var imageNamePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
var ecrHostPattern = regexp.MustCompile(`^[0-9]{12}\.dkr\.ecr\.([a-z0-9-]+)\.amazonaws\.com$`)

// ParseImageSpec parses a CLI image spec and derives its supported registry.
func ParseImageSpec(value string) (ImageSpec, error) {
	parts := strings.Split(value, ",")
	fields := make(map[string]string, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, raw, ok := strings.Cut(part, "=")
		if !ok {
			return ImageSpec{}, validationError("image", "expected key=value field %q", part)
		}
		key = strings.TrimSpace(key)
		raw = strings.TrimSpace(raw)
		if key == "" || raw == "" {
			return ImageSpec{}, validationError("image", "empty key or value in %q", part)
		}
		if _, exists := fields[key]; exists {
			return ImageSpec{}, validationError(key, "duplicate image field")
		}
		switch key {
		case "name", "uri", "dockerfile", "context":
			fields[key] = raw
		default:
			return ImageSpec{}, validationError(key, "unsupported image field")
		}
	}

	spec := ImageSpec{
		Name:       fields["name"],
		URI:        fields["uri"],
		Dockerfile: fields["dockerfile"],
		Context:    fields["context"],
	}
	if spec.Context == "" {
		spec.Context = "."
	}

	if err := validateImageSpecFields(spec); err != nil {
		return ImageSpec{}, err
	}

	host, kind, err := detectRegistry(spec.URI)
	if err != nil {
		return ImageSpec{}, err
	}
	spec.RegistryHost = host
	spec.RegistryKind = kind
	return spec, nil
}

func validateWorkflowImages(images []ImageSpec) ([]ImageSpec, RegistryKind, error) {
	if len(images) == 0 {
		return nil, "", validationError("image", "at least one image is required")
	}

	seenNames := make(map[string]struct{}, len(images))
	normalized := make([]ImageSpec, 0, len(images))
	var workflowRegistry RegistryKind

	for _, image := range images {
		if image.Context == "" {
			image.Context = "."
		}
		if err := validateImageSpecFields(image); err != nil {
			return nil, "", err
		}
		if _, exists := seenNames[image.Name]; exists {
			return nil, "", validationError("name", "duplicate image name %q", image.Name)
		}
		seenNames[image.Name] = struct{}{}

		host, kind, err := detectRegistry(image.URI)
		if err != nil {
			return nil, "", err
		}
		image.RegistryHost = host
		image.RegistryKind = kind
		if workflowRegistry == "" {
			workflowRegistry = kind
		}
		if workflowRegistry != kind {
			return nil, "", validationError("uri", "mixed registry kinds are not supported")
		}

		normalized = append(normalized, image)
	}

	return normalized, workflowRegistry, nil
}

func validateImageSpecFields(spec ImageSpec) error {
	switch {
	case spec.Name == "":
		return validationError("name", "image name is required")
	case spec.URI == "":
		return validationError("uri", "image uri is required")
	case spec.Dockerfile == "":
		return validationError("dockerfile", "image dockerfile is required")
	case strings.ContainsAny(spec.Name, "\r\n"):
		return validationError("name", "image name must be a single line")
	case !imageNamePattern.MatchString(spec.Name):
		return validationError("name", "image name must contain only letters, numbers, dot, underscore, or dash")
	case strings.ContainsAny(spec.URI, "\r\n"):
		return validationError("uri", "image uri must be a single line")
	case strings.ContainsAny(spec.Dockerfile, "\r\n"):
		return validationError("dockerfile", "image dockerfile must be a single line")
	case strings.ContainsAny(spec.Context, "\r\n"):
		return validationError("context", "image context must be a single line")
	}
	return nil
}

func detectRegistry(uri string) (string, RegistryKind, error) {
	if strings.Contains(uri, "@") {
		return "", "", validationError("uri", "image uri must not include a digest")
	}

	host, path, ok := strings.Cut(uri, "/")
	if !ok || host == "" || path == "" {
		return "", "", validationError("uri", "image uri must include registry host and repository path")
	}
	if strings.Contains(path, ":") {
		return "", "", validationError("uri", "image uri must not include a tag")
	}
	if strings.HasSuffix(path, "/") || strings.Contains(path, "//") {
		return "", "", validationError("uri", "image uri must include a valid repository path")
	}

	switch {
	case host == "ghcr.io":
		return host, RegistryGHCR, nil
	case ecrHostPattern.MatchString(host):
		return host, RegistryECR, nil
	default:
		return "", "", validationError("uri", "unsupported-registry: %s", host)
	}
}

func ecrRegion(host string) string {
	matches := ecrHostPattern.FindStringSubmatch(host)
	if matches == nil {
		return ""
	}
	return matches[1]
}
