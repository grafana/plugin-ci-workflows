package workflow

import (
	"github.com/google/uuid"
)

// type mock func(*TestingWorkflow)

// TestingWorkflow represents a workflow that is meant to be temporary and used for testing purposes.
// It generates a unique filename for each instance.
// It embeds BaseWorkflow to inherit its the properties for marshaling
// a GitHub Actions Workflows in YAML format.
// It implements the Workflow interface.
type TestingWorkflow struct {
	// BaseWorkflow is the base definition of a generic GitHub Actions workflow.
	BaseWorkflow

	// uuid is a unique identifier for this testing workflow.
	// It's used to generate a unique file name.
	uuid uuid.UUID

	// baseWorkflowName is the name of the base workflow this testing workflow is derived from.
	// This is optional and is used to create more readable file names.
	baseWorkflowName string

	// children contains any other TestingWorkflows that need to be mocked as part of this test workflow
	// and should be created as temporary files alongside this workflow, thus acting as child workflows.
	children map[string]*TestingWorkflow
}

func (t *TestingWorkflow) FileName() string {
	return "act-" + t.baseWorkflowName + "-" + t.uuid.String() + ".yml"
}

func (t *TestingWorkflow) Children() []Workflow {
	children := make([]Workflow, 0, len(t.children))
	for _, child := range t.children {
		children = append(children, child)
	}
	return children
}

func (t *TestingWorkflow) GetChild(name string) *TestingWorkflow {
	return t.children[name]
}

func (t *TestingWorkflow) AddChild(name string, child *TestingWorkflow) {
	t.children[name] = child
}

func NewTestingWorkflow(baseName string, workflow BaseWorkflow, opts ...NewTestingWorkflowOption) *TestingWorkflow {
	wf := TestingWorkflow{
		uuid:             uuid.New(),
		baseWorkflowName: baseName,
		BaseWorkflow:     workflow,
		children:         map[string]*TestingWorkflow{},
	}
	for _, opt := range opts {
		opt(&wf)
	}
	return &wf
}

type NewTestingWorkflowOption func(*TestingWorkflow)

/* func WithChildWorkflow(name string, workflow TestingWorkflow) NewTestingWorkflowOption {
	return func(t *TestingWorkflow) {
		t.children[name] = workflow
	}
} */

func WithUUID(id uuid.UUID) NewTestingWorkflowOption {
	return func(t *TestingWorkflow) {
		t.uuid = id
	}
}

/* func WithMock(mock mock) NewTestingWorkflowOption {
	return func(t *TestingWorkflow) {
		mock(t)
	}
} */
