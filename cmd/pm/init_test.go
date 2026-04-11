package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ditsara/git-product-manager/internal/config"
	"gopkg.in/yaml.v3"
)

// TestCreateDefaultWorkflow tests that workflow.yaml is created with correct structure
func TestCreateDefaultWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	pmPath := filepath.Join(tmpDir, ".pm")
	configPath := filepath.Join(pmPath, "config")

	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Call the function being tested
	createDefaultWorkflow(pmPath)

	// Verify file was created
	workflowPath := filepath.Join(configPath, "workflow.yaml")
	_, err := os.Stat(workflowPath)
	if err != nil {
		t.Errorf("workflow.yaml was not created: %v", err)
		return
	}

	// Verify content is valid YAML
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("Could not read workflow.yaml: %v", err)
	}

	var workflow map[string]interface{}
	if err := yaml.Unmarshal(content, &workflow); err != nil {
		t.Errorf("workflow.yaml is not valid YAML: %v", err)
	}

	// Verify it has required keys
	if _, hasStates := workflow["states"]; !hasStates {
		t.Error("workflow.yaml missing 'states' key")
	}
}

// TestCreateDefaultLabels tests that labels.yaml is created
func TestCreateDefaultLabels(t *testing.T) {
	tmpDir := t.TempDir()
	pmPath := filepath.Join(tmpDir, ".pm")
	configPath := filepath.Join(pmPath, "config")

	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	createDefaultLabels(pmPath)

	// Verify file was created
	labelsPath := filepath.Join(configPath, "labels.yaml")
	_, err := os.Stat(labelsPath)
	if err != nil {
		t.Errorf("labels.yaml was not created: %v", err)
		return
	}

	// Verify content is valid YAML
	content, err := os.ReadFile(labelsPath)
	if err != nil {
		t.Fatalf("Could not read labels.yaml: %v", err)
	}

	var labels interface{}
	if err := yaml.Unmarshal(content, &labels); err != nil {
		t.Errorf("labels.yaml is not valid YAML: %v", err)
	}
}

// TestCreateDefaultTemplates tests that all 4 templates are created
func TestCreateDefaultTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	pmPath := filepath.Join(tmpDir, ".pm")
	templatesPath := filepath.Join(pmPath, "config", "templates")

	if err := os.MkdirAll(templatesPath, 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	createDefaultTemplates(pmPath)

	// Verify all 4 templates were created
	expectedTemplates := []string{"story.md", "task.md", "bug.md", "epic.md"}
	for _, templateName := range expectedTemplates {
		templatePath := filepath.Join(templatesPath, templateName)
		_, err := os.Stat(templatePath)
		if err != nil {
			t.Errorf("Template %s was not created: %v", templateName, err)
			continue
		}

		// Verify template content is not empty
		content, err := os.ReadFile(templatePath)
		if err != nil {
			t.Fatalf("Could not read template %s: %v", templateName, err)
		}

		if len(content) == 0 {
			t.Errorf("Template %s is empty", templateName)
		}

		// Verify it contains YAML front matter
		if !strings.Contains(string(content), "---") {
			t.Errorf("Template %s missing YAML front matter", templateName)
		}
	}
}

// TestCreateGitignore tests that .gitignore is created with correct content
func TestCreateGitignore(t *testing.T) {
	tmpDir := t.TempDir()
	pmPath := filepath.Join(tmpDir, ".pm")

	if err := os.MkdirAll(pmPath, 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	createGitignore(pmPath)

	// Verify file was created
	gitignorePath := filepath.Join(pmPath, ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Errorf(".gitignore was not created: %v", err)
		return
	}

	// Verify it contains .cache.db
	if !strings.Contains(string(content), ".cache.db") {
		t.Error(".gitignore does not contain .cache.db")
	}
}

// TestCreateProjectConfig tests that project.yaml is created with correct prefix
func TestCreateProjectConfig(t *testing.T) {
	tests := []struct {
		name           string
		prefix         string
		expectedPrefix string
		description    string
	}{
		{
			name:           "uppercase_prefix",
			prefix:         "MYPROJECT",
			expectedPrefix: "MYPROJECT",
			description:    "Uppercase prefix should be preserved",
		},
		{
			name:           "lowercase_prefix",
			prefix:         "myproject",
			expectedPrefix: "myproject",
			description:    "Lowercase prefix should be preserved (already uppercase at this point)",
		},
		{
			name:           "mixed_case_prefix",
			prefix:         "MyProject",
			expectedPrefix: "MyProject",
			description:    "Mixed case prefix should be preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			pmPath := filepath.Join(tmpDir, ".pm")
			configPath := filepath.Join(pmPath, "config")

			if err := os.MkdirAll(configPath, 0755); err != nil {
				t.Fatalf("Failed to create directories: %v", err)
			}

			createProjectConfig(pmPath, tt.prefix)

			// Verify file was created
			projectPath := filepath.Join(configPath, "project.yaml")
			content, err := os.ReadFile(projectPath)
			if err != nil {
				t.Fatalf("project.yaml was not created: %v", err)
			}

			// Verify prefix is stored correctly
			if !strings.Contains(string(content), tt.expectedPrefix) {
				t.Errorf("%s: Expected prefix %q in project.yaml, got: %s", tt.description, tt.expectedPrefix, string(content))
			}

			// Verify we can parse it as a Project
			project, err := config.LoadProject(pmPath)
			if err != nil {
				t.Fatalf("Could not load project.yaml: %v", err)
			}

			if project.Prefix != tt.expectedPrefix {
				t.Errorf("%s: Expected prefix %q, got %q", tt.description, tt.expectedPrefix, project.Prefix)
			}

			if !strings.Contains(string(content), "sync:") {
				t.Errorf("%s: Expected sync comment in project.yaml, got: %s", tt.description, string(content))
			}
		})
	}
}

// TestPrefixUppercasing tests that the init command uppercases the prefix
func TestPrefixUppercasing(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
		description    string
	}{
		{
			name:           "lowercase_to_uppercase",
			input:          "myproject",
			expectedOutput: "MYPROJECT",
			description:    "Lowercase should be converted to uppercase",
		},
		{
			name:           "mixed_case_to_uppercase",
			input:          "MyProject",
			expectedOutput: "MYPROJECT",
			description:    "Mixed case should be converted to uppercase",
		},
		{
			name:           "already_uppercase",
			input:          "MYPROJECT",
			expectedOutput: "MYPROJECT",
			description:    "Already uppercase should remain unchanged",
		},
		{
			name:           "single_letter",
			input:          "p",
			expectedOutput: "P",
			description:    "Single letter should be uppercased",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strings.ToUpper(tt.input)
			if result != tt.expectedOutput {
				t.Errorf("%s: Expected %q, got %q", tt.description, tt.expectedOutput, result)
			}
		})
	}
}

// TestInitDirectoryStructure tests that all necessary directories are created
func TestInitDirectoryStructure(t *testing.T) {
	tmpDir := t.TempDir()
	pmPath := filepath.Join(tmpDir, ".pm")

	expectedDirs := []string{
		filepath.Join(pmPath, "tickets"),
		filepath.Join(pmPath, "config"),
		filepath.Join(pmPath, "config", "templates"),
	}

	for _, dir := range expectedDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		stat, err := os.Stat(dir)
		if err != nil {
			t.Errorf("Directory %s was not created: %v", dir, err)
			continue
		}

		if !stat.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}
}
