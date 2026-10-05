package promotion

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const RepositoryWriteAuthorization = "repository-write"

var ErrInsufficientRepositoryPermission = errors.New("human lacks repository write permission")

var loginPattern = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,38})$`)

func validateAuthorization(mode, human string, assignees []string) error {
	if mode != "" && mode != RepositoryWriteAuthorization {
		return fmt.Errorf("authorization must be repository-write")
	}
	if (mode == "") == (human == "") {
		return fmt.Errorf("set exactly one of authorization or human_login")
	}
	if human != "" && !loginPattern.MatchString(human) {
		return fmt.Errorf("invalid human_login")
	}
	seen := map[string]bool{}
	for _, login := range assignees {
		key := strings.ToLower(login)
		if !loginPattern.MatchString(login) || seen[key] {
			return fmt.Errorf("assignees must contain unique human logins")
		}
		seen[key] = true
	}
	return nil
}
func validateAuthorizationNodes(node *yaml.Node) error {
	for i := 0; i < len(node.Content); i += 2 {
		key, value := node.Content[i].Value, node.Content[i+1]
		if key == "authorization" || key == "human_login" {
			if value.Tag != "!!str" || value.Value == "" {
				return fmt.Errorf("%s must be a nonempty string", key)
			}
		}
		if key == "assignees" {
			if value.Kind != yaml.SequenceNode {
				return fmt.Errorf("assignees must be a list")
			}
			for _, item := range value.Content {
				if item.Tag != "!!str" {
					return fmt.Errorf("assignees must contain strings")
				}
			}
		}
	}
	return nil
}
func cloneAssignees(values []string) []string {
	if values == nil {
		return nil
	}
	return append([]string{}, values...)
}

// AuthorizeHuman reads effective repository permission, including team grants.
// Server failures and incomplete or mismatched identities never grant authority.
func (c Client) AuthorizeHuman(ctx context.Context, login string) error {
	if c.Authorization == "" {
		if c.HumanLogin != "" && login == c.HumanLogin {
			return nil
		}
		return fmt.Errorf("actor is not the configured human")
	}
	if c.Authorization != RepositoryWriteAuthorization || !loginPattern.MatchString(login) {
		return fmt.Errorf("actor is not an authorized human")
	}
	var permission struct {
		Permission string `json:"permission"`
		User       struct {
			Login string `json:"login"`
			Type  string `json:"type"`
		} `json:"user"`
	}
	status, err := c.request(ctx, "GET", c.repoPath("collaborators/"+url.PathEscape(login)+"/permission"), nil, &permission)
	if err != nil {
		return err
	}
	if status != 200 || permission.User.Login != login || permission.User.Type != "User" {
		return fmt.Errorf("actor lacks verified repository write permission")
	}
	if permission.Permission == "read" || permission.Permission == "triage" || permission.Permission == "none" {
		return ErrInsufficientRepositoryPermission
	}
	if permission.Permission != "write" && permission.Permission != "maintain" && permission.Permission != "admin" {
		return fmt.Errorf("unknown effective repository permission")
	}
	return nil
}
func (c Client) assignmentLogins() []string {
	if c.Assignees != nil {
		return cloneAssignees(c.Assignees)
	}
	if c.Authorization == "" && c.HumanLogin != "" {
		return []string{c.HumanLogin}
	}
	return nil
}
func (c Client) withAssignments(payload map[string]any) map[string]any {
	if len(c.assignmentLogins()) == 0 {
		delete(payload, "assignees")
	}
	return payload
}
func (c Client) assign(ctx context.Context, number int) error {
	names := c.assignmentLogins()
	if len(names) == 0 {
		return nil
	}
	status, err := c.request(ctx, "POST", c.repoPath(fmt.Sprintf("issues/%d/assignees", number)), map[string]any{"assignees": names}, nil)
	if err != nil {
		return err
	}
	if status != 201 {
		return fmt.Errorf("release assignment failed")
	}
	return nil
}

// MarshalYAML preserves explicit unassignment when generated proposals rewrite
// desired policy. Omitting nil retains legacy default assignment semantics.
func (p Policy) MarshalYAML() (any, error) {
	type plainPolicy Policy
	var node yaml.Node
	if err := node.Encode(plainPolicy(p)); err != nil {
		return nil, err
	}
	if p.Assignees != nil && len(p.Assignees) == 0 {
		node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "assignees"}, &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: []*yaml.Node{}})
	}
	return &node, nil
}
