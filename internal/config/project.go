package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Project struct {
	Prefix         string   `yaml:"prefix"`
	AssigneeDomain string   `yaml:"assignee_domain,omitempty"`
	Members        []string `yaml:"members,omitempty"`
	Sync           Sync     `yaml:"sync,omitempty"`
}

type Sync struct {
	Branch           string   `yaml:"branch,omitempty"`
	AutoSync         AutoSync `yaml:"auto_sync,omitempty"`
	ConflictStrategy string   `yaml:"conflict_strategy,omitempty"`
}

type AutoSync struct {
	PullOnList   bool `yaml:"pull_on_list,omitempty"`
	PushOnChange bool `yaml:"push_on_change,omitempty"`
}

func LoadProject(pmPath string) (*Project, error) {
	path := filepath.Join(pmPath, "config", "project.yaml")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var p Project
	if err := yaml.Unmarshal(content, &p); err != nil {
		return nil, err
	}

	return &p, nil
}

func SaveProject(pmPath string, project *Project) error {
	path := filepath.Join(pmPath, "config", "project.yaml")
	content, err := yaml.Marshal(project)
	if err != nil {
		return err
	}

	return os.WriteFile(path, content, 0644)
}
