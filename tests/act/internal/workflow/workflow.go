package workflow

import "github.com/goccy/go-yaml"

type Permissions map[string]string
type Secrets map[string]string

type Workflow struct {
	Name        string
	On          On
	Permissions Permissions
	Jobs        map[string]Job
}

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

func (w Workflow) Marshal() ([]byte, error) {
	return yaml.Marshal(w)
}
