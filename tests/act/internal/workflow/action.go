package workflow

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

// CompositeAction represents the inputs and steps of a composite GitHub Action.
type CompositeAction struct {
	Inputs map[string]WorkflowCallInput `yaml:"inputs"`
	Runs   Job                          `yaml:"runs"`
}

// NewCompositeActionFromFile reads and parses a composite action definition.
func NewCompositeActionFromFile(path string) (CompositeAction, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return CompositeAction{}, fmt.Errorf("read composite action file: %w", err)
	}

	var action CompositeAction
	if err := yaml.Unmarshal(data, &action); err != nil {
		return CompositeAction{}, fmt.Errorf("decode composite action file: %w", err)
	}
	return action, nil
}
