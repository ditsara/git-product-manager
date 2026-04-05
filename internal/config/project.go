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
