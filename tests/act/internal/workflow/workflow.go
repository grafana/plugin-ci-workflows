// Package workflow contains types and functions to define GitHub Actions workflows for testing with act.
// It provides a way to programmatically create workflows and jobs in a structured and type-safe manner.
package workflow

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

const (
	PCIWFBaseRef = "grafana/plugin-ci-workflows/.github/workflows"
)

// Workflow is an interface for workflows that can be marshaled to YAML format.
type Workflow interface {
	// FileName returns the file name for the workflow.
	FileName() string

	// Marshal converts the Workflow instance to its YAML representation.
	Marshal() ([]byte, error)

	// Children returns the child workflows of this workflow.
	Children() []*TestingWorkflow

	// Jobs returns the jobs defined in the workflow.
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

// NewBaseWorkflowFromFile creates a BaseWorkflow instance by reading and parsing a YAML file at the given path.
func NewBaseWorkflowFromFile(path string) (bw BaseWorkflow, err error) {
	f, err := os.Open(path)
	if err != nil {
		return BaseWorkflow{}, fmt.Errorf("open workflow file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	if err := yaml.NewDecoder(f).Decode(&bw); err != nil {
		return BaseWorkflow{}, fmt.Errorf("decode workflow file: %w", err)
	}
	return bw, nil
}

// Marshal converts the Workflow instance to its YAML representation.
func (w *BaseWorkflow) Marshal() ([]byte, error) {
	return yaml.Marshal(w)
}

// Permissions is the YAML representation of GitHub Actions job permissions.
type Permissions map[string]string

// Secrets is the YAML representation of GitHub Actions job secrets.
type Secrets map[string]string

// Steps is the YAML representation of a list of GitHub Actions steps.
type Steps []Step

// Job is the YAML representation of a GitHub Actions job.
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

// ReplaceStepAtIndex replaces (mocks) a step at the given index with the provided steps.
// It's similar to ReplaceStep, but uses the step index instead of the step id.
// This can be used for mocking steps in tests.
// The target step is replaced in place by the new steps.
// The original step's "If" condition is preserved and applied to all new steps.
// The original step's ID is preserved and applied to the first new step.
// If more than one step is provided, they will be injected at the same position as the original step
// in place of the original step.
// In this case, only the first new step will keep the original step's ID, and the others will have no ID.
// If the step index is out of range or no steps are provided, an error is returned.
func (j *Job) ReplaceStepAtIndex(stepIndex int, steps ...Step) error {
	if len(steps) == 0 {
		return errors.New("no steps provided to replace")
	}
	if stepIndex < 0 || stepIndex >= len(j.Steps) {
		return fmt.Errorf("step index %d out of range", stepIndex)
	}
	originalStep := j.Steps[stepIndex]

	// Preserve original step "If" condition if present
	if originalStep.If != "" {
		for i := range steps {
			steps[i].If = originalStep.If
		}
	}

	// Preserve the original step ID, but only for the first step.
	// If we replace multiple steps, only the first one should keep the original ID.
	steps[0].ID = originalStep.ID

	// Replace the step with the new steps, injecting them at the same position
	j.Steps = append(j.Steps[:stepIndex], append(steps, j.Steps[stepIndex+1:]...)...)
	return nil
}

// ReplaceStep replaces (mocks) a step with the given id with the provided steps.
// It's similar to ReplaceStepAtIndex, but looks up the step by its id.
// See the documentation of ReplaceStepAtIndex for more details.
func (j *Job) ReplaceStep(id string, steps ...Step) error {
	stepIndex := j.getStepIndex(id)
	if stepIndex == -1 {
		return fmt.Errorf("step with id %q not found", id)
	}
	return j.ReplaceStepAtIndex(stepIndex, steps...)
}

// RemoveStep removes a step at the given index from the job's steps.
// This is similar to RemoveStep, but uses the step index instead of the step id.
// This can be used for removing steps in tests, for example to skip certain actions
// that are not relevant to the test in order to speed up execution.
// Be careful when removing steps that are required by other steps (e.g.: steps that set outputs
// used by later steps), as this may cause the workflow to fail.
func (j *Job) RemoveStepAtIndex(stepIndex int) error {
	if stepIndex < 0 || stepIndex >= len(j.Steps) {
		return fmt.Errorf("step index %d out of range", stepIndex)
	}
	// Remove the step
	j.Steps = append(j.Steps[:stepIndex], j.Steps[stepIndex+1:]...)
	return nil
}

// RemoveStep is similar to RemoveStepAtIndex, but looks up the step by its id.
// See the documentation of RemoveStepAtIndex for more details.
func (j *Job) RemoveStep(id string) error {
	stepIndex := j.getStepIndex(id)
	if stepIndex == -1 {
		return fmt.Errorf("step with id %q not found", id)
	}
	// Remove the step
	return j.RemoveStepAtIndex(stepIndex)
}

// getStepIndex returns the index of the step with the given id.
// If the step is not found, -1 is returned.
func (j *Job) getStepIndex(id string) int {
	for i, step := range j.Steps {
		if step.ID == id {
			return i
		}
	}
	return -1
}

// GetStep retrieves a step with the given id from the job's steps.
// If the step is not found, nil is returned.
func (j *Job) GetStep(id string) *Step {
	for i, step := range j.Steps {
		if step.ID == id {
			return &j.Steps[i]
		}
	}
	return nil
}

// RemoveAllStepsAfter removes all steps after the step with the given id (exclusive).
// The step with the given id is preserved.
// If the step with the given id is not found, an error is returned.
func (j *Job) RemoveAllStepsAfter(id string) error {
	stepIndex := j.getStepIndex(id)
	if stepIndex == -1 {
		return fmt.Errorf("step with id %q not found", id)
	}
	j.Steps = j.Steps[:stepIndex+1]
	return nil
}

// ContainerJob is the YAML representation of a GitHub Actions job running in a container.
type ContainerJob struct {
	Image   string   `yaml:"image,omitempty"`
	Volumes []string `yaml:"volumes,omitempty"`
}

// Step is the YAML representation of a GitHub Actions step.
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

// On is the YAML representation of GitHub Actions workflow triggers.
type On struct {
	Push              OnPush              `yaml:"push,omitempty"`
	PullRequest       OnPullRequest       `yaml:"pull_request,omitempty"`
	PullRequestTarget OnPullRequestTarget `yaml:"pull_request_target,omitempty"`
	WorkflowCall      OnWorkflowCall      `yaml:"workflow_call,omitempty"`
}

// OnPush is the YAML representation of GitHub Actions push event trigger.
type OnPush struct {
	Branches []string `yaml:"branches,omitempty"`
}

// OnPullRequest is the YAML representation of GitHub Actions pull_request event trigger.
type OnPullRequest struct {
	Branches []string `yaml:"branches,omitempty"`
}

// OnPullRequestTarget is the YAML representation of GitHub Actions pull_request_target event trigger.
type OnPullRequestTarget struct {
	Branches []string `yaml:"branches,omitempty"`
}

// OnWorkflowCall is the YAML representation of GitHub Actions workflow_call event trigger.
type OnWorkflowCall struct {
	Inputs  map[string]WorkflowCallInput  `yaml:"inputs,omitempty"`
	Outputs map[string]WorkflowCallOutput `yaml:"outputs,omitempty"`
}

// WorkflowCallInput is the YAML representation of a GitHub Actions workflow call input field.
type WorkflowCallInput struct {
	Description string `yaml:"description,omitempty"`
	Type        string `yaml:"type,omitempty"`
	Required    bool   `yaml:"required,omitempty"`
	Default     any    `yaml:"default,omitempty"`
}

// WorkflowCallOutput is the YAML representation of a GitHub Actions workflow call output field.
type WorkflowCallOutput struct {
	Description string `yaml:"description,omitempty"`
	Value       string `yaml:"value,omitempty"`
}

// Commands is a utility type that represents a list of shell commands.
// It provides a String method to join the commands into a single string
// separated by newlines, suitable for use in a Step's Run field.
// This is handy for defining multi-line shell scripts without having to manually concat strings.
// Example:
//
// ```go
//
//	step := Step{
//		Name: "Example Step",
//		Run:  Commands{
//			"echo Hello, World!",
//			"ls -la"
//		}.String(),
//		Shell: "bash",
//	}
//
// ```
type Commands []string

// String joins the commands into a single string separated by newlines.
func (c Commands) String() string {
	return strings.Join(c, "\n")
}

// Input is a helper function to create a pointer to a value.
// It is useful to avoid having to write `&value` everywhere.
func Input[T any](value T) *T {
	return &value
}

// SetJobInput sets a job input value if it's not nil.
// It uses generics to avoid the interface nil gotcha where a typed nil pointer
// passed to an `any` parameter results in a non-nil interface value.
func SetJobInput[T any](job *Job, key string, value *T) {
	if value != nil {
		job.With[key] = *value
	}
}
