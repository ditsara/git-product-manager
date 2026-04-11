package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProject(t *testing.T) {
	tempDir := t.TempDir()
	pmPath := filepath.Join(tempDir, ".pm")
	configDir := filepath.Join(pmPath, "config")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	projectYAML := `prefix: TESTPROJECT`
	projectPath := filepath.Join(configDir, "project.yaml")
	if err := os.WriteFile(projectPath, []byte(projectYAML), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	project, err := LoadProject(pmPath)
	if err != nil {
		t.Fatalf("LoadProject() unexpected error = %v", err)
	}

	if project.Prefix != "TESTPROJECT" {
		t.Errorf("LoadProject() Prefix = %v, want 'TESTPROJECT'", project.Prefix)
	}
}

func TestLoadProjectWithSyncConfig(t *testing.T) {
	tempDir := t.TempDir()
	pmPath := filepath.Join(tempDir, ".pm")
	configDir := filepath.Join(pmPath, "config")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	projectYAML := `prefix: TESTPROJECT
sync:
  branch: main
  auto_sync:
    pull_on_list: true
    push_on_change: false
  conflict_strategy: ours
`
	projectPath := filepath.Join(configDir, "project.yaml")
	if err := os.WriteFile(projectPath, []byte(projectYAML), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	project, err := LoadProject(pmPath)
	if err != nil {
		t.Fatalf("LoadProject() unexpected error = %v", err)
	}

	if project.Sync.Branch != "main" {
		t.Fatalf("LoadProject() Sync.Branch = %q, want main", project.Sync.Branch)
	}
	if !project.Sync.AutoSync.PullOnList {
		t.Error("LoadProject() Sync.AutoSync.PullOnList = false, want true")
	}
	if project.Sync.AutoSync.PushOnChange {
		t.Error("LoadProject() Sync.AutoSync.PushOnChange = true, want false")
	}
	if project.Sync.ConflictStrategy != "ours" {
		t.Fatalf("LoadProject() Sync.ConflictStrategy = %q, want ours", project.Sync.ConflictStrategy)
	}
}

func TestLoadProjectInvalidPath(t *testing.T) {
	_, err := LoadProject("/nonexistent/path")
	if err == nil {
		t.Error("LoadProject() expected error for nonexistent path, got nil")
	}
}

func TestLoadProjectInvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	pmPath := filepath.Join(tempDir, ".pm")
	configDir := filepath.Join(pmPath, "config")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	invalidYAML := `prefix: [[[invalid`
	projectPath := filepath.Join(configDir, "project.yaml")
	if err := os.WriteFile(projectPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := LoadProject(pmPath)
	if err == nil {
		t.Error("LoadProject() expected error for invalid YAML, got nil")
	}
}

func TestSaveProject(t *testing.T) {
	tempDir := t.TempDir()
	pmPath := filepath.Join(tempDir, ".pm")
	configDir := filepath.Join(pmPath, "config")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	project := &Project{
		Prefix: "MYPROJECT",
	}

	err := SaveProject(pmPath, project)
	if err != nil {
		t.Fatalf("SaveProject() unexpected error = %v", err)
	}

	// Verify file was created
	projectPath := filepath.Join(configDir, "project.yaml")
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Error("SaveProject() did not create project.yaml file")
	}

	// Load it back and verify
	loaded, err := LoadProject(pmPath)
	if err != nil {
		t.Fatalf("LoadProject() after save unexpected error = %v", err)
	}

	if loaded.Prefix != "MYPROJECT" {
		t.Errorf("SaveProject() then LoadProject() Prefix = %v, want 'MYPROJECT'", loaded.Prefix)
	}
}

func TestSaveProjectWithSyncConfig(t *testing.T) {
	tempDir := t.TempDir()
	pmPath := filepath.Join(tempDir, ".pm")
	configDir := filepath.Join(pmPath, "config")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	project := &Project{
		Prefix: "MYPROJECT",
		Sync: Sync{
			Branch: "main",
			AutoSync: AutoSync{
				PullOnList:   true,
				PushOnChange: true,
			},
			ConflictStrategy: "theirs",
		},
	}

	if err := SaveProject(pmPath, project); err != nil {
		t.Fatalf("SaveProject() unexpected error = %v", err)
	}

	loaded, err := LoadProject(pmPath)
	if err != nil {
		t.Fatalf("LoadProject() after save unexpected error = %v", err)
	}

	if loaded.Sync.Branch != "main" {
		t.Fatalf("loaded Sync.Branch = %q, want main", loaded.Sync.Branch)
	}
	if !loaded.Sync.AutoSync.PullOnList || !loaded.Sync.AutoSync.PushOnChange {
		t.Fatalf("loaded Sync.AutoSync = %+v, want both true", loaded.Sync.AutoSync)
	}
	if loaded.Sync.ConflictStrategy != "theirs" {
		t.Fatalf("loaded Sync.ConflictStrategy = %q, want theirs", loaded.Sync.ConflictStrategy)
	}
}

func TestSaveProjectInvalidPath(t *testing.T) {
	project := &Project{
		Prefix: "TEST",
	}

	// Try to save to a nonexistent directory without creating it
	err := SaveProject("/nonexistent/invalid/path", project)
	if err == nil {
		t.Error("SaveProject() expected error for invalid path, got nil")
	}
}
