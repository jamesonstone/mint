package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/spf13/cobra"
)

// Repository controller routing treats events as data. Authority is fetched
// independently from GitHub before any request or observation dispatch.
func runPolicyControl(cmd *cobra.Command, policy promotion.Policy, f productionFlags) error {
	data, err := os.ReadFile(f.Input)
	if err != nil {
		return err
	}
	var event struct {
		Action      string
		Inputs      map[string]string
		PullRequest struct{ Number int } `json:"pull_request"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}
	client := promotion.Client{APIURL: f.APIURL, Token: os.Getenv(f.TokenEnv), Repository: policy.Repository, HumanLogin: policy.HumanLogin, Authorization: policy.Authorization, Assignees: policy.Assignees}
	if err := client.VerifyRepository(cmd.Context()); err != nil {
		return err
	}
	if f.Event == "workflow_run" {
		if handled, err := runAdapterCallback(cmd, policy, f, data); handled || err != nil {
			return err
		}
		return runBuildCallback(cmd, policy, f, data)
	}
	if f.Event == "push" {
		names := make([]string, 0, len(policy.Environments))
		for name := range policy.Environments {
			names = append(names, name)
		}
		sort.Strings(names)
		results := map[string]string{}
		failures := 0
		for _, name := range names {
			cfg, err := policy.ForEnvironment(name)
			if err != nil {
				return err
			}
			if err := client.AuthorizeAmbientRun(cmd.Context(), cfg, f.RunID, f.Event); err != nil {
				return err
			}
			if cfg.Mode != "deployment" || cfg.Scope == "local" || cfg.ObservationWorkflow == "" {
				results[name] = "observer not configured"
				continue
			}
			if err := client.DispatchObservation(cmd.Context(), cfg); err != nil {
				results[name] = err.Error()
				failures++
			} else {
				results[name] = "observation requested; live outcome remains unobserved"
			}
		}
		if err := writeDeploymentJSON(cmd, results, f.Output); err != nil {
			return err
		}
		if failures > 0 {
			return fmt.Errorf("%d observation requests failed", failures)
		}
		return nil
	}
	if f.Event == "workflow_dispatch" {
		f.Environment = event.Inputs["environment"]
		if f.Environment == "" {
			return fmt.Errorf("select the deployment environment")
		}
	}
	if f.Event == "pull_request_target" && event.Action == "closed" && event.PullRequest.Number > 0 {
		pr, err := client.Pull(cmd.Context(), event.PullRequest.Number)
		if err != nil {
			return err
		}
		if pr.Merged {
			beforeSHA, err := client.PolicyParent(cmd.Context(), pr.MergeSHA)
			if err != nil {
				return err
			}
			oldData, oldErr := client.ReadFile(cmd.Context(), f.Config, beforeSHA)
			newData, newErr := client.ReadFile(cmd.Context(), f.Config, pr.MergeSHA)
			if oldErr == nil && newErr == nil && string(oldData) != string(newData) {
				before, err := promotion.ParsePolicy(oldData)
				if err != nil {
					return err
				}
				after, err := promotion.ParsePolicy(newData)
				if err != nil {
					return err
				}
				name, err := promotion.PolicyRequestEnvironment(before, after)
				if err == nil {
					cfg, err := before.ForEnvironment(name)
					if err != nil {
						return err
					}
					if err := client.AuthorizeControlRun(cmd.Context(), cfg, f.RunID, f.Event); err != nil {
						return err
					}
					f.Environment, f.PR, f.MergeSHA = name, pr.Number, pr.MergeSHA
					return runProduction(cmd, "policy-request", f)
				}
				for name, env := range before.Environments {
					next, ok := after.Environments[name]
					if ok && (env.Target != next.Target || env.Follow != next.Follow || env.Operation != next.Operation || env.Reason != next.Reason) {
						return err
					}
				}
			}
		}
		if pr.Merged && strings.Contains(pr.Body, "<!-- mint:proposal:") {
			declarationData, err := client.ReadFile(cmd.Context(), ".mint/proposal.json", pr.MergeSHA)
			if err != nil {
				return err
			}
			var declaration promotion.Declaration
			if err := json.Unmarshal(declarationData, &declaration); err != nil {
				return err
			}
			cfg, err := policy.ForEnvironment(declaration.Environment)
			if err != nil {
				return err
			}
			if err := client.AuthorizeControlRun(cmd.Context(), cfg, f.RunID, f.Event); err != nil {
				return err
			}
			f.Environment, f.PR, f.MergeSHA = cfg.Environment, pr.Number, pr.MergeSHA
			return runProduction(cmd, "intent", f)
		}

	}
	cfg, err := policy.ForEnvironment(f.Environment)
	if err != nil {
		return err
	}
	if err := client.AuthorizeControlRun(cmd.Context(), cfg, f.RunID, f.Event); err != nil {
		return err
	}
	// Route the authenticated event through the shared request implementation.
	return runProductionWithControl(cmd, policy, cfg, f)
}

func runProductionWithControl(cmd *cobra.Command, policy promotion.Policy, cfg promotion.Config, f productionFlags) error {
	f.Environment = cfg.Environment
	f.Event = "authorized-" + f.Event
	return runProduction(cmd, "control", f)
}

func (o *productionOperation) requestPromotion(ctx context.Context, version, reason string) (any, bool, error) {
	o.snapshot.State.Target = version
	if version == "" {
		p := o.snapshot.State.Proposals["normal"]
		p.Selection = "latest"
		if p.ID != "" {
			o.snapshot.State.Proposals["normal"] = p
		}
	}
	o.flags.Kind, o.flags.Summary = "normal", reason
	if o.config.Deploy == "automatic" {
		o.config.Deploy = "reviewed"
	}
	return o.propose(ctx)
}
