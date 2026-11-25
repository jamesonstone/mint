package cli

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed init_templates/*.yml
var initWorkflowTemplates embed.FS

type initialPromotion struct {
	Name         string
	Prerequisite string
}

func renderInitWorkflows(stages []pipelineStage) ([]byte, []byte, error) {
	plan := struct {
		Continuous []initialPromotion
		Manual     []initialPromotion
	}{}
	var previous, previousContinuous string
	for _, stage := range stages {
		if !environmentName.MatchString(stage.Name) {
			return nil, nil, fmt.Errorf("invalid environment name %q", stage.Name)
		}
		switch stage.Deployment {
		case "continuous":
			plan.Continuous = append(plan.Continuous, initialPromotion{stage.Name, previousContinuous})
			previousContinuous = stage.Name
		case "manual":
			plan.Manual = append(plan.Manual, initialPromotion{stage.Name, previous})
		default:
			return nil, nil, fmt.Errorf("invalid deployment mode %q for %s", stage.Deployment, stage.Name)
		}
		previous = stage.Name
	}

	render, err := template.New("init").Delims("<<:", ":>>").ParseFS(initWorkflowTemplates, "init_templates/mint-deploy.yml")
	if err != nil {
		return nil, nil, fmt.Errorf("parse deployment workflow template: %w", err)
	}
	var deploy bytes.Buffer
	if err := render.ExecuteTemplate(&deploy, "mint-deploy.yml", plan); err != nil {
		return nil, nil, fmt.Errorf("render deployment workflow: %w", err)
	}
	promote, err := initWorkflowTemplates.ReadFile("init_templates/mint-promote.yml")
	if err != nil {
		return nil, nil, fmt.Errorf("read promotion workflow template: %w", err)
	}
	return deploy.Bytes(), promote, nil
}
