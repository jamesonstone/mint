package cli

import (
	"context"
	"fmt"
	"reflect"

	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/jamesonstone/mint/pkg/release"
)

// candidate authenticates one producer run and reads its manifest once.
func (o *productionOperation) candidate(ctx context.Context) (any, bool, error) {
	run, err := o.client.TrustedBuildRun(ctx, o.flags.RunID, o.config, o.config.BuildWorkflow)
	if err != nil {
		return nil, false, err
	}
	var c promotion.Candidate
	if err := o.client.RunManifest(ctx, run.ID, "mint-candidate", &c); err != nil {
		return nil, false, err
	}
	return o.registerCandidate(ctx, c, run)
}

// registerCandidate shares registration for direct callbacks and recovery scans.
// Callers supply only GitHub-authenticated run and archive evidence.
func (o *productionOperation) registerCandidate(ctx context.Context, c promotion.Candidate, run promotion.WorkflowRun) (any, bool, error) {
	s := &o.snapshot.State
	if run.Repository.FullName != s.Repository || run.HeadRepository.FullName != s.Repository || !promotion.WorkflowPathMatches(run.Path, o.config.BuildWorkflow) || run.Status != "completed" || run.Conclusion != "success" || !o.config.BuildEventAllowed(run.Event) {
		return nil, false, fmt.Errorf("candidate requires a successful trusted producer run")
	}
	// A run that workflow_run started takes no inputs and executes at the
	// default-branch head. It can attest only that head, never an isolated hotfix.
	if run.Event == "workflow_run" && c.Kind != "normal" {
		return nil, false, fmt.Errorf("a producer run started by workflow_run can supply only a normal candidate")
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
	identity, err := release.VersionIdentity(ctx, o.proof.WorkDir, c.SourceSHA, o.config.ControlPaths, o.config.Component.ReleaseScope())
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
}
