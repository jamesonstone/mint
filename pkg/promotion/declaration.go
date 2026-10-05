package promotion

import (
	"encoding/json"
	"fmt"
)

// Declaration is the reviewed release-control document frozen from a merge SHA.
type Declaration struct {
	Schema      int       `json:"schema_version"`
	Repository  string    `json:"repository"`
	Environment string    `json:"environment"`
	ProposalID  string    `json:"proposal_id"`
	Kind        string    `json:"kind"`
	BaselineID  string    `json:"baseline_id"`
	Selection   string    `json:"selection"`
	Candidate   Candidate `json:"candidate"`
	Summary     string    `json:"summary"`
	Notes       string    `json:"notes"`
}

// Declare excludes mutable PR/head metadata from the source declaration.
func (s State) Declare(p Proposal) (Declaration, error) {
	c, ok := s.Candidates[p.CandidateSHA]
	if !ok {
		return Declaration{}, fmt.Errorf("proposal candidate unavailable")
	}
	return Declaration{Schema: 1, Repository: s.Repository, Environment: s.Environment, ProposalID: p.ID, Kind: p.Kind, BaselineID: p.BaselineID, Selection: p.Selection, Candidate: c, Summary: p.Summary, Notes: p.Notes}, nil
}

// ValidateDeclaration binds the complete merged selection to trusted state.
func (s State) ValidateDeclaration(d Declaration) (Proposal, error) {
	p, ok := s.Proposals[d.Kind]
	if !ok {
		return p, fmt.Errorf("unknown proposal kind")
	}
	expected, err := s.Declare(p)
	if err != nil {
		return p, err
	}
	left, _ := json.Marshal(expected)
	right, _ := json.Marshal(d)
	if string(left) != string(right) {
		return p, fmt.Errorf("declaration differs from reconciled proposal; revalidate and review")
	}
	return p, nil
}
