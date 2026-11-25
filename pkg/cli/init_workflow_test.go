package cli

import (
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type renderedInitWorkflow struct {
	On map[string]struct {
		Inputs map[string]struct {
			Options []string
		}
	}
	Permissions map[string]string
	Jobs        map[string]struct {
		If          string
		Needs       []string
		Uses        string
		With        map[string]string
		Outputs     map[string]string
		Environment string
		Steps       []struct {
			Run  string
			Uses string
		}
	}
}

const exactBranchCheck = `[[ "$GITHUB_REF" != "refs/heads/$DEFAULT_BRANCH" ]]`

func TestInitWorkflowRouting(t *testing.T) {
	for _, tc := range []struct {
		name       string
		stages     []pipelineStage
		continuous []string
		manual     []string
	}{
		{
			name: "mixed",
			stages: []pipelineStage{
				{Name: "dev", Deployment: "continuous"},
				{Name: "test", Deployment: "manual"},
				{Name: "stg", Deployment: "continuous"},
				{Name: "prd", Deployment: "manual"},
			},
			continuous: []string{"dev", "stg"},
			manual:     []string{"test", "prd"},
		},
		{
			name: "all continuous",
			stages: []pipelineStage{
				{Name: "stg", Deployment: "continuous"},
				{Name: "prd", Deployment: "continuous"},
			},
			continuous: []string{"stg", "prd"},
		},
		{
			name: "all manual",
			stages: []pipelineStage{
				{Name: "stg", Deployment: "manual"},
				{Name: "prd", Deployment: "manual"},
			},
			manual: []string{"stg", "prd"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			deploy, promote := parseInitWorkflows(t, tc.stages)
			if _, ok := deploy.On["push"]; !ok {
				t.Fatal("missing build trigger")
			}
			guard := "github.ref == format('refs/heads/{0}', github.event.repository.default_branch)"
			if got := deploy.Jobs["build"].If; got != "github.event_name == 'push' && "+guard {
				t.Errorf("build job guard = %q", got)
			}
			if got := deploy.Jobs["build"].Outputs["version"]; got != "${{ steps.retain.outputs.version }}" {
				t.Errorf("build job must pass the retained version to promotion, got %q", got)
			}
			steps := deploy.Jobs["build"].Steps
			if len(steps) != 4 || !strings.Contains(steps[1].Run, exactBranchCheck) ||
				!strings.Contains(steps[1].Run, "bash .github/mint/check") ||
				!strings.Contains(steps[3].Run, "bash .github/mint/build") {
				t.Error("build job must check the project before building")
			}

			const reusable = "./.github/workflows/mint-promote.yml"
			wantJobs := 1 + len(tc.continuous)
			dependency := []string{"build"}
			for _, name := range tc.continuous {
				job := deploy.Jobs["continuous-"+name]
				if !reflect.DeepEqual(job.Needs, dependency) || job.Uses != reusable || job.With["environment"] != name ||
					job.With["artifact"] != "${{ needs.build.outputs.artifact }}" ||
					job.With["version"] != "${{ needs.build.outputs.version }}" {
					t.Errorf("unexpected push promotion for %s: %+v", name, job)
				}
				dependency = []string{"build", "continuous-" + name}
			}
			dispatch, manual := deploy.On["workflow_dispatch"]
			if manual != (len(tc.manual) > 0) {
				t.Fatalf("manual event presence = %t", manual)
			}
			if manual {
				wantJobs += 2
				if got := dispatch.Inputs["environment"].Options; !reflect.DeepEqual(got, tc.manual) {
					t.Errorf("manual targets = %v, want %v", got, tc.manual)
				}
				resolver := deploy.Jobs["select-retained"]
				if resolver.If != "github.event_name == 'workflow_dispatch' && "+guard || len(resolver.Needs) > 0 {
					t.Errorf("manual resolution must be independent from build: %+v", resolver)
				}
				script := resolver.Steps[1].Run
				if !strings.Contains(script, exactBranchCheck) {
					t.Error("manual resolution must check the exact default branch")
				}
				if !strings.Contains(script, ".version != $version") {
					t.Error("manual resolution must reject an artifact for a different version")
				}
				previous := ""
				for _, stage := range tc.stages {
					if stage.Deployment == "manual" && !strings.Contains(script, "\""+stage.Name+"\") prerequisite=\""+previous+"\"") {
						t.Errorf("manual prerequisite for %s must be %q", stage.Name, previous)
					}
					previous = stage.Name
				}
				job := deploy.Jobs["manual"]
				if job.Uses != reusable || !reflect.DeepEqual(job.Needs, []string{"select-retained"}) ||
					job.With["artifact"] != "${{ needs.select-retained.outputs.artifact }}" {
					t.Errorf("unexpected manual promotion: %+v", job)
				}
			}
			if len(deploy.Jobs) != wantJobs {
				t.Errorf("deployment job count = %d, want %d", len(deploy.Jobs), wantJobs)
			}
			if promote.Jobs["promote"].Environment != "${{ inputs.environment }}" {
				t.Error("promotion job must use the selected GitHub environment")
			}
			if deploy.Permissions["contents"] != "read" || promote.Permissions["contents"] != "read" {
				t.Error("scaffold grants more than read-only repository access")
			}
		})
	}
}

func TestInitWorkflowOnlyRetainsEvidenceAfterVerification(t *testing.T) {
	_, promote := parseInitWorkflows(t, []pipelineStage{{Name: "stg", Deployment: "continuous"}})
	steps := promote.Jobs["promote"].Steps
	if len(steps) != 4 || !strings.HasPrefix(steps[3].Uses, "actions/upload-artifact@") {
		t.Fatal("promotion must upload evidence in its final step")
	}
	script := steps[2].Run
	if !strings.Contains(script, exactBranchCheck) {
		t.Error("promotion must check the exact default branch")
	}
	deployAt := strings.Index(script, "bash .github/mint/deploy")
	observeAt := strings.Index(script, "bash .github/mint/verify")
	compareAt := strings.Index(script, "mint verify --expected")
	evidenceAt := strings.Index(script, "mint-verified.json")
	if deployAt < 0 || observeAt <= deployAt || compareAt <= observeAt || evidenceAt <= compareAt {
		t.Fatalf("expected deploy, runtime observation, comparison, then receipt in order:\n%s", script)
	}
}

func parseInitWorkflows(t *testing.T, stages []pipelineStage) (renderedInitWorkflow, renderedInitWorkflow) {
	t.Helper()
	deploy, promote, err := renderInitWorkflows(stages)
	if err != nil {
		t.Fatal(err)
	}
	var deployWorkflow, promoteWorkflow renderedInitWorkflow
	if err := yaml.Unmarshal(deploy, &deployWorkflow); err != nil {
		t.Fatalf("parse deployment YAML: %v\n%s", err, deploy)
	}
	if err := yaml.Unmarshal(promote, &promoteWorkflow); err != nil {
		t.Fatalf("parse promotion YAML: %v\n%s", err, promote)
	}
	return deployWorkflow, promoteWorkflow
}
