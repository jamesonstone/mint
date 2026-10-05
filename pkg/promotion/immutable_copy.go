package promotion

// Detach immutable object identities at journal/API boundaries. Maps and slices
// in a bundle or provenance list must never change an approved intent by alias.
func cloneCandidate(c Candidate) Candidate {
	if c.Artifact.Members != nil {
		members := make(map[string]ArtifactObject, len(c.Artifact.Members))
		for name, member := range c.Artifact.Members {
			members[name] = member
		}
		c.Artifact.Members = members
	}
	if c.Changes != nil {
		c.Changes = append([]Change{}, c.Changes...)
	}
	return c
}
func cloneIntent(i Intent) Intent { i.Candidate = cloneCandidate(i.Candidate); return i }
func cloneBaseline(b Baseline) Baseline {
	b.Candidate = cloneCandidate(b.Candidate)
	if b.Shipped != nil {
		b.Shipped = append([]Change{}, b.Shipped...)
	}
	return b
}
