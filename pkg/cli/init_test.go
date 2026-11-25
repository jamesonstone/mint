package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestGuidedInitDefaults(t *testing.T) {
	t.Chdir(t.TempDir())
	var output bytes.Buffer
	cmd := newInitCommand()
	cmd.SetIn(strings.NewReader("\n\n\n\n\ny\n"))
	cmd.SetOut(&output)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	file, err := os.Open(initConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	cfg, err := readPipelineConfig(file)
	if err != nil {
		t.Fatal(err)
	}
	want := []pipelineStage{
		{Name: "dev", Deployment: "continuous"},
		{Name: "stg", Deployment: "continuous"},
		{Name: "prd", Deployment: "manual"},
	}
	if cfg.Artifact != "binary" || cfg.Deployment != "" || !slices.Equal(cfg.Environments, want) {
		t.Fatalf("config = %#v, want a binary and stages %v", cfg, want)
	}
	for _, path := range []string{initConfigPath, initDeployPath, initPromotionPath} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() == 0 || !strings.Contains(output.String(), path) {
			t.Fatalf("%s was empty or absent from the preview", path)
		}
	}
}

func TestGuidedInitCustomChoicesAndRetries(t *testing.T) {
	t.Chdir(t.TempDir())
	var output bytes.Buffer
	input := "zip\nimage\nqa,qa\nqa, canary\nauto\nmanual\ncontinuous\nyes\n"
	if err := runGuidedInit(strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	cfg, err := os.ReadFile(initConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	config, err := readPipelineConfig(bytes.NewReader(cfg))
	if err != nil {
		t.Fatal(err)
	}
	want := []pipelineStage{
		{Name: "qa", Deployment: "manual"},
		{Name: "canary", Deployment: "continuous"},
	}
	if config.Artifact != "image" || !slices.Equal(config.Environments, want) {
		t.Fatalf("config = %#v, want an image and stages %v", config, want)
	}
	for _, message := range []string{"enter binary or image", "duplicate environment", "enter continuous or manual"} {
		if !strings.Contains(output.String(), message) {
			t.Errorf("missing validation message %q", message)
		}
	}
}

func TestGuidedInitWritesNothingWithoutConfirmation(t *testing.T) {
	for _, test := range []struct {
		name  string
		input string
		err   bool
	}{
		{name: "decline", input: "\n\n\n\n\nn\n"},
		{name: "default decline", input: "\n\n\n\n\n\n"},
		{name: "EOF before confirmation", input: "\n\n\n\n\n", err: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			var output bytes.Buffer
			err := runGuidedInit(strings.NewReader(test.input), &output)
			if (err != nil) != test.err {
				t.Fatalf("error = %v, want error = %v", err, test.err)
			}
			entries, err := os.ReadDir(".")
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) > 0 {
				t.Fatalf("created files before confirmation: %v", entries)
			}
		})
	}
}

func TestGuidedInitPreservesExistingFiles(t *testing.T) {
	for _, path := range []string{initConfigPath, initDeployPath, initPromotionPath} {
		t.Run(path, func(t *testing.T) {
			t.Chdir(t.TempDir())
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			original := []byte("preexisting\n")
			if err := os.WriteFile(path, original, 0644); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			err := runGuidedInit(strings.NewReader("\n\n\n\n\ny\n"), &output)
			if err == nil || !strings.Contains(err.Error(), path+" already exists") {
				t.Fatalf("error = %v, want existing %s", err, path)
			}
			data, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(data, original) {
				t.Fatalf("existing file = %q, error = %v", data, err)
			}
			for _, other := range []string{initConfigPath, initDeployPath, initPromotionPath} {
				if other != path {
					if _, err := os.Stat(other); !os.IsNotExist(err) {
						t.Errorf("unexpected file %s, error = %v", other, err)
					}
				}
			}
		})
	}
}

func TestInitFileInstallRollsBackOnCollision(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(filepath.Dir(initDeployPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(initDeployPath, []byte("created by another process"), 0644); err != nil {
		t.Fatal(err)
	}
	files := []initFile{
		{path: initConfigPath, content: []byte("new config")},
		{path: initDeployPath, content: []byte("new workflow")},
	}
	if err := installInitFiles(files); err == nil {
		t.Fatal("expected collision")
	}
	if _, err := os.Stat(initConfigPath); !os.IsNotExist(err) {
		t.Fatalf("partial config remained, error = %v", err)
	}
	content, err := os.ReadFile(initDeployPath)
	if err != nil || string(content) != "created by another process" {
		t.Fatalf("existing workflow = %q, error = %v", content, err)
	}
}
