package cli

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/spf13/cobra"
)

func TestReconciledQueueCannotAdvanceLaterOrPausedEnvironment(t *testing.T) {
	for _, scenario := range []string{"reviewed", "automatic", "manual", "paused", "later-baseline", "in-flight", "unpublished", "missing-intent", "failed", "transport-failure", "wrong-environment"} {
		t.Run(scenario, func(t *testing.T) {
			cfg := promotion.Config{Schema: 2, Mode: "deployment", Scope: "shared", Repository: "owner/repo", Environment: "stage", Deploy: "reviewed"}
			if scenario == "automatic" {
				cfg.Deploy = "automatic"
			}
			if scenario == "manual" {
				cfg.Deploy = "manual"
			}
			intent := promotion.Intent{ID: "recovered", Environment: cfg.Environment, Status: "deployed"}
			state := promotion.NewState(cfg.Repository, cfg.Environment)
			state.Schema = 2
			state.Intents[intent.ID] = intent
			state.Baseline = &promotion.Baseline{ID: intent.ID}
			if scenario == "paused" {
				state.Paused = true
			}
			if scenario == "later-baseline" {
				state.Baseline.ID = "newer-release"
			}
			if scenario == "in-flight" {
				state.InFlight = "next-deployment"
			}
			if scenario == "missing-intent" {
				state.Intents = map[string]promotion.Intent{}
			}
			if scenario == "unpublished" {
				intent.Status = "publication_pending"
			}
			if scenario == "failed" {
				recorded := intent
				recorded.Status = "deployment_failed"
				state.Intents[intent.ID] = recorded
			}
			if scenario == "wrong-environment" {
				intent.Environment = "other"
			}
			reads := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reads++
				if r.Method != "GET" {
					t.Error("unexpected mutation", r.Method)
					w.WriteHeader(500)
					return
				}
				switch {
				case strings.Contains(r.URL.Path, "/git/ref/"):
					_ = json.NewEncoder(w).Encode(map[string]any{"object": map[string]string{"sha": strings.Repeat("a", 40)}})
				case strings.Contains(r.URL.Path, "/git/commits/"):
					_ = json.NewEncoder(w).Encode(map[string]any{"tree": map[string]string{"sha": strings.Repeat("b", 40)}})
				case strings.Contains(r.URL.Path, "/git/trees/"):
					_ = json.NewEncoder(w).Encode(map[string]any{"tree": []any{map[string]string{"path": "release-state.json", "type": "blob", "sha": strings.Repeat("c", 40)}}})
				case strings.Contains(r.URL.Path, "/git/blobs/"):
					data, _ := json.Marshal(state)
					_ = json.NewEncoder(w).Encode(map[string]string{"encoding": "base64", "content": base64.StdEncoding.EncodeToString(data)})
				default:
					t.Error("unexpected read", r.URL.Path)
					w.WriteHeader(500)
				}
			}))
			defer server.Close()
			t.Setenv("MINT_QUEUE_TEST_TOKEN", "test")
			var out bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetContext(t.Context())
			cmd.SetOut(&out)
			f := productionFlags{APIURL: server.URL, TokenEnv: "MINT_QUEUE_TEST_TOKEN", RunID: 99, Event: "workflow_dispatch", Output: "operator-result.json", Kind: "hotfix", Pin: "prior-fix"}
			calls := 0
			execute := func(cmd *cobra.Command, operation string, child productionFlags) error {
				calls++
				if operation != "recovery-propose" || child.Environment != cfg.Environment || child.RunID != 99 || child.Event != "workflow_dispatch" || child.IntentID != intent.ID || child.Output != "" || child.Kind != "normal" || child.Pin != "" {
					t.Fatal("queue lost exact human recovery authority", operation, child)
				}
				_, _ = fmt.Fprint(cmd.OutOrStdout(), "suppressed-child")
				if scenario == "transport-failure" {
					return fmt.Errorf("PR unavailable")
				}
				return nil
			}
			err := advanceReconciledQueueWith(cmd, cfg, f, intent, execute)
			advance := scenario == "reviewed" || scenario == "automatic" || scenario == "transport-failure"
			if (calls == 1) != advance || calls > 1 {
				t.Fatal("stale/manual/paused request advanced queue", calls)
			}
			if (err != nil) != (scenario == "transport-failure" || scenario == "wrong-environment") {
				t.Fatal(err)
			}
			if (scenario == "manual" || scenario == "unpublished" || scenario == "wrong-environment") && reads != 0 {
				t.Fatal("inactive queue performed API reads", reads)
			}
			_, _ = fmt.Fprint(cmd.OutOrStdout(), "restored")
			if out.String() != "restored" {
				t.Fatal("output not restored", out.String())
			}
		})
	}
}
