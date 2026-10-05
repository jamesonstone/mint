package promotion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the authenticated GitHub transport. Credentials are environment-only.
type Client struct {
	APIURL, Token, Repository, HumanLogin string
	HTTP                                  *http.Client
}

func (c Client) request(ctx context.Context, method, path string, body, out any) (int, error) {
	base, err := url.Parse(c.APIURL)
	if err != nil || base.Host == "" || (base.Scheme != "https" && base.Hostname() != "127.0.0.1" && base.Hostname() != "localhost") {
		return 0, fmt.Errorf("GitHub API must use HTTPS (loopback tests excepted)")
	}
	if c.Token == "" {
		return 0, fmt.Errorf("GitHub release credential is required")
	}
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.APIURL, "/")+"/"+path, reader)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	hc := c.HTTP
	if hc == nil {
		hc = &http.Client{Timeout: 45 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	res, err := hc.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		if out != nil {
			if err := json.NewDecoder(io.LimitReader(res.Body, 32<<20)).Decode(out); err != nil {
				return res.StatusCode, err
			}
		}
		return res.StatusCode, nil
	}
	if res.StatusCode == 404 {
		return res.StatusCode, nil
	}
	return res.StatusCode, fmt.Errorf("GitHub %s %s returned %d; no mutation retry was attempted", method, path, res.StatusCode)
}
func (c Client) repoPath(path string) string { return "repos/" + c.Repository + "/" + path }

// VerifyRepository accepts repository installation tokens without GET /user.
// GitHub authorizes each write with the job-scoped permissions; this read fences
// the credential to the configured repository, not a personal account.
func (c Client) VerifyRepository(ctx context.Context) error {
	var repo struct {
		FullName string `json:"full_name"`
	}
	status, err := c.request(ctx, "GET", "repos/"+c.Repository, nil, &repo)
	if err != nil {
		return err
	}
	if status != 200 || repo.FullName != c.Repository {
		return fmt.Errorf("release credential cannot access configured repository")
	}
	return nil
}

const AutomationLogin = "github-actions[bot]"
const AutomationEmail = "41898282+github-actions[bot]@users.noreply.github.com"

// IsReleaseAuthor allows only the designated human and built-in Actions identity
// for generated release controls. It never authorizes application source changes.
func (c Client) IsReleaseAuthor(login string) bool {
	return login == AutomationLogin || (c.HumanLogin != "" && login == c.HumanLogin)
}

// WorkflowRun contains server-attested run identity and conclusion.
type WorkflowRun struct {
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Actor     struct {
		Login string `json:"login"`
	} `json:"actor"`
	ID         int64  `json:"id"`
	HeadSHA    string `json:"head_sha"`
	HeadBranch string `json:"head_branch"`
	Event      string `json:"event"`
	Path       string `json:"path"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HTMLURL    string `json:"html_url"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	HeadRepository struct {
		FullName string `json:"full_name"`
	} `json:"head_repository"`
}

// TrustedRun authenticates provenance from GitHub rather than a workflow payload.
func (c Client) ObservedRun(ctx context.Context, id int64, path string) (WorkflowRun, error) {
	var run WorkflowRun
	status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("actions/runs/%d", id)), nil, &run)
	if err != nil {
		return run, err
	}
	if status != 200 || run.ID != id || run.Repository.FullName != c.Repository || run.HeadRepository.FullName != c.Repository || !WorkflowPathMatches(run.Path, path) || (run.Event != "push" && run.Event != "workflow_dispatch") {
		return run, fmt.Errorf("workflow run is not a successful trusted build")
	}
	return run, nil
}

// WorkflowRunIdentity authenticates callback identity without trusting event payloads.
func (c Client) WorkflowRunIdentity(ctx context.Context, id int64) (WorkflowRun, error) {
	var run WorkflowRun
	status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("actions/runs/%d", id)), nil, &run)
	if err != nil {
		return run, err
	}
	if status != 200 || run.ID != id || id <= 0 || run.Repository.FullName != c.Repository || run.HeadRepository.FullName != c.Repository {
		return run, fmt.Errorf("workflow callback does not belong to configured repository")
	}
	return run, nil
}

// DispatchChecks explicitly launches validation for token-authored proposal heads.
func (c Client) DispatchChecks(ctx context.Context, workflow, branch string, pr ...int) error {
	payload := map[string]any{"ref": branch}
	if len(pr) > 0 && pr[0] > 0 {
		payload["inputs"] = map[string]string{"mint_pr": fmt.Sprint(pr[0])}
	}
	status, err := c.request(ctx, "POST", c.repoPath("actions/workflows/"+url.PathEscape(workflow)+"/dispatches"), payload, nil)
	if err == nil && status != 204 {
		err = fmt.Errorf("validation dispatch unavailable")
	}
	return err
}

// TrustedRun additionally requires a completed successful producer run.
func (c Client) TrustedRun(ctx context.Context, id int64, path string) (WorkflowRun, error) {
	run, err := c.ObservedRun(ctx, id, path)
	if err != nil {
		return run, err
	}
	if run.Status != "completed" || run.Conclusion != "success" {
		return run, fmt.Errorf("workflow has no successful completion")
	}
	return run, nil
}
