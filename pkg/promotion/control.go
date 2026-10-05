package promotion

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var hotfixTitle = regexp.MustCompile(`^hotfix(?:\([^)\n]+\))?!?:\s+\S`)

// ControlRequest is a repository Actions request, not deployment authorization.
type ControlRequest struct {
	Operation        string `json:"operation"`
	IntentID         string `json:"intent_id"`
	ObservationRunID int64  `json:"observation_run_id"`
	FixPR            int    `json:"fix_pr"`
	Issue            int    `json:"issue"`
	Version          string `json:"target_version"`
	Reason           string `json:"reason"`
}

// ControlWorkflowPath allows existing adapters to select their trusted controller.
func (cfg Config) ControlWorkflowPath() string {
	if cfg.ControlWorkflow != "" {
		return cfg.ControlWorkflow
	}
	return ".github/workflows/mint-recovery.yaml"
}

// AuthorizeControlRun binds requests to a live default-branch workflow and the
// configured human. Event text and CLI actor flags never establish authority.
func (c Client) AuthorizeControlRun(ctx context.Context, cfg Config, id int64, event string) error {
	if id <= 0 || (event != "workflow_dispatch" && event != "pull_request_target") {
		return fmt.Errorf("production requests require a trusted Actions dispatch or merged PR event")
	}
	var run WorkflowRun
	status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("actions/runs/%d", id)), nil, &run)
	if err != nil {
		return err
	}
	if status != 200 || run.ID != id || !WorkflowPathMatches(run.Path, cfg.ControlWorkflowPath()) || run.Event != event || run.Status != "in_progress" || run.HeadBranch != cfg.DefaultBranch || run.Repository.FullName != cfg.Repository || run.HeadRepository.FullName != cfg.Repository || run.Actor.Login != cfg.HumanLogin {
		return fmt.Errorf("request is not an authorized active default-branch control run")
	}
	return nil
}

// ControlEvent reads only typed request fields, then fetches authoritative PR
// identity. A title requests preparation but cannot bypass review or merge.
func (c Client) ControlEvent(ctx context.Context, cfg Config, event string, data []byte) (ControlRequest, error) {
	var e struct {
		Action     string
		Repository struct {
			FullName string `json:"full_name"`
		}
		Inputs      map[string]string
		PullRequest struct{ Number int } `json:"pull_request"`
	}
	if err := json.Unmarshal(data, &e); err != nil {
		return ControlRequest{}, err
	}
	if e.Repository.FullName != cfg.Repository {
		return ControlRequest{}, fmt.Errorf("foreign control event")
	}
	if event == "workflow_dispatch" {
		if cfg.Schema == 2 && e.Inputs["environment"] != cfg.Environment {
			return ControlRequest{}, fmt.Errorf("request environment does not match selected policy")
		}
		r := ControlRequest{Operation: e.Inputs["operation"], Reason: e.Inputs["reason"], Version: e.Inputs["target_version"], IntentID: e.Inputs["intent_id"]}
		var err error
		if e.Inputs["observation_run_id"] != "" {
			r.ObservationRunID, err = strconv.ParseInt(e.Inputs["observation_run_id"], 10, 64)
			if err != nil || r.ObservationRunID <= 0 {
				return r, fmt.Errorf("observation run must be a positive integer")
			}
		}
		if r.FixPR, err = optionalNumber(e.Inputs["fix_pr"]); err != nil {
			return r, err
		}
		if r.Issue, err = optionalNumber(e.Inputs["issue"]); err != nil {
			return r, err
		}
		if r.Operation != "hotfix" && r.Operation != "rollback" && (cfg.Schema != 2 || (r.Operation != "promote" && r.Operation != "resume" && r.Operation != "observe" && r.Operation != "reconcile")) {
			return r, fmt.Errorf("select hotfix or rollback; roll-forward hotfix is recommended when practical")
		}
		return r, nil
	}
	if event != "pull_request_target" || e.Action != "closed" || e.PullRequest.Number <= 0 {
		return ControlRequest{}, nil
	}
	pr, err := c.Pull(ctx, e.PullRequest.Number)
	if err != nil {
		return ControlRequest{}, err
	}
	if !pr.Merged || pr.Base.Ref != cfg.DefaultBranch || strings.Contains(pr.Body, "<!-- mint:") || !hotfixTitle.MatchString(pr.Title) {
		return ControlRequest{}, nil
	}
	return ControlRequest{Operation: "hotfix", FixPR: pr.Number, Reason: pr.Title}, nil
}

func optionalNumber(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("fix PR and issue must be positive repository numbers")
	}
	return n, nil
}
