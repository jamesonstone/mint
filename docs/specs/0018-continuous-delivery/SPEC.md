---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "verify"
feature:
  id: "0018"
  slug: "continuous-delivery"
  dir: "0018-continuous-delivery"
---
# SPEC
## PURPOSE
Make Continuous Delivery a usable Mint option for every environment that a project names. A project turns it on for one environment with `deploy: automatic` and `follow: latest`. The latest eligible build then goes to the project adapter with no release pull request. A project must not write its own delivery controller to get this behavior.
## CONTEXT
Schema 2 has the option since spec 0010: an `automatic` environment freezes a new eligible candidate and dispatches the promotion adapter. Three rules stop a common setup from using it.

1. Mint accepts a producer run only when its event is `push` or `workflow_dispatch` (`ObservedRun`, `registerCandidate`, `SuccessfulBuilds`). A project that builds only after its checks pass starts the build with `workflow_run`. Mint refuses that run, so no candidate exists and nothing is delivered.
2. The generated environment controller has one fixed gate, `vars.MINT_RELEASE_ENABLED`. Schema 1 production workflows read the same variable. A repository that keeps a schema 1 production policy and adds a schema 2 policy for other environments cannot turn one on and keep the other off.
3. A build that completes while a deployment is in flight makes the producer callback fail, although the candidate is registered and the completion callback selects the latest candidate later. With frequent merges, the controller of a Continuous Delivery environment fails often for a normal state.

A consumer runs its sandbox delivery beside Mint because of these rules. The documentation also does not use the name Continuous Delivery.

GitHub reports a run that `workflow_run` started with the default branch as `head_branch`, the repository as `head_repository`, and the default-branch head at trigger time as `head_sha`. This is true for each upstream run, also for a pull request from a fork. The run always uses the workflow file from the default branch.
## REQUIREMENTS
- Policy field `build_events` (schema 2, top level) lists the events that can start a trusted producer run: `push`, `workflow_dispatch`, `workflow_run`. An absent field means `push` and `workflow_dispatch`. Reject an empty list, a repeated value, and an unknown value. Schema 1 keeps its strict parser and does not accept the field.
- Candidate registration, the producer callback, the recovery scan, and the baseline build check use the same event set. Promotion, baseline, and observation adapter runs do not change.
- A producer run that `workflow_run` started supplies only a `normal` candidate. Its head commit must be the candidate source commit. All other producer checks stay: repository, head repository, workflow path, default branch, successful completion, version tag on the source commit.
- Policy field `activation_variable` (schema 2, top level) names the repository variable that the generated environment controller reads. An absent field means `MINT_RELEASE_ENABLED`. Accept only an upper-case name that starts with `MINT_`.
- The authority digest of a policy that does not use the new fields does not change. Each new field is part of the digest. The order of `build_events` and an explicitly written default do not change the digest. A target-only request pull request cannot change the fields.
- A `normal` candidate that arrives while a deployment is in flight is registered, and the producer callback reports that reconciliation waits. The callback does not fail while the adapter run of the fenced intent is still active, or when the fence was released in the meantime. Each other refusal stays a failure: a hotfix candidate, a fence that outlives its deployment run, a fence with no recorded run, an intent that is frozen but not started or not settled, and a rollback pause.
- The reference, the public site, and the README name Continuous Delivery, show the policy, and state the producer duty for `workflow_run`.

Non-goals: provider code, a change to version identity, a change to `manual` or `reviewed` selection, per-environment controller gates, consumer activation.
## ACCEPTED PLAN
Add the two fields to `Policy` and to the effective `Config` with `omitempty` JSON names. Add one helper that answers which producer events a policy trusts, and use it at each producer check. Keep `ObservedRun` and `TrustedRun` for adapter runs and add producer variants that take the effective policy. Render the controller gate from the policy. Mark the in-flight deferral with a sentinel error and report it as a result in the producer callback. Add tests for parsing, the digest, each producer path, the generated workflow, and the callback. Update the reference, the site guides, and the README in the same pull request. Publish through the normal reviewed pull request and release before a consumer pins the commit.
## DECISIONS
- Accept `workflow_run` only by policy opt-in. The default stays strict, so an existing policy keeps its trust set and its digest.
- Mint does not check which run started the producer. The run API does not give the upstream run. The producer workflow is default-branch code and owns that check: the triggering run must be a successful `push` run on the default branch of the same repository, and its head commit must be `github.sha`. The guides state this duty and give a sample with a job-level condition. Rejected: a Mint check of required checks on the source commit for `automatic` environments. It changes the meaning of an eligible candidate for all policies and needs a second callback path.
- A chained producer cannot supply a hotfix candidate. A `workflow_run` run takes no inputs, so an isolated hotfix build stays on `workflow_dispatch`.
- One activation variable for each policy file, not for each environment. An environment is already inactive when it has no `configuration_sha256` or no verified baseline, and `deploy` selects the delivery mode. A second gate inside one controller job adds states without a use.
- The activation variable keeps the `MINT_` prefix. The compatibility adapters hash every repository variable outside that prefix into the runtime configuration identity. A switch outside the prefix changes that identity each time it flips, and a schema 1 production policy in the same repository then refuses its artifact.
- The effective event list is sorted, and the default set is stored as absent. The same trust set then has one digest, and a policy author who writes the default does not invalidate pending approvals.
- The in-flight deferral is a result, not a failure, for a `normal` candidate and only while the deployment run is active. The completion callback of a verified deployment reconciles the latest candidate. A hotfix has no such continuation, so its error stays visible. A fence that outlives its run never releases itself: an adapter that ended with no settled outcome, or a controller that was switched off during the deployment. The callback reads the journal and the fenced run, keeps the failure in that state, and points to `reconcile`. Without this rule each later build reports a wait and no failed run shows the stall.
- Version identity does not change. A candidate needs an immutable version tag on its source commit, and each source commit after the anchor needs one. A project that stopped tagging must start again before it activates Continuous Delivery. Rejected: untagged candidates for sandboxes. They remove the identity that rollback, history, and prerequisites use.
- Provider operations stay in project adapters (spec 0010, AC6).
## DISCOVERIES
- Run identity. Public repositories whose `workflow_run` workflows follow pull request workflows show the default branch as `head_branch`, the base repository as `head_repository`, and the default-branch head as `head_sha` on each chained run. Mint's branch and repository checks therefore pass for a chained run after every upstream run. The producer workflow must make the upstream check.
- Refusal form. A chained producer that ends `success` with no `mint-candidate` makes the callback fail with a missing manifest on an activated environment. `skipped` and `failure` give a clean result. The guides require a job-level condition.
- Liveness. A chained producer builds a commit only while it is still the default-branch head when its checks finish. Merges that arrive faster than the checks run give no candidate until the newest commit passes.
- Authority digest. The digest is the SHA-256 of the JSON form of the effective `Config`. A new field changes the digest of each existing policy unless its JSON name is omitted when unset. The tests pin the v1.0.3 digests and the v1.0.3 controller text of the shared policy fixture.
- Rollout. Releases before this change reject a policy that contains either field. A change to either field changes the digest of every environment in the policy file, so a pending approval or deployment start is refused. The guides give the order.
- Version tags. Candidate registration and change collection need a version tag on the candidate source commit and on each non-control first-parent commit after the anchor. A consumer that stopped `version-main` gets one tag for each untagged first-parent commit after its newest tagged ancestor on the first producer run. This is existing behavior.
- Start of deployment. A schema 2 adapter run must call the `start` operation itself after it uploads `mint-request`. The completion callback finds the intent by the recorded deployment run. No change was necessary.
- Documentation tools. `make docs-check` uses `python3`. The local default interpreter (Python 3.14) has no Markdown package. The checks ran with Python 3.13.12 and Markdown 3.11, the versions that the `Documentation` workflow uses.
## VALIDATION
Local, on the branch head, macOS, Go 1.26.11:
- PASS `go test ./...`; PASS `go test -race ./pkg/promotion ./pkg/cli`.
- PASS `python3 -m unittest discover -s adapters/tests` (26 tests).
- PASS `go vet ./...`; PASS `make build`; PASS `gofmt -l cmd pkg` (no output).
- PASS documentation tests (3 tests) and `scripts/build_docs.py` with Python 3.13.12 and Markdown 3.11: 8 pages, links and anchors valid; `llms.txt` and `llms-full.txt` name Continuous Delivery.
- PASS built binary: `mint deployment policy` shows the two fields and shows no `BuildEvents` for an explicitly written default; `mint deployment workflow` writes `vars.MINT_SANDBOX_ENABLED == 'true'` for a policy that names it and `vars.MINT_RELEASE_ENABLED == 'true'` for a policy that does not; `actionlint` accepts the generated controller and the sample producer workflow; `GITHUB_TOKEN` and `CD_ENABLED` as the variable and `pull_request_target` as a producer event are refused.
- New tests: `pkg/promotion/build_events_test.go` and `pkg/cli/build_events_test.go`, plus one scenario in the completion callback test. They cover parsing and rejection, the v1.0.3 digests and controller text, digest binding and normalization, target-only request refusal at both comparison points, producer run authentication, the recovery scan, candidate registration, the producer callback with the in-flight result, the schema 1 source adapter, and baseline import.
- Fresh-context review, two passes: no blocking finding in the trust change. It built v1.0.3 and the branch and found the digest, the effective policy JSON, and the generated controller identical for seven policies without the fields. Its mutants of each producer check, of the digest binding, of the validation, and of the in-flight result were caught. Its findings are applied in this change: the guides, the `MINT_` prefix, digest order, four untested paths, and the rule that a fence which outlives its run stays a failure. The tests for that last rule were written after the second pass; two mutants of the rule in a scratch copy failed them.
- Hosted pull request checks: not observed when this section was written.
- Not checked: a real GitHub `workflow_run` producer against a live repository, and a live automatic deployment. Both need a consumer that pins the published commit.
## OUTCOME
Built: `build_events`, `activation_variable`, producer run variants, the policy-driven controller gate, the in-flight result in the producer callback, tests, and guides. Behavior of a policy that uses neither field is unchanged, except that a `normal` build behind a deployment whose run is still active no longer fails the producer callback. No release was published and no consumer was changed. Remaining risk: a consumer producer that does not check its triggering run can deliver a default-branch commit whose checks did not pass, and a producer that uses content of the triggering run can register that content; Mint cannot detect either case.
## REPOSITORY MEMORY
- Created: this specification.
- Updated: `docs/references/environment-lifecycle.md` (Continuous Delivery, producer events, activation switch, rollout), `website/content/environments.md`, `website/content/integration.md`, `website/content/how-it-works.md`, and `README.md`.
- Unchanged: `docs/references/rules/mint-deployment-lifecycle.md` is Kit-managed and already names the automatic path. `docs/CONSTITUTION.md` has no new invariant. Spec 0010 stays the record of the environment model.
