package promotion

import (
	"fmt"
	"sort"
)

// defaultBuildEvents is the strict producer trust set of every policy that does
// not list build_events.
var defaultBuildEvents = []string{"push", "workflow_dispatch"}

// validateBuildEvents accepts an absent list or a nonempty unique subset of the
// producer events. A run that workflow_run started executes default-branch
// workflow code at the default-branch head for every upstream run, so the
// producer itself must check what triggered it; Mint cannot see that run.
func validateBuildEvents(events []string) error {
	if events == nil {
		return nil
	}
	if len(events) == 0 {
		return fmt.Errorf("build_events must list at least one producer event")
	}
	seen := map[string]bool{}
	for _, event := range events {
		if event != "push" && event != "workflow_dispatch" && event != "workflow_run" {
			return fmt.Errorf("build_events permits push, workflow_dispatch and workflow_run")
		}
		if seen[event] {
			return fmt.Errorf("duplicate build event %s", event)
		}
		seen[event] = true
	}
	return nil
}

// effectiveBuildEvents is the authority form of build_events: sorted, and nil
// for the default set, so order and an explicit default do not change the
// authority digest.
func effectiveBuildEvents(events []string) []string {
	sorted := append([]string(nil), events...)
	sort.Strings(sorted)
	if len(sorted) == len(defaultBuildEvents) {
		same := true
		for i, event := range sorted {
			same = same && event == defaultBuildEvents[i]
		}
		if same {
			return nil
		}
	}
	return sorted
}

// BuildEventAllowed reports whether policy trusts a producer run that event
// started. Promotion, baseline and observation adapters keep their fixed sets.
func (cfg Config) BuildEventAllowed(event string) bool {
	events := cfg.BuildEvents
	if len(events) == 0 {
		events = defaultBuildEvents
	}
	for _, allowed := range events {
		if event == allowed {
			return true
		}
	}
	return false
}
