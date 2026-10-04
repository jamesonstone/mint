package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/jamesonstone/mint/pkg/release"
	"github.com/spf13/cobra"
	"io"
	"os"
	"reflect"
	"strings"
)

type productionFlags struct {
	Config, Input, Output, APIURL, TokenEnv, Kind, Event, IntentID, MergeSHA, Outcome, Summary, Pin string
	Version, Reason, MintRef                                                                        string
	Issue, FixPR                                                                                    int
	RunID                                                                                           int64
	PR                                                                                              int
}
type productionOperation struct {
	config   promotion.Config
	client   promotion.Client
	snapshot promotion.JournalSnapshot
	flags    productionFlags
	proof    promotion.GitProof
}

func init() {
	names := []string{"status", "candidate", "propose", "validate", "intent", "start", "finish", "published", "bootstrap", "prepare-hotfix", "publish", "version-hotfix", "propose-rollback", "status-pr", "validate-review", "scan", "hotfix", "rollback", "control", "workflow", "report"}
	group := &cobra.Command{Use: "production", Short: "Reconcile reviewed production proposals and exact deployment intents"}
	for _, name := range names {
		var f productionFlags
		operation := name
		cmd := &cobra.Command{Use: operation, Short: productionHelp(operation), Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error { return runProduction(cmd, operation, f) }}
		flags := cmd.Flags()
		flags.StringVar(&f.Config, "config", ".mint.yaml", "repository-owned release policy")
		flags.StringVar(&f.Input, "input", "", "typed manifest JSON file")
		flags.StringVar(&f.Output, "output", "", "write resulting JSON manifest")
		flags.StringVar(&f.APIURL, "api-url", "https://api.github.com", "GitHub API URL")
		flags.StringVar(&f.TokenEnv, "token-env", "GH_TOKEN", "job-scoped GitHub credential environment variable")
		flags.StringVar(&f.Kind, "kind", "normal", "normal or hotfix proposal")
		flags.StringVar(&f.Event, "event", "candidate", "candidate, close or reopen")
		flags.StringVar(&f.IntentID, "intent-id", "", "frozen intent identity")
		flags.StringVar(&f.MergeSHA, "merge-sha", "", "exact reviewed proposal merge SHA")
		flags.StringVar(&f.Outcome, "outcome", "", "verified success or failure")
		flags.StringVar(&f.Summary, "summary", "", "manual summary override")
		flags.StringVar(&f.Pin, "pin", "", "explicit candidate SHA")
		flags.Int64Var(&f.RunID, "run-id", 0, "server-verified workflow run")
		flags.StringVar(&f.Version, "to", "", "previously deployed version; default previous verified deployment")
		flags.StringVar(&f.Reason, "reason", "", "reason for the production recovery request")
		flags.StringVar(&f.MintRef, "mint-ref", "", "published immutable Mint commit for generated Actions workflow")
		flags.IntVar(&f.Issue, "issue", 0, "issue for a newly authored production fix")
		flags.IntVar(&f.FixPR, "fix-pr", 0, "merged reviewed fix PR to isolate from queued main")
		if operation == "status-pr" {
			cmd.Deprecated = "use report; outcomes are attached to the original release PR"
		}
		flags.IntVar(&f.PR, "pr", 0, "exact trusted proposal/source PR number")
		group.AddCommand(cmd)
	}
	releaseCmd.AddCommand(group)
}
func readJSON(path string, value any) error {
	if path == "" {
		return fmt.Errorf("typed input file required")
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if decoder.Decode(new(any)) != io.EOF {
		return fmt.Errorf("trailing input content")
	}
	return nil
}
func runProduction(cmd *cobra.Command, operation string, f productionFlags) error {
	cfg, err := promotion.LoadConfig(f.Config)
	if err != nil {
		return err
	}
	client := promotion.Client{APIURL: f.APIURL, Token: os.Getenv(f.TokenEnv), Repository: cfg.Repository, HumanLogin: cfg.HumanLogin}
	if operation == "workflow" {
		return writeControlWorkflow(cmd, cfg, f)
	}
	if operation != "status" && operation != "validate" && operation != "validate-review" {
		if err := client.VerifyRepository(cmd.Context()); err != nil {
			return err
		}
	}
	snapshot, err := client.LoadJournal(cmd.Context(), cfg.Environment)
	if err != nil {
		return err
	}
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}
	op := productionOperation{config: cfg, client: client, snapshot: snapshot, flags: f, proof: promotion.GitProof{Context: cmd.Context(), WorkDir: workDir}}
	value, mutated, err := op.execute(cmd.Context(), operation)
	if err != nil {
		return err
	}
	if mutated {
		if _, err := client.SaveJournal(cmd.Context(), op.snapshot, "production "+operation); err != nil {
			return err
		}
	}
	if (operation == "propose" || operation == "propose-rollback" || operation == "rollback" || (operation == "control" && oIsProposal(value))) && mutated {
		p := value.(promotion.Proposal)
		if p.State == "open" {
			if err := client.DispatchChecks(cmd.Context(), cfg.ValidationWorkflow, p.Branch, p.PR); err != nil {
				return err
			}
		}
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if f.Output != "" {
		if err := os.WriteFile(f.Output, append(data, '\n'), 0600); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return err
}
func (o *productionOperation) execute(ctx context.Context, operation string) (any, bool, error) {
	s := &o.snapshot.State
	f := o.flags
	switch operation {
	case "hotfix":
		return o.requestHotfix(ctx)
	case "rollback":
		return o.requestRollback(ctx)
	case "control":
		return o.control(ctx)
	case "scan":
		return o.scan(ctx)
	case "validate-review":
		return o.review(ctx)
	case "status":
		return s, false, nil
	case "candidate":
		var c promotion.Candidate
		if err := o.client.RunManifest(ctx, f.RunID, "mint-candidate", &c); err != nil {
			return nil, false, err
		}
		run, err := o.client.TrustedRun(ctx, f.RunID, o.config.BuildWorkflow)
		if err != nil {
			return nil, false, err
		}
		if c.Kind == "normal" && run.HeadBranch != o.config.DefaultBranch {
			return nil, false, fmt.Errorf("normal candidate is not built from default branch")
		}
		if (c.Kind == "normal" && run.HeadSHA != c.SourceSHA) || run.ID != c.RunID || run.HeadBranch != o.config.DefaultBranch {
			return nil, false, fmt.Errorf("candidate does not match successful build")
		}
		if err := o.proof.VerifyTag(c.Version, c.SourceSHA); err != nil {
			return nil, false, err
		}
		if prior, exists := s.Candidates[c.SourceSHA]; exists {
			if c.Repository != prior.Repository || c.Environment != prior.Environment || c.Version != prior.Version || c.Kind != prior.Kind || !reflect.DeepEqual(c.Artifact, prior.Artifact) || c.SourcePR != prior.SourcePR || c.BaselineID != prior.BaselineID {
				return nil, false, fmt.Errorf("replayed source/build conflicts with immutable candidate")
			}
			return prior, false, nil
		}
		if c.Kind == "hotfix" {
			if err := o.client.VerifyHotfixSource(ctx, c, o.config.RequiredChecks); err != nil {
				return nil, false, err
			}
		}
		identity, err := release.VersionIdentity(ctx, o.proof.WorkDir, c.SourceSHA, o.config.ControlPaths)
		if err != nil {
			return nil, false, err
		}
		c.ControlOnly = identity.ControlOnly
		c.RunURL = run.HTMLURL
		c.SourceDate, err = o.proof.SourceDate(c.SourceSHA)
		if err != nil {
			return nil, false, err
		}
		changes, err := o.client.CollectChanges(ctx, o.proof, *s, c, o.config.ControlPaths)
		if err != nil {
			return nil, false, err
		}
		c.Changes = changes
		e := promotion.BuildEvidence{Repository: run.Repository.FullName, SourceSHA: c.SourceSHA, TagSHA: c.SourceSHA, ArtifactDigest: c.Artifact.Digest, Configuration: c.Artifact.Configuration, RunID: run.ID, Success: true, Trusted: true}
		return c, true, s.RegisterCandidate(c, e)
	case "propose-rollback":
		if s.Baseline == nil {
			return nil, false, fmt.Errorf("verified baseline required")
		}
		p, err := s.ReconcileRollback(f.Pin, "rollback-"+s.Baseline.ID, f.Summary)
		if err != nil {
			return nil, false, err
		}
		if p.State != "open" {
			return p, false, nil
		}
		p, err = o.client.SyncProposal(ctx, o.config, s, p)
		return p, true, err
	case "status-pr", "report":
		i, ok := s.Intents[f.IntentID]
		if !ok {
			return nil, false, fmt.Errorf("unknown intent")
		}
		p, err := o.client.SyncStatus(ctx, o.config, i)
		return p, false, err
	case "propose":
		return o.propose(ctx)
	case "validate", "intent":
		return o.intent(ctx, operation == "intent")
	case "start":
		return o.start(ctx)
	case "finish":
		return o.finish(ctx)
	case "published":
		if err := o.client.VerifyPublication(ctx, s.Intents[f.IntentID]); err != nil {
			return nil, false, err
		}
		return s, true, s.MarkPublished(f.IntentID)
	case "publish":
		i, ok := s.Intents[f.IntentID]
		if !ok {
			return nil, false, fmt.Errorf("unknown intent")
		}
		if err := o.client.PublishIntent(ctx, i); err != nil {
			return nil, false, err
		}
		return i, true, s.MarkPublished(i.ID)
	case "bootstrap":
		return o.bootstrap(ctx)
	case "prepare-hotfix":
		return o.hotfix(ctx)
	case "version-hotfix":
		return o.versionHotfix(ctx)
	}
	return nil, false, fmt.Errorf("unsupported production operation %s", strings.TrimSpace(operation))
}

func productionHelp(operation string) string {
	switch operation {
	case "hotfix":
		return "Prepare an isolated fix by reviewed PR or new-fix issue"
	case "rollback":
		return "Propose restoring a verified deployed version; prefer a roll-forward hotfix"
	case "control":
		return "Handle authenticated repository Actions recovery requests"
	case "workflow":
		return "Generate the repository Actions hotfix/rollback entry point"
	case "report", "status-pr":
		return "Record deployment outcome on the original release PR"
	case "status":
		return "Inspect authoritative production state without mutation"
	default:
		return "Run the " + operation + " lifecycle adapter"
	}
}
