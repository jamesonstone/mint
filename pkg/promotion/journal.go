package promotion

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

// ErrFileNotFound distinguishes an absent optional document from a transport failure.
var ErrFileNotFound = errors.New("repository file not found")

const journalBranch = "mint-release-state"
const journalPath = "release-state.json"

// JournalSnapshot is a durable state read pinned to a Git revision.
type JournalSnapshot struct {
	State    State
	Revision string
}
type gitObject struct {
	SHA    string `json:"sha"`
	Object struct {
		SHA string `json:"sha"`
	} `json:"object"`
}

// LoadJournal reads metadata independently of Actions artifact retention.
func (c Client) LoadJournal(ctx context.Context, environment string) (JournalSnapshot, error) {
	var ref gitObject
	status, err := c.request(ctx, "GET", c.repoPath("git/ref/heads/"+journalBranch), nil, &ref)
	if err != nil {
		return JournalSnapshot{}, err
	}
	if status == 404 {
		return JournalSnapshot{State: NewState(c.Repository, environment)}, nil
	}
	if status != 200 || !shaPattern.MatchString(ref.Object.SHA) {
		return JournalSnapshot{}, fmt.Errorf("invalid journal ref")
	}
	var file struct{ Content, Encoding string }
	status, err = c.request(ctx, "GET", c.repoPath("contents/"+journalPath+"?ref="+ref.Object.SHA), nil, &file)
	if err != nil {
		return JournalSnapshot{}, err
	}
	if status != 200 || file.Encoding != "base64" {
		return JournalSnapshot{}, fmt.Errorf("journal content unavailable")
	}
	data, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return JournalSnapshot{}, err
	}
	state, err := DecodeState(data, c.Repository, environment)
	if err != nil {
		return JournalSnapshot{}, err
	}
	return JournalSnapshot{State: state, Revision: ref.Object.SHA}, nil
}

// SaveJournal commits metadata with a non-forcing CAS ref update. A concurrent
// writer fails instead of overwriting evidence; replay reloads and reconciles.
func (c Client) SaveJournal(ctx context.Context, snapshot JournalSnapshot, message string) (string, error) {
	data, err := json.MarshalIndent(snapshot.State, "", "  ")
	if err != nil {
		return "", err
	}
	commit, err := c.createCommit(ctx, "", map[string]string{journalPath: string(data)}, snapshot.Revision, "mint: "+message)
	if err != nil {
		return "", err
	}
	if snapshot.Revision == "" {
		status, err := c.request(ctx, "POST", c.repoPath("git/refs"), map[string]any{"ref": "refs/heads/" + journalBranch, "sha": commit}, nil)
		if err != nil {
			return "", err
		}
		if status != 201 {
			return "", fmt.Errorf("journal initialization failed")
		}
	} else {
		status, err := c.request(ctx, "PATCH", c.repoPath("git/refs/heads/"+journalBranch), map[string]any{"sha": commit, "force": false}, nil)
		if err != nil {
			return "", err
		}
		if status != 200 {
			return "", fmt.Errorf("journal CAS failed")
		}
	}
	return commit, nil
}
func (c Client) createCommit(ctx context.Context, baseTree string, files map[string]string, parent, message string, extraParents ...string) (string, error) {
	entries := []map[string]any{}
	for path, content := range files {
		entries = append(entries, map[string]any{"path": path, "mode": "100644", "type": "blob", "content": content})
	}
	body := map[string]any{"tree": entries}
	if baseTree != "" {
		body["base_tree"] = baseTree
	}
	var tree gitObject
	status, err := c.request(ctx, "POST", c.repoPath("git/trees"), body, &tree)
	if err != nil {
		return "", err
	}
	if status != 201 || !shaPattern.MatchString(tree.SHA) {
		return "", fmt.Errorf("git tree creation failed")
	}
	parents := []string{}
	if parent != "" {
		parents = append(parents, parent)
	}
	parents = append(parents, extraParents...)
	identity := map[string]string{"name": c.HumanName, "email": c.HumanEmail}
	body = map[string]any{"tree": tree.SHA, "parents": parents, "message": message, "author": identity, "committer": identity}
	var commit gitObject
	status, err = c.request(ctx, "POST", c.repoPath("git/commits"), body, &commit)
	if err != nil {
		return "", err
	}
	if status != 201 || !shaPattern.MatchString(commit.SHA) {
		return "", fmt.Errorf("git commit creation failed")
	}
	return commit.SHA, nil
}

// ReadFile is an exact-commit read used for merged release declarations.
func (c Client) ReadFile(ctx context.Context, path, sha string) ([]byte, error) {
	if !shaPattern.MatchString(sha) {
		return nil, fmt.Errorf("file reads require an exact commit")
	}
	var file struct{ Content, Encoding string }
	status, err := c.request(ctx, "GET", c.repoPath("contents/"+url.PathEscape(path)+"?ref="+sha), nil, &file)
	if err != nil {
		return nil, err
	}
	if status == 404 {
		return nil, ErrFileNotFound
	}
	if status != 200 || file.Encoding != "base64" {
		return nil, fmt.Errorf("exact release declaration unavailable")
	}
	return base64.StdEncoding.DecodeString(file.Content)
}
