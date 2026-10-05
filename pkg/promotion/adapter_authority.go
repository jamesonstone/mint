package promotion

import (
	"context"
	"fmt"
)

// AuthorizeAdapterRun binds legacy integration routing to server-owned workflow
// identity. The event payload and caller-selected role never establish authority.
func (c Client) AuthorizeAdapterRun(ctx context.Context, cfg Config, id int64, role, event string) error {
	workflow := cfg.ControlWorkflowPath()
	allowed := false
	switch role {
	case "source":
		workflow, allowed = cfg.BuildWorkflow, event == "push" || event == "workflow_dispatch"
	case "promote":
		workflow, allowed = cfg.PromotionWorkflow, event == "push"
	case "control":
		allowed = event == "workflow_run" || event == "workflow_dispatch" || event == "pull_request_target"
	}
	if cfg.Schema != 1 || id <= 0 || !allowed {
		return fmt.Errorf("compatibility adapter requires a supported schema 1 Actions event")
	}
	var run WorkflowRun
	status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("actions/runs/%d", id)), nil, &run)
	if err != nil {
		return err
	}
	if status != 200 || run.ID != id || !WorkflowPathMatches(run.Path, workflow) || run.Event != event || run.Status != "in_progress" || run.HeadBranch != cfg.DefaultBranch || run.Repository.FullName != cfg.Repository || run.HeadRepository.FullName != cfg.Repository {
		return fmt.Errorf("adapter is not an active trusted default-branch workflow")
	}
	if event == "workflow_dispatch" || event == "pull_request_target" {
		c.Authorization, c.HumanLogin = cfg.Authorization, cfg.HumanLogin
		return c.AuthorizeHuman(ctx, run.Actor.Login)
	}
	return nil
}
