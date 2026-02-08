package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Workflow struct {
	States       []string            `yaml:"states"`
	InitialState string              `yaml:"initial_state"`
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

// GetCompletedStates returns the list of states in the "completed" state group
// Returns nil if no completed group is defined
func (w *Workflow) GetCompletedStates() []string {
	if completedStates, ok := w.StateGroups["completed"]; ok {
		return completedStates
	}
	return nil
}

// IsCompleted checks if a given status is in the "completed" state group
func (w *Workflow) IsCompleted(status string) bool {
	completedStates := w.GetCompletedStates()
	for _, state := range completedStates {
		if state == status {
			return true
		}
	}
	return false
}

// GetStateGroup returns the states in a named state group
// Returns nil if the group doesn't exist
func (w *Workflow) GetStateGroup(groupName string) []string {
	if states, ok := w.StateGroups[groupName]; ok {
		return states
	}
	return nil
}
