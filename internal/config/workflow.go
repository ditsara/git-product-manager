package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Workflow struct {
	States       []string          `yaml:"states"`
	InitialState string            `yaml:"initial_state"`
	StateGroups  map[string][]string `yaml:"state_groups"`
}

func LoadWorkflow(path string) (*Workflow, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var w Workflow
	if err := yaml.Unmarshal(content, &w); err != nil {
		return nil, err
	}

	return &w, nil
}

func (w *Workflow) IsValidState(state string) bool {
	for _, s := range w.States {
		if s == state {
			return true
		}
	}
	return false
}
