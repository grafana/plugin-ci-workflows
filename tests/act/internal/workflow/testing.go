package workflow

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
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

// FileName returns the file name for the testing workflow.
func (t *TestingWorkflow) FileName() string {
	return "act-" + t.baseWorkflowName + "-" + t.uuid.String() + ".yml"
}

// Children returns the child workflows of this testing workflow.
func (t *TestingWorkflow) Children() []Workflow {
	children := make([]Workflow, 0, len(t.children))
	for _, child := range t.children {
		children = append(children, child)
	}
	return children
}

// Jobs returns the jobs defined in the testing workflow.
func (t *TestingWorkflow) Jobs() map[string]*Job {
	return t.BaseWorkflow.Jobs
}

// GetChild retrieves a child TestingWorkflow by its name.
func (t *TestingWorkflow) GetChild(name string) *TestingWorkflow {
	return t.children[name]
}

// AddChild adds a child TestingWorkflow to this testing workflow.
func (t *TestingWorkflow) AddChild(name string, child *TestingWorkflow) {
	t.children[name] = child
}

// UUID returns the unique identifier for this testing workflow.
func (t *TestingWorkflow) UUID() uuid.UUID {
	return t.uuid
}

type MockStepFactory func(originalStep Step) (Step, error)

func (t *TestingWorkflow) MockAllStepsUsingAction(actionPrefix string, mockStepFactory MockStepFactory) error {
	for _, job := range t.Jobs() {
		for i, step := range job.Steps {
			if strings.HasPrefix(step.Uses, actionPrefix) {
				mockedStep, err := mockStepFactory(step)
				if err != nil {
					return fmt.Errorf("mock step factory: %w", err)
				}
				if err := job.ReplaceStepAtIndex(i, mockedStep); err != nil {
					return fmt.Errorf("replace step: %w", err)
				}
			}
		}
	}
	return nil
}

// NewTestingWorkflow creates a new TestingWorkflow instance.
// It accepts a base workflow name, a BaseWorkflow instance, and optional configuration options.
func NewTestingWorkflow(baseName string, workflow BaseWorkflow, opts ...TestingWorkflowOption) *TestingWorkflow {
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
	for jid, j := range wf.Jobs() {
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

// TestingWorkflowOption defines a function type for configuring TestingWorkflow instances.
type TestingWorkflowOption func(*TestingWorkflow)

// WithUUID is a TestingWorkflowOption that sets a specific UUID for the TestingWorkflow.
// This is useful for linking child workflows to their parents in tests.
// If both have the same UUID, they will have predictable file names.
func WithUUID(id uuid.UUID) TestingWorkflowOption {
	return func(t *TestingWorkflow) {
		t.uuid = id
	}
}

// WithPullRequestTargetTrigger is a TestingWorkflowOption that sets a pull_request_target trigger to the workflow.
// This can be used to test workflows that respond to pull_request_target events.
func WithPullRequestTargetTrigger(branches []string) TestingWorkflowOption {
	return func(t *TestingWorkflow) {
		t.On.PullRequestTarget = OnPullRequestTarget{
			Branches: branches,
		}
	}
}

// WithRemoveAllStepsAfter removes all steps after the given step ID
// in the given job ID in the workflow.
// This can be used to stop the workflow at a certain point for testing purposes.
func WithRemoveAllStepsAfter(t *testing.T, jobID, stepID string) TestingWorkflowOption {
	return func(twf *TestingWorkflow) {
		job, ok := twf.BaseWorkflow.Jobs[jobID]
		require.True(t, ok, fmt.Errorf("job %q not found", jobID))
		err := job.RemoveAllStepsAfter(stepID)
		require.NoError(t, err, "remove all steps after %q in job %q", stepID, jobID)
	}
}

// WithOnlyOneJob keeps only the given job ID and its dependencies
// in the workflow, removing all other jobs.
// This can be used to run only a specific job for testing purposes.
func WithOnlyOneJob(t *testing.T, jobID string) TestingWorkflowOption {
	return func(twf *TestingWorkflow) {
		onlyJob, ok := twf.BaseWorkflow.Jobs[jobID]
		require.True(t, ok, fmt.Errorf("job %q not found", jobID))

		// Remove all jobs except the given one and its dependencies
		for k := range twf.BaseWorkflow.Jobs {
			if k == jobID || slices.Contains(onlyJob.Needs, k) {
				continue
			}
			delete(twf.BaseWorkflow.Jobs, k)
		}
	}
}

// WithNoOpStep modifies the TestingWorkflow to replace the step with the given ID
// in the job with the given name with a no-op step.
// This can be used to skip steps that are not relevant for the test or that would fail otherwise.
func WithNoOpStep(t *testing.T, jobID, stepID string) TestingWorkflowOption {
	return func(twf *TestingWorkflow) {
		err := twf.BaseWorkflow.Jobs[jobID].ReplaceStep(stepID, NoOpStep(stepID))
		require.NoError(t, err)
	}
}

// WithMockedGCS modifies the workflow to mock all GCS upload steps
// (which use the google-github-actions/upload-cloud-storage action)
// to instead copy files to a local folder mounted into the act container at /gcs.
// It also takes all google-github-actions/auth steps and no-ops them,
// as authentication is not needed for local file copy.
// This allows testing GCS upload functionality without actually accessing GCS.
// Since GCS is only used in trusted contexts, callers should most likely also use WithMockedWorkflowContext.
func WithMockedGCS(t *testing.T) TestingWorkflowOption {
	return func(twf *TestingWorkflow) {
		require.NoError(t, twf.MockAllStepsUsingAction(GCSLoginAction, func(step Step) (Step, error) {
			return NoOpStep(step.ID), nil
		}))
		require.NoError(t, twf.MockAllStepsUsingAction(GCSUploadAction, func(step Step) (Step, error) {
			return MockGCSUploadStep(step)
		}))
	}
}
