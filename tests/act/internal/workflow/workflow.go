// Package workflow contains types and functions to define GitHub Actions workflows for testing with act.
// It provides a way to programmatically create workflows and jobs in a structured and type-safe manner.
package workflow

import "github.com/goccy/go-yaml"

// Marshalable is an interface for workflows that can be marshaled to YAML format.
type Marshalable interface {
	Marshal() ([]byte, error)
}

// Workflow represents a GitHub Actions workflow definition.
// It implements the Marshalable interface to allow conversion to YAML format.
type Workflow struct {
	Name        string
	On          On
	Permissions Permissions
	Jobs        map[string]Job
}

// Marshal converts the Workflow instance to its YAML representation.
func (w Workflow) Marshal() ([]byte, error) {
	return yaml.Marshal(w)
}

// FirstJob returns a pointer to the first job in the workflow.
// If there are no jobs, it returns empty string and nil.
func (w Workflow) FirstJob() *Job {
	for _, job := range w.Jobs {
		return &job
	}
	return nil
}

type Permissions map[string]string
type Secrets map[string]string

type Job struct {
	Name        string
	Uses        string
	Permissions Permissions
	With        map[string]any
	Secrets     Secrets
}

type On struct {
	Push        OnPush        `yaml:"push,omitempty"`
	PullRequest OnPullRequest `yaml:"pull_request,omitempty"`
}

type OnPush struct {
	Branches []string `yaml:"branches,omitempty"`
}

type OnPullRequest struct {
	Branches []string `yaml:"branches,omitempty"`
}
