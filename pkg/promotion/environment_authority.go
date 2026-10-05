package promotion

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// AuthorityDigest excludes mutable desired selection while freezing workflow,
// operator, checks, configuration, publication and prerequisite authority.
func AuthorityDigest(cfg Config) (string, error) {
	cfg.Target, cfg.Follow, cfg.Operation, cfg.Reason = "", "", "", ""
	data, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(data)), nil
}

// Resume records consumption before clearing the rollback pause. Replaying an
// old resume after a newer rollback cannot resume that later deployment stream.
func (s *State) Resume(key string) (Proposal, bool, error) {
	if key == "" || s.Schema != 2 || s.Baseline == nil || s.InFlight != "" {
		return Proposal{}, false, fmt.Errorf("resume requires a verified idle environment and immutable request identity")
	}
	if prior, ok := s.ControlRequests[key]; ok {
		return prior, false, nil
	}
	if s.ControlRequests == nil {
		s.ControlRequests = map[string]Proposal{}
	}
	request := Proposal{ID: key, Kind: "resume", State: "resumed", BaselineID: s.Baseline.ID}
	s.ControlRequests[key] = request
	s.Paused = false
	if s.Target == "" {
		p := s.Proposals["normal"]
		p.Selection = "latest"
		if p.ID != "" {
			s.Proposals["normal"] = p
		}
	}
	return request, true, nil
}
