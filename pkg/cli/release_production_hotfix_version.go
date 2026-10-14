package cli

import (
	"context"
	"fmt"
	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/jamesonstone/mint/pkg/release"
	"os"
	"os/exec"
	"strings"
)

func (o *productionOperation) versionHotfix(ctx context.Context) (any, bool, error) {
	s := &o.snapshot.State
	if s.Baseline == nil || s.InFlight != "" {
		return nil, false, fmt.Errorf("production must have a verified idle baseline before hotfix versioning")
	}
	pr, err := o.client.Pull(ctx, o.flags.PR)
	if err != nil {
		return nil, false, err
	}
	baseline, marked := o.client.Component.HotfixSourceBaseline(pr.Body)
	if !pr.Merged || !o.client.IsReleaseAuthor(pr.User.Login) || !strings.HasPrefix(pr.Base.Ref, o.client.Component.HotfixBasePrefix()) || !marked || baseline != s.Baseline.ID {
		return nil, false, fmt.Errorf("hotfix source must be reviewed and merged against its production base")
	}
	if err := o.client.CheckHead(ctx, pr, o.config.RequiredChecks); err != nil {
		return nil, false, err
	}
	selected, selectErr := release.SelectTag(ctx, release.SelectTagOptions{Commitish: pr.MergeSHA, WorkDir: o.proof.WorkDir, Scope: o.config.Component.ReleaseScope()})
	if selectErr == nil {
		return promotion.MainVersionSelection{SourceSHA: pr.MergeSHA, Version: selected.VersionTag, SourcePR: pr.Number, BaselineID: s.Baseline.ID}, false, nil
	}
	cmd := exec.CommandContext(ctx, "git", "tag", "--list")
	cmd.Dir = o.proof.WorkDir
	tags, err := cmd.Output()
	if err != nil {
		return nil, false, err
	}
	version, err := release.AllocateHotfixVersion(s.Baseline.Candidate.Version, strings.Fields(string(tags)), o.config.Component.ReleaseScope())
	if err != nil {
		return nil, false, err
	}
	c := promotion.Candidate{SourceSHA: pr.MergeSHA, Version: version, SourcePR: pr.Number, Kind: "hotfix", BaselineID: s.Baseline.ID}
	if _, err := o.client.CollectChanges(ctx, o.proof, *s, c, o.config.ControlPaths); err != nil {
		return nil, false, err
	}
	notes, err := os.CreateTemp("", "mint-hotfix-notes-*")
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = os.Remove(notes.Name()) }()
	if _, err := fmt.Fprintf(notes, "Isolated reviewed hotfix #%d from verified production %s.\n", pr.Number, s.Baseline.Candidate.Version); err != nil {
		_ = notes.Close()
		return nil, false, err
	}
	if err := notes.Close(); err != nil {
		return nil, false, err
	}
	if _, err := release.CreateReleaseTag(ctx, release.TagOptions{Tag: version, Target: pr.MergeSHA, NotesFile: notes.Name(), Remote: "origin", Push: true, WorkDir: o.proof.WorkDir, Scope: o.config.Component.ReleaseScope()}); err != nil {
		return nil, false, err
	}
	return promotion.MainVersionSelection{SourceSHA: pr.MergeSHA, Version: version, SourcePR: pr.Number, BaselineID: s.Baseline.ID}, false, nil
}
