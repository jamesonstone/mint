package promotion

import (
	"bytes"
	"fmt"
	"io"
	"regexp"

	"gopkg.in/yaml.v3"
)

// ParsePolicy rejects YAML indirection and ambiguous fields before decoding intent.
func ParsePolicy(data []byte) (Policy, error) {
	document, err := policyDocument(data)
	if err != nil {
		return Policy{}, err
	}
	schema := 0
	for i := 0; i < len(document.Content); i += 2 {
		if document.Content[i].Value == "schema_version" {
			if err := document.Content[i+1].Decode(&schema); err != nil {
				return Policy{}, err
			}
		}
	}
	if schema == 1 {
		cfg, err := parseLegacyConfig(data)
		if err != nil {
			return Policy{}, err
		}
		return legacyPolicy(cfg), nil
	}
	if schema != 2 {
		return Policy{}, fmt.Errorf("unsupported release policy schema")
	}
	var p Policy
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&p); err != nil {
		return Policy{}, err
	}
	if err := p.validate(); err != nil {
		return Policy{}, err
	}
	return p, nil
}
func policyDocument(data []byte) (*yaml.Node, error) {
	var document yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&document); err != nil {
		return nil, err
	}
	if decoder.Decode(new(yaml.Node)) != io.EOF {
		return nil, fmt.Errorf("release policy must contain exactly one YAML document")
	}
	if len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("release policy must be a YAML mapping")
	}
	if err := validatePolicyNode(document.Content[0]); err != nil {
		return nil, err
	}
	if err := validateAuthorizationNodes(document.Content[0]); err != nil {
		return nil, err
	}
	return document.Content[0], nil
}
func validatePolicyNode(node *yaml.Node) error {
	if node.Kind == yaml.AliasNode || node.Anchor != "" {
		return fmt.Errorf("release policy cannot use YAML aliases or anchors")
	}
	switch node.Tag {
	case "!!map", "!!seq", "!!str", "!!int", "!!bool", "!!null":
	default:
		return fmt.Errorf("unsupported release policy YAML tag")
	}
	if node.Kind == yaml.MappingNode {
		seen := map[string]bool{}
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || key.Value == "<<" {
				return fmt.Errorf("release policy must use explicit string keys")
			}
			if seen[key.Value] {
				return fmt.Errorf("duplicate release policy field %s", key.Value)
			}
			seen[key.Value] = true
		}
	}
	for _, child := range node.Content {
		if err := validatePolicyNode(child); err != nil {
			return err
		}
	}
	return nil
}

func (p *Policy) validate() error {
	if err := validateAuthorization(p.Authorization, p.HumanLogin, p.Assignees); err != nil {
		return err
	}
	if p.Mode != "deployment" && p.Mode != "artifact" && p.Mode != "package" {
		return fmt.Errorf("policy mode must be deployment, artifact or package")
	}
	if !regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`).MatchString(p.Repository) || !SafeRepositoryPath(p.DefaultBranch) || p.BuildWorkflow == "" || len(p.Environments) == 0 {
		return fmt.Errorf("incomplete environment release policy")
	}
	for _, workflow := range []string{p.BuildWorkflow, p.ValidationWorkflow, p.ControlWorkflow, p.BaselineWorkflow, p.BaselineBuildWorkflow} {
		if workflow != "" && !ValidWorkflowPath(workflow) {
			return fmt.Errorf("policy workflow must be directly under .github/workflows")
		}
	}
	if p.Artifact.Repository != "" && !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:/-]*[A-Za-z0-9]$`).MatchString(p.Artifact.Repository) {
		return fmt.Errorf("invalid artifact repository")
	}
	if p.ControlPaths == nil {
		p.ControlPaths = []string{"CHANGELOG.md", ".mint/proposal.json", ".mint/summary.md"}
	}
	if err := validateControlPaths(p.ControlPaths); err != nil {
		return err
	}
	if err := validateCheckNames(p.RequiredChecks); err != nil {
		return err
	}
	tags := map[string]string{}
	publisher := ""
	for name, env := range p.Environments {
		if !ValidEnvironmentName(name) {
			return fmt.Errorf("unsafe environment name %s", name)
		}
		if err := validateEnvironment(env); err != nil {
			return fmt.Errorf("environment %s: %w", name, err)
		}
		if env.Operation == "" {
			env.Operation = "release"
			p.Environments[name] = env
		}
		if p.Mode == "deployment" && env.Publish {
			if env.Scope != "shared" || publisher != "" {
				return fmt.Errorf("deployment policy permits one shared canonical release publisher")
			}
			publisher = name
		}
		if env.ImageTag != "" {
			if prior, ok := tags[env.ImageTag]; ok {
				return fmt.Errorf("image_tag collision between %s and %s", prior, name)
			}
			tags[env.ImageTag] = name
		}
	}
	if p.DefaultEnvironment != "" {
		if _, ok := p.Environments[p.DefaultEnvironment]; !ok {
			return fmt.Errorf("default_environment must name a configured environment")
		}
	}
	return p.validateDependencies()
}
