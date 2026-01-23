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

// Children returns the child workflows of this testing workflow as a slice.
func (t *TestingWorkflow) Children() []*TestingWorkflow {
	children := make([]*TestingWorkflow, 0, len(t.children))
	for _, child := range t.children {
		children = append(children, child)
	}
	return children
}

// ChildrenRecursive returns all child workflows of this testing workflow, recursively, as a slice.
func (t *TestingWorkflow) ChildrenRecursive() []*TestingWorkflow {
	children := make([]*TestingWorkflow, 0, len(t.children))
	for _, child := range t.children {
		children = append(children, child)
		children = append(children, child.ChildrenRecursive()...)
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

// AddUUIDToAllJobsRecursive adds a UUID to each job in the workflow and all its children (recursively) in order to
// make unique container names and allow tests to run in parallel, so that
// container names created by act don't clash
func (t *TestingWorkflow) AddUUIDToAllJobsRecursive() {
	uid := t.UUID().String()
	allWorkflows := t.ChildrenRecursive()
	// Add the main workflow as well
	allWorkflows = append(allWorkflows, t)
	// Add UUID to all jobs to avoid container name clashes
	for _, wf := range allWorkflows {
		for _, j := range wf.Jobs() {
			if j.Name != "" {
				j.Name += "-"
			}
			j.Name += uid
		}
	}
}

// MockStepFactory is a function that creates a mocked step from an original step.
type MockStepFactory func(originalStep Step) (Step, error)

// MockAllStepsUsingAction modifies all steps in the workflow that use the given action prefix
// by replacing them with the mocked step created by the given mockStepFactory function.
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

// WithOnlyOneJob keeps only the given job ID, removing all other jobs.
// If removeDependencies is true, it will also remove all dependencies of the given job.
// You normally don't want to remove dependencies, otherwise the workflow might fail if it consumes the output of a dependency.
func WithOnlyOneJob(t *testing.T, jobID string, removeDependencies bool) TestingWorkflowOption {
	return func(twf *TestingWorkflow) {
		onlyJob, ok := twf.BaseWorkflow.Jobs[jobID]
		require.True(t, ok, fmt.Errorf("job %q not found", jobID))

		// Remove all jobs
		for k := range twf.BaseWorkflow.Jobs {
			// Do not remove the given job if it's a dependency and we don't want to remove dependencies
			if k == jobID || (!removeDependencies && isDependency(twf, onlyJob, k)) {
				continue
			}
			delete(twf.BaseWorkflow.Jobs, k)
		}
		// Remove all dependencies from the only job left in the workflow, otherwise it won't run
		// because it depends on a job that has been removed.
		if removeDependencies {
			onlyJob.Needs = nil
		}
	}
}

// isDependency checks if the given jobID is a dependency (direct or indirect) of the given job.
func isDependency(wf *TestingWorkflow, job *Job, jobID string) bool {
	if slices.Contains(job.Needs, jobID) {
		// Direct dependency
		return true
	}
	// Recursively check for indirect dependencies
	for _, need := range job.Needs {
		if isDependency(wf, wf.Jobs()[need], jobID) {
			return true
		}
	}
	return false
}

// WithoutJob removes the given job from the workflow.
func WithoutJob(jobID string) TestingWorkflowOption {
	return func(twf *TestingWorkflow) {
		delete(twf.BaseWorkflow.Jobs, jobID)
	}
}

// WithNoOpJobWithOutputs modifies the TestingWorkflow to replace the given job with a no-op job that sets the given outputs.
// This can be used to skip jobs that are not relevant for the test or that would fail otherwise.
// This is useful combined with WithOnlyOneJob to remove all jobs except the given one.
// Then, WithNoOpJobWithOutputs can be used on the remaining job's dependencies to set mock outputs rather than running the actual job.
// This way the remaining job can run (needs.X.outputs.Y will be available) and the test can run.
func WithNoOpJobWithOutputs(t *testing.T, jobID string, outputs map[string]string) TestingWorkflowOption {
	return func(twf *TestingWorkflow) {
		job, ok := twf.BaseWorkflow.Jobs[jobID]
		require.True(t, ok, fmt.Errorf("job %q not found", jobID))

		// Reset uses and nil in case the job uses an action rather than steps
		job.Uses = ""
		job.With = nil

		// Enforce runs-on/empty strategy otherwise the job can't run the steps
		job.RunsOn = "ubuntu-arm64-small"
		job.Strategy = Strategy{}

		// Set the steps to a no-op step that sets the outputs
		mockStep := MockOutputsStep(outputs)
		mockStep.ID = "set-outputs"
		job.Outputs = make(map[string]string, len(outputs))
		for k := range outputs {
			job.Outputs[k] = fmt.Sprintf("${{ steps.%s.outputs.%s }}", mockStep.ID, k)
		}
		job.Steps = Steps{mockStep}
	}
}

// WithReplacedStep replaces the step with the given ID in the given job with the given step.
// This can be used to replace a step with a mocked step for testing purposes.
func WithReplacedStep(t *testing.T, jobID string, stepID string, step Step) TestingWorkflowOption {
	return func(twf *TestingWorkflow) {
		job, ok := twf.BaseWorkflow.Jobs[jobID]
		require.True(t, ok, fmt.Errorf("job %q not found", jobID))
		err := job.ReplaceStep(stepID, step)
		require.NoError(t, err, "replace step %q in job %q", stepID, jobID)
	}
}

// WithNoOpStep modifies the TestingWorkflow to replace the step with the given ID
// in the job with the given name with a no-op step.
// This can be used to skip steps that are not relevant for the test or that would fail otherwise.
func WithNoOpStep(t *testing.T, jobID, stepID string) TestingWorkflowOption {
	return func(twf *TestingWorkflow) {
		step := twf.BaseWorkflow.Jobs[jobID].GetStep(stepID)
		require.NotNilf(t, step, "step with id %q not found in job %q", stepID, jobID)
		err := twf.BaseWorkflow.Jobs[jobID].ReplaceStep(stepID, NoOpStep(*step))
		require.NoError(t, err)
	}
}

// WithMatrix sets the matrix for the given job.
// This can be used to test workflows that use a dynamic matrix.
// act doesn't support dynamic matrix values, so this is a workaround to set the matrix for the given job.
//
// For example, if the matrix is:
// ```
//
//	matrix:
//	  environment: ${{ fromJson(needs.setup.outputs.environments) }}
//
// ```
//
// It will not be expanded correctly at runtime (it will be empty), so we need to set the matrix manually.
// More information: https://github.com/go-gitea/gitea/issues/25179
// TODO: remove this once act supports dynamic matrix values.
func WithMatrix(job string, matrix map[string][]string) TestingWorkflowOption {
	return func(twf *TestingWorkflow) {
		matrixMap := make(map[string]any)
		for k, v := range matrix {
			matrixMap[k] = v
		}
		twf.BaseWorkflow.Jobs[job].Strategy.Matrix = matrixMap
	}
}

// WithEnvironment sets the environment variables for the given step in the given job.
// This can be used to set environment variables for testing purposes.
// This is also handy for passing mocked data to be used when running act tests, if the step supports it.
func WithEnvironment(t *testing.T, job string, step string, environment map[string]string) TestingWorkflowOption {
	return func(twf *TestingWorkflow) {
		step := twf.BaseWorkflow.Jobs[job].GetStep(step)
		require.NotNilf(t, step, "step %q not found in job %q", step, job)
		for k, v := range environment {
			step.Env[k] = v
		}
	}
}

// WithPullRequestTrigger is a TestingWorkflowOption that sets a pull_request trigger to the workflow.
// This can be used to test workflows that respond to pull_request events.
func WithPullRequestTrigger(branches []string) TestingWorkflowOption {
	return func(t *TestingWorkflow) {
		t.On = On{
			PullRequest: OnPullRequest{
				Branches: branches,
			},
		}
	}
}

// WithPullRequestTargetTrigger is a TestingWorkflowOption that sets a pull_request_target trigger to the workflow.
// This can be used to test workflows that respond to pull_request_target events.
func WithPullRequestTargetTrigger(branches []string) TestingWorkflowOption {
	return func(t *TestingWorkflow) {
		t.On = On{
			PullRequestTarget: OnPullRequestTarget{
				Branches: branches,
			},
		}
	}
}

// WithWorkflowDispatchTrigger is a TestingWorkflowOption that sets a workflow_dispatch trigger to the workflow.
// This can be used to test workflows that are manually triggered via the GitHub UI or API.
func WithWorkflowDispatchTrigger(inputs map[string]WorkflowCallInput) TestingWorkflowOption {
	return func(t *TestingWorkflow) {
		t.On = On{
			WorkflowDispatch: OnWorkflowDispatch{
				Inputs: inputs,
			},
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
			return NoOpStep(step), nil
		}))
		require.NoError(t, twf.MockAllStepsUsingAction(GCSUploadAction, func(step Step) (Step, error) {
			return MockGCSUploadStep(step)
		}))
	}
}

// WithMockedVault modifies the SimpleCD workflow to mock all Vault secrets steps
// (which use the grafana/shared-workflows/actions/get-vault-secrets action)
// to instead return the provided mock secrets.
// This allows testing CD workflows without actually accessing Vault.
//
// The secrets map should contain the secret names as keys and their mock values.
// For example:
//
//	secrets := VaultSecrets{
//	    "GCOM_PUBLISH_TOKEN_DEV": "mock-dev-token",
//	    "GCOM_PUBLISH_TOKEN_OPS": "mock-ops-token",
//	    "GCOM_PUBLISH_TOKEN_PROD": "mock-prod-token",
//	}
func WithMockedVault(t *testing.T, secrets VaultSecrets) TestingWorkflowOption {
	return func(twf *TestingWorkflow) {
		err := twf.MockAllStepsUsingAction(VaultSecretsAction, func(step Step) (Step, error) {
			return MockVaultSecretsStep(step, secrets)
		})
		require.NoError(t, err)
	}
}

// WithMockedGitHubAppToken modifies the workflow to mock the GitHub app token creation step
// (which use the actions/create-github-app-token action)
// to instead return the provided mock token.
// This allows testing GitHub app token creation functionality without actually creating a GitHub app.
// Since GitHub app tokens are only used in trusted contexts, callers should most likely also use WithMockedWorkflowContext.
func WithMockedGitHubAppToken(t *testing.T, token ...string) TestingWorkflowOption {
	if len(token) == 0 {
		token = []string{"MOCK_GITHUB_APP_TOKEN"}
	}
	return func(twf *TestingWorkflow) {
		err := twf.MockAllStepsUsingAction(GitHubAppTokenAction, func(originalStep Step) (Step, error) {
			return MockGitHubAppTokenStep(originalStep, token[0])
		})
		require.NoError(t, err)
	}
}
