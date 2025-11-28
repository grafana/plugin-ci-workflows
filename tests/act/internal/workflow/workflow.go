// Package workflow contains types and functions to define GitHub Actions workflows for testing with act.
// It provides a way to programmatically create workflows and jobs in a structured and type-safe manner.
package workflow

import (
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

// Workflow is an interface for workflows that can be marshaled to YAML format.
type Workflow interface {
	// FileName returns the file name for the workflow.
	FileName() string

	// Marshal converts the Workflow instance to its YAML representation.
	Marshal() ([]byte, error)

	Children() []Workflow
	Jobs() map[string]*Job
}

// BaseWorkflow represents a GitHub Actions workflow definition.
type BaseWorkflow struct {
	Name        string
	On          On
	Permissions Permissions       `yaml:"permissions,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	Jobs        map[string]*Job
}

func NewBaseWorkflowFromFile(path string) (BaseWorkflow, error) {
	f, err := os.Open(path)
	if err != nil {
		return BaseWorkflow{}, fmt.Errorf("open workflow file: %w", err)
	}
	defer f.Close()
	var bw BaseWorkflow
	if err := yaml.NewDecoder(f).Decode(&bw); err != nil {
		return BaseWorkflow{}, fmt.Errorf("decode workflow file: %w", err)
	}
	return bw, nil
}

// Marshal converts the Workflow instance to its YAML representation.
func (w *BaseWorkflow) Marshal() ([]byte, error) {
	return yaml.Marshal(w)
}

type Permissions map[string]string
type Secrets map[string]string

type Steps []Step

type Job struct {
	Name string `yaml:"name,omitempty"`

	If string `yaml:"if,omitempty"`

	RunsOn  string            `yaml:"runs-on,omitempty"`
	Needs   []string          `yaml:"needs,omitempty"`
	Outputs map[string]string `yaml:"outputs,omitempty"`

	Permissions Permissions `yaml:"permissions,omitempty"`

	Uses string         `yaml:"uses,omitempty"`
	With map[string]any `yaml:"with,omitempty"`

	Secrets Secrets `yaml:"secrets,omitempty"`

	Steps Steps `yaml:"steps,omitempty"`

	Container ContainerJob `yaml:"container,omitempty"`
}

func (j *Job) ReplaceStep(id string, steps ...Step) error {
	stepIndex := j.GetStepIndex(id)
	if stepIndex == -1 {
		return fmt.Errorf("step with id %q not found", id)
	}
	// Replace the step with the new steps, injecting them at the same position
	j.Steps = append(j.Steps[:stepIndex], append(steps, j.Steps[stepIndex+1:]...)...)
	return nil
}

func (j *Job) RemoveStep(id string) error {
	stepIndex := j.GetStepIndex(id)
	if stepIndex == -1 {
		return fmt.Errorf("step with id %q not found", id)
	}
	// Remove the step
	j.Steps = append(j.Steps[:stepIndex], j.Steps[stepIndex+1:]...)
	return nil
}

func (j *Job) NoOpStep(id string) error {
	stepIndex := j.GetStepIndex(id)
	if stepIndex == -1 {
		return fmt.Errorf("step with id %q not found", id)
	}
	j.Steps[stepIndex] = Step{
		Name:  j.Steps[stepIndex].Name,
		ID:    j.Steps[stepIndex].ID,
		Run:   "echo 'noop-ed step for testing'",
		Shell: "bash",
	}
	return nil
}

func (j *Job) GetStepIndex(id string) int {
	for i, step := range j.Steps {
		if step.ID == id {
			return i
		}
	}
	return -1
}

func (j *Job) GetStep(id string) *Step {
	for i, step := range j.Steps {
		if step.ID == id {
			return &j.Steps[i]
		}
	}
	return nil
}

type ContainerJob struct {
	Image   string   `yaml:"image,omitempty"`
	Volumes []string `yaml:"volumes,omitempty"`
}

type Step struct {
	Name string `yaml:"name,omitempty"`
	ID   string `yaml:"id,omitempty"`

	If string `yaml:"if,omitempty"`

	Uses string         `yaml:"uses,omitempty"`
	With map[string]any `yaml:"with,omitempty"`

	Run              string `yaml:"run,omitempty"`
	Shell            string `yaml:"shell,omitempty"`
	WorkingDirectory string `yaml:"working-directory,omitempty"`

	Env map[string]string `yaml:"env,omitempty"`
}

type On struct {
	Push         OnPush         `yaml:"push,omitempty"`
	PullRequest  OnPullRequest  `yaml:"pull_request,omitempty"`
	WorkflowCall OnWorkflowCall `yaml:"workflow_call,omitempty"`
}

type OnPush struct {
	Branches []string `yaml:"branches,omitempty"`
}

type OnPullRequest struct {
	Branches []string `yaml:"branches,omitempty"`
}

type OnWorkflowCall struct {
	Inputs  map[string]WorkflowCallInput  `yaml:"inputs,omitempty"`
	Outputs map[string]WorkflowCallOutput `yaml:"outputs,omitempty"`
}

type WorkflowCallInput struct {
	Description string `yaml:"description,omitempty"`
	Type        string `yaml:"type,omitempty"`
	Required    bool   `yaml:"required,omitempty"`
	Default     any    `yaml:"default,omitempty"`
}

type WorkflowCallOutput struct {
	Description string `yaml:"description,omitempty"`
	Value       string `yaml:"value,omitempty"`
}

type Commands []string

func (c Commands) String() string {
	return strings.Join(c, "\n")
}
