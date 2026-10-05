package promotion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// AuthorizeAmbientRun permits trusted repository controller code to request
// read-only observations and policy-authorized automatic reconciliation.
func (c Client) AuthorizeAmbientRun(ctx context.Context, cfg Config, id int64, event string) error {
	_, err := c.AmbientRun(ctx, cfg, id, event)
	return err
}
func (c Client) AmbientRun(ctx context.Context, cfg Config, id int64, event string) (WorkflowRun, error) {
	var run WorkflowRun
	if event != "push" && event != "workflow_run" {
		return run, fmt.Errorf("ambient reconciliation requires a repository merge or producer event")
	}
	status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("actions/runs/%d", id)), nil, &run)
	if err != nil {
		return run, err
	}
	if status != 200 || run.ID != id || !WorkflowPathMatches(run.Path, cfg.ControlWorkflowPath()) || run.Event != event || run.Status != "in_progress" || run.HeadBranch != cfg.DefaultBranch || run.Repository.FullName != cfg.Repository || run.HeadRepository.FullName != cfg.Repository {
		return run, fmt.Errorf("ambient run is not an active trusted default-branch controller")
	}
	return run, nil
}

// DispatchObservation invokes the configured read-only project adapter. The
// eventual authenticated observation artifact remains the only live evidence.
func (c Client) DispatchObservation(ctx context.Context, cfg Config) error {
	if cfg.Mode != "deployment" || cfg.Scope != "shared" || !ValidWorkflowPath(cfg.ObservationWorkflow) {
		return fmt.Errorf("trusted observation adapter is not configured")
	}
	return c.dispatchEnvironment(ctx, cfg, cfg.ObservationWorkflow, map[string]string{"environment": cfg.Environment})
}

// DispatchIntent sends only a frozen journal intent, never a mutable image tag.
// Adapters upload mint-request before start, then mint-deployment after runtime
// verification. A dispatch response is not deployment or runtime evidence.
func (c Client) DispatchIntent(ctx context.Context, cfg Config, i Intent) error {
	if cfg.Schema != 2 || i.Environment != cfg.Environment || i.Status != "pending" || !ValidWorkflowPath(cfg.PromotionWorkflow) {
		return fmt.Errorf("dispatch requires a pending exact environment intent")
	}
	data, err := json.Marshal(i)
	if err != nil {
		return err
	}
	return c.dispatchEnvironment(ctx, cfg, cfg.PromotionWorkflow, map[string]string{"environment": cfg.Environment, "intent_id": i.ID, "intent": string(data)})
}
func (c Client) dispatchEnvironment(ctx context.Context, cfg Config, workflow string, inputs map[string]string) error {
	status, err := c.request(ctx, "POST", c.repoPath("actions/workflows/"+url.PathEscape(workflow)+"/dispatches"), map[string]any{"ref": cfg.DefaultBranch, "inputs": inputs}, nil)
	if err != nil {
		return err
	}
	if status != 204 {
		return fmt.Errorf("environment adapter dispatch was not accepted")
	}
	return nil
}
