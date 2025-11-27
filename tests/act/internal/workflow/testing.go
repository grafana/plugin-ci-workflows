package workflow

import (
	"github.com/google/uuid"
)

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

	// Add a job to get the workflow run ID and output it.
	// This is useful for retrieving artifacts by the workflow run ID.
	const getWorkflowRunIDJobName = "get-workflow-run-id"
	workflow.Jobs[getWorkflowRunIDJobName] = &Job{
		Name:   "Get workflow run ID",
		RunsOn: "ubuntu-arm64-small",
		Steps: Steps{
			{
				Name:  "Get workflow run ID",
				ID:    "run-id",
				Run:   `echo run-id=${{ github.run_id }} >> $GITHUB_OUTPUT`,
				Shell: "bash",
			},
		},
	}

	// Make sure all other jobs depend on the get-workflow-run-id job, so it runs first
	for jid, j := range wf.Jobs {
		if jid == getWorkflowRunIDJobName {
			continue
		}
		j.Needs = append(j.Needs, getWorkflowRunIDJobName)
	}

	// Apply options to customize the TestingWorkflow instance.
	for _, opt := range opts {
		opt(&wf)
	}
	return &wf
}

type NewTestingWorkflowOption func(*TestingWorkflow)

func withUUID(id uuid.UUID) NewTestingWorkflowOption {
	return func(t *TestingWorkflow) {
		t.uuid = id
	}
}
