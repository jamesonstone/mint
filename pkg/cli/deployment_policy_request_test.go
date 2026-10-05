package cli

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jamesonstone/mint/pkg/promotion"
)

const reviewedEnvironmentYAML = `schema_version: 2
mode: deployment
repository: owner/repo
default_branch: main
human_login: human
build_workflow: .github/workflows/build.yaml
validation_workflow: .github/workflows/checks.yaml
required_checks: [native]
environments:
  production:
    scope: shared
    deploy: reviewed
    follow: latest
    configuration_sha256: sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc
    promotion_workflow: .github/workflows/deploy.yaml
    publish: true
`

func TestYAMLRequestFreezesTheSameReviewedPRUnderPreChangePolicy(t *testing.T) {
	for _, kind := range []string{"fixed", "canonical-latest"} {
		for _, scenario := range []string{"valid", "stale-approval", "policy-downgrade", "tampered-declaration"} {
			if kind == "fixed" && scenario == "tampered-declaration" {
				continue
			}
			t.Run(kind+"/"+scenario, func(t *testing.T) {
				op, c, run := candidateFixture(t)
				c.Environment = "build"
				c.ControlOnly = false
				c.RunURL = run.HTMLURL
				policy, err := promotion.ParsePolicy([]byte(reviewedEnvironmentYAML))
				if err != nil {
					t.Fatal(err)
				}
				cfg, err := policy.ForEnvironment("production")
				if err != nil {
					t.Fatal(err)
				}
				op.config = cfg
				op.snapshot.State.Schema = 2
				op.snapshot.State.Configuration = cfg.Configuration
				op.snapshot.State.Publish = cfg.Publish
				evidence := promotion.BuildEvidence{Repository: c.Repository, SourceSHA: c.SourceSHA, TagSHA: c.SourceSHA, ArtifactDigest: c.Artifact.Digest, Configuration: c.Artifact.Configuration, RunID: c.RunID, Success: true, Trusted: true}
				if err := op.snapshot.State.RegisterCandidate(c, evidence); err != nil {
					t.Fatal(err)
				}
				op.flags.Config, op.flags.PR, op.flags.MergeSHA = ".mint.yaml", 42, strings.Repeat("f", 40)
				pr := promotion.PullRequest{Number: 42, Merged: true, MergeSHA: op.flags.MergeSHA}
				pr.Head.SHA = strings.Repeat("d", 40)
				pr.Head.Ref = "GH-42"
				pr.Head.Repo.FullName = "owner/repo"
				pr.Base.Ref = "main"
				pr.Base.SHA = op.snapshot.State.Baseline.Candidate.SourceSHA
				pr.Base.Repo.FullName = "owner/repo"
				pr.User.Login = promotion.AutomationLogin
				after := strings.Replace(reviewedEnvironmentYAML, "follow: latest", "target: v1.0.1", 1)
				var declaration promotion.Declaration
				if kind == "canonical-latest" {
					after = reviewedEnvironmentYAML
					after = strings.Replace(after, "    follow: latest", "    follow: latest\n    reason: Approved latest build", 1)
					authority := cfg
					authority.PolicyPath = op.flags.Config
					authority.ControlPaths = append(authority.ControlPaths, op.flags.Config)
					op.snapshot.State.PolicyDigest, err = promotion.AuthorityDigest(authority)
					if err != nil {
						t.Fatal(err)
					}
					proposal, err := op.snapshot.State.Reconcile("normal", "candidate", "", "production:latest", "Approved latest build", op.proof)
					if err != nil {
						t.Fatal(err)
					}
					proposal.PR, proposal.Branch, proposal.HeadSHA = pr.Number, pr.Head.Ref, pr.Head.SHA
					op.snapshot.State.Proposals["normal"] = proposal
					declaration, err = op.snapshot.State.Declare(proposal)
					if err != nil {
						t.Fatal(err)
					}
					pr.Body = fmt.Sprintf("<!-- mint:proposal:%s:normal -->", proposal.ID)
					if scenario == "tampered-declaration" {
						declaration.Candidate.Artifact.Reference = "unreviewed/object"
					}
				}
				if scenario == "policy-downgrade" {
					after = strings.Replace(after, "deploy: reviewed", "deploy: automatic", 1)
				}
				declarationData, _ := json.Marshal(declaration)
				calls := map[string]int{}
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != "GET" {
						t.Errorf("YAML request created another approval artifact: %s %s", r.Method, r.URL.Path)
						w.WriteHeader(500)
						return
					}
					calls[r.URL.Path]++
					switch {
					case r.URL.Path == "/repos/owner/repo/pulls/42":
						_ = json.NewEncoder(w).Encode(pr)
					case r.URL.Path == "/repos/owner/repo/git/commits/"+pr.MergeSHA:
						_ = json.NewEncoder(w).Encode(map[string]any{"sha": pr.MergeSHA, "parents": []any{map[string]string{"sha": pr.Base.SHA}}})
					case r.URL.Path == "/repos/owner/repo/pulls/42/commits":
						_ = json.NewEncoder(w).Encode([]any{map[string]string{"sha": pr.Head.SHA}})
					case r.URL.Path == "/repos/owner/repo/git/commits/"+pr.Head.SHA:
						_ = json.NewEncoder(w).Encode(map[string]any{"sha": pr.Head.SHA, "parents": []any{map[string]string{"sha": pr.Base.SHA}}})
					case r.URL.Path == "/repos/owner/repo/contents/.mint.yaml":
						data := after
						if r.URL.Query().Get("ref") == pr.Base.SHA {
							data = reviewedEnvironmentYAML
						}
						_ = json.NewEncoder(w).Encode(map[string]string{"encoding": "base64", "content": base64.StdEncoding.EncodeToString([]byte(data))})
					case r.URL.Path == "/repos/owner/repo/contents/.mint/proposal.json":
						_ = json.NewEncoder(w).Encode(map[string]string{"encoding": "base64", "content": base64.StdEncoding.EncodeToString(declarationData)})
					case r.URL.Path == "/repos/owner/repo/pulls/42/files":
						_ = json.NewEncoder(w).Encode([]any{map[string]string{"filename": ".mint.yaml", "status": "modified"}})
					case strings.HasSuffix(r.URL.Path, "/check-runs"):
						if !strings.Contains(r.URL.Path, pr.Head.SHA) {
							t.Error("checks not bound to reviewed head", r.URL.Path)
						}
						_ = json.NewEncoder(w).Encode(map[string]any{"total_count": 1, "check_runs": []any{map[string]string{"name": "native", "status": "completed", "conclusion": "success"}}})
					case r.URL.Path == "/repos/owner/repo/pulls/42/reviews":
						approvedSHA := pr.Head.SHA
						if scenario == "stale-approval" {
							approvedSHA = pr.Base.SHA
						}
						_ = json.NewEncoder(w).Encode([]any{map[string]any{"state": "APPROVED", "commit_id": approvedSHA, "user": map[string]string{"login": "reviewer"}}})
					default:
						t.Error("unexpected request", r.URL.Path)
						w.WriteHeader(500)
					}
				}))
				defer server.Close()
				op.client = promotion.Client{APIURL: server.URL, Token: "test", Repository: "owner/repo", HumanLogin: "human"}
				result, changed, err := op.policyRequest(context.Background(), true)
				if scenario != "valid" {
					if err == nil || changed || len(op.snapshot.State.Intents) != 0 {
						t.Fatal("unreviewed or weakened request froze", result, changed, err)
					}
					return
				}
				if err != nil || !changed {
					t.Fatal(result, changed, err)
				}
				intent, ok := result.(promotion.Intent)
				if !ok || intent.MergeSHA != pr.MergeSHA || intent.Candidate.SourceSHA != c.SourceSHA || intent.Configuration != cfg.Configuration || !intent.Publish || intent.Environment != "production" {
					t.Fatal("request did not freeze exact reviewed intent", result)
				}
				if op.snapshot.State.Proposals["normal"].PR != 42 || calls["/repos/owner/repo/pulls/42/reviews"] != 1 {
					t.Fatal("same PR/current approval not used", calls)
				}
				if kind == "fixed" && op.snapshot.State.Proposals["normal"].Selection != "pinned" {
					t.Fatal("fixed request resumed tracking", op.snapshot.State.Proposals["normal"])
				}
				if kind == "canonical-latest" && op.snapshot.State.Proposals["normal"].Selection != "latest" {
					t.Fatal("latest proposal lost canonical selection")
				}
			})
		}
	}
}
