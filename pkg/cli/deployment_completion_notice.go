package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/spf13/cobra"
)

// noticeWait bounds how long a completion notice waits for its adapter run.
// The last job of the run sends the notice, so the run ends soon after.
type noticeWait struct {
	Attempts int
	Interval time.Duration
}

var defaultNoticeWait = noticeWait{Attempts: 60, Interval: 10 * time.Second}

// runAdapterCompletionNotice handles the repository_dispatch event that an
// adapter run sends from its last job. GitHub creates no workflow_run event
// for a run that the workflow token dispatched, but it does create a run for a
// repository_dispatch that the workflow token sends. The run ID in the payload
// is data only: the completion callback authenticates the run and its manifest.
func runAdapterCompletionNotice(cmd *cobra.Command, policy promotion.Policy, f productionFlags, data []byte) error {
	return runAdapterCompletionNoticeWith(cmd, policy, f, data, runProduction, defaultNoticeWait)
}

func runAdapterCompletionNoticeWith(cmd *cobra.Command, policy promotion.Policy, f productionFlags, data []byte, execute func(*cobra.Command, string, productionFlags) error, wait noticeWait) error {
	id, err := completionNoticeRunID(data)
	if err != nil {
		return err
	}
	if len(policy.Environments) == 0 {
		return fmt.Errorf("configured environments required")
	}
	client := promotion.Client{APIURL: f.APIURL, Token: os.Getenv(f.TokenEnv), Repository: policy.Repository, HumanLogin: policy.HumanLogin, Authorization: policy.Authorization, Assignees: policy.Assignees, Component: policy.ComponentScope()}
	configs := make([]promotion.Config, 0, len(policy.Environments))
	for name := range policy.Environments {
		cfg, err := policy.ForEnvironment(name)
		if err != nil {
			return err
		}
		if err := client.AuthorizeAmbientRun(cmd.Context(), cfg, f.RunID, f.Event); err != nil {
			return err
		}
		configs = append(configs, cfg)
	}
	results := map[string]string{}
	for attempt := 1; ; attempt++ {
		run, err := client.WorkflowRunIdentity(cmd.Context(), id)
		if err != nil {
			return err
		}
		if !adapterRunPath(configs, run.Path) {
			results["notice"] = "the notice does not name an adapter run; nothing changed"
			return writeDeploymentJSON(cmd, results, f.Output)
		}
		if run.Status == "completed" {
			break
		}
		if attempt >= wait.Attempts {
			results["notice"] = "the adapter run is still active; the next producer callback settles it"
			return writeDeploymentJSON(cmd, results, f.Output)
		}
		select {
		case <-cmd.Context().Done():
			return cmd.Context().Err()
		case <-time.After(wait.Interval):
		}
	}
	// The same path as a workflow_run completion: run authentication, finish
	// from the verified manifest, then queued reconciliation.
	_, err = runAdapterCallbackWith(cmd, policy, f, []byte(fmt.Sprintf(`{"workflow_run":{"id":%d}}`, id)), execute)
	return err
}

func adapterRunPath(configs []promotion.Config, path string) bool {
	for _, cfg := range configs {
		if cfg.PromotionWorkflow != "" && promotion.WorkflowPathMatches(path, cfg.PromotionWorkflow) {
			return true
		}
		if cfg.ObservationWorkflow != "" && promotion.WorkflowPathMatches(path, cfg.ObservationWorkflow) {
			return true
		}
	}
	return false
}

// completionNoticeRunID reads the adapter run ID from the notice payload. It
// accepts a JSON number or a decimal string, because `gh api -f` sends strings.
func completionNoticeRunID(data []byte) (int64, error) {
	var event struct {
		Action        string
		ClientPayload struct {
			RunID json.RawMessage `json:"run_id"`
		} `json:"client_payload"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return 0, err
	}
	if event.Action != promotion.AdapterCompletedEvent {
		return 0, fmt.Errorf("repository event %q is not a Mint adapter completion notice", event.Action)
	}
	raw := bytes.TrimSpace(event.ClientPayload.RunID)
	if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		raw = raw[1 : len(raw)-1]
	}
	id, err := strconv.ParseInt(string(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("adapter completion notice lacks a workflow run")
	}
	return id, nil
}
