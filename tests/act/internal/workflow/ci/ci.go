package ci

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/stretchr/testify/require"
)

// Workflow is a predefined GitHub Actions workflow for testing plugins using act.
// It uses the plugin-ci-workflows CI workflow as a base, with sane default values
// and allows customization through options.
// It implements the Marshalable interface to allow conversion to YAML format.
// Instances must be created using NewWorkflow.
type Workflow struct {
	*workflow.TestingWorkflow
}

// NewWorkflow creates a new Workflow instance with default settings.
// The caller can provide options to customize the workflow.
func NewWorkflow(opts ...WorkflowOption) (Workflow, error) {
	ciBaseWf := workflow.BaseWorkflow{
		Name: "CI",
		On: workflow.On{
			Push: workflow.OnPush{
				Branches: []string{"main"},
			},
			PullRequest: workflow.OnPullRequest{
				Branches: []string{"main"},
			},
		},
		Jobs: map[string]*workflow.Job{
			"ci": {
				Name: "CI",
				// This will be populated later with the child testing workflow reference
				// Uses: "..."
				Permissions: workflow.Permissions{
					"contents": "read",
					"id-token": "write",
				},
				With: map[string]any{
					"plugin-version-suffix": "${{ github.event_name == 'pull_request' && github.event.pull_request.head.sha || '' }}",
					"testing":               true,
				},
				Secrets: workflow.Secrets{
					"GITHUB_TOKEN": "${{ secrets.GITHUB_TOKEN }}",
				},
			},
		},
	}

	// We need to create a child testing workflow for the called "ci.yml" workflow, in order to mock jobs/steps in it
	// Read the base workflow from file to create the child BaseWorkflow
	childBaseWf, err := workflow.NewBaseWorkflowFromFile(filepath.Join(".github", "workflows", "ci.yml"))
	if err != nil {
		return Workflow{}, fmt.Errorf("new base workflow from file for child ci workflow: %w", err)
	}

	// Workflow to be returned
	testingWf := Workflow{workflow.NewTestingWorkflow("simple-ci", ciBaseWf)}

	// Add the child workflow (ci) now, so further customization can be done on it via opts.
	// Use the same UUID as the parent, so they have the same uuid in the file name
	// and it is easier to correlate them.
	childTestingWf := workflow.NewTestingWorkflow("ci", childBaseWf, workflow.WithUUID(testingWf.UUID()))
	testingWf.AddChild("ci", childTestingWf)

	// Change the parent workflow so it calls the mocked child workflow
	testingWf.BaseWorkflow.Jobs["ci"].Uses = workflow.PCIWFBaseRef + "/" + testingWf.GetChild("ci").FileName() + "@main"

	// Add uuid to each job in the workflow and all its children in order to
	// make unique contianer names and allow tests to run in parallel, so that
	// container names created by act don't clash
	// TODO: move to TestingWorkflow instead?
	for _, wf := range append([]workflow.Workflow{testingWf.TestingWorkflow}, testingWf.Children()...) {
		for _, j := range wf.Jobs() {
			if j.Name != "" {
				j.Name = j.Name + "-" + testingWf.UUID().String()
			} else {
				j.Name = testingWf.UUID().String()
			}
		}
	}

	// Apply options to customize the Workflow instance.
	// These opts can also modify the child testing workflow.
	for _, opt := range opts {
		opt(&testingWf)
	}
	return testingWf, nil
}

// CIWorkflow returns the TestingWorkflow instance representing the "ci" child workflow.
// This can be used to further customize/mock steps and jobs in the child workflow.
func (w *Workflow) CIWorkflow() *workflow.TestingWorkflow {
	return w.GetChild("ci")
}

// WorkflowOption is a function that modifies a Workflow instance during its construction.
type WorkflowOption func(*Workflow)

// WithPluginDirectoryInput sets the plugin-directory input for the CI job in the workflow.
func WithPluginDirectoryInput(dir string) WorkflowOption {
	return func(w *Workflow) {
		w.BaseWorkflow.Jobs["ci"].With["plugin-directory"] = dir
	}
}

// WithDistArtifactPrefixInput sets the dist-artifacts-prefix input for the CI job in the workflow.
func WithDistArtifactPrefixInput(prefix string) WorkflowOption {
	return func(w *Workflow) {
		w.BaseWorkflow.Jobs["ci"].With["dist-artifacts-prefix"] = prefix
	}
}

// WithPlaywrightInput sets the run-playwright input for the CI job in the workflow.
func WithPlaywrightInput(enabled bool) WorkflowOption {
	return func(w *Workflow) {
		w.BaseWorkflow.Jobs["ci"].With["run-playwright"] = enabled
	}
}

// WithRunPluginValidatorInput sets the run-plugin-validator input for the CI job in the workflow.
func WithRunPluginValidatorInput(enabled bool) WorkflowOption {
	return func(w *Workflow) {
		w.BaseWorkflow.Jobs["ci"].With["run-plugin-validator"] = enabled
	}
}

// WithPluginValidatorConfigInput sets the plugin-validator-config input for the CI job in the workflow.
func WithPluginValidatorConfigInput(config string) WorkflowOption {
	return func(w *Workflow) {
		w.BaseWorkflow.Jobs["ci"].With["plugin-validator-config"] = config
	}
}

// WithRunTruffleHogInput sets the run-trufflehog input for the CI job in the workflow.
func WithRunTruffleHogInput(enabled bool) WorkflowOption {
	return func(w *Workflow) {
		w.BaseWorkflow.Jobs["ci"].With["run-trufflehog"] = enabled
	}
}

// WithAllowUnsignedInput sets the allow-unsigned input for the CI job in the workflow.
func WithAllowUnsignedInput(enabled bool) WorkflowOption {
	return func(w *Workflow) {
		w.BaseWorkflow.Jobs["ci"].With["allow-unsigned"] = enabled
	}
}

// WithTestingInput sets the testing input for the CI job in the workflow.
func WithTestingInput(testing bool) WorkflowOption {
	return func(w *Workflow) {
		w.BaseWorkflow.Jobs["ci"].With["testing"] = testing
	}
}

// WithMockedDist modifies the workflow to mock the test-and-build job
// to copy pre-built dist files (js + assets + backend executable, NOT the ZIP files)
// from a mockdata folder instead of building them.
// This can be used for tests that need to assert on side-effects of building the plugin,
// without actually building it, which saves execution time.
// The distFolder is relative to tests/act/mockdata (e.g.: `dist/simple-frontend`).
// The distFolder should use slashes as path separators.
// The function will convert it to the correct OS-specific separators when needed.
// The distFolder is sanity-checked to ensure they contain valid data.
func WithMockedDist(t *testing.T, distFolder string) WorkflowOption {
	return func(w *Workflow) {
		testAndBuild := w.CIWorkflow().BaseWorkflow.Jobs["test-and-build"]
		distFolder = filepath.FromSlash(distFolder)

		// Sanity check that the folder contains dist files
		_, err := os.Stat(filepath.Join(workflow.LocalMockdataPath(distFolder), "plugin.json"))
		if err != nil && os.IsNotExist(err) {
			require.FailNowf(t, "malformed dist folder", "the specified dist folder %q doesn't seem to contain dist artifacts (plugin.json is missing)", distFolder)
		}

		require.NoError(t, testAndBuild.ReplaceStep(
			"frontend",
			workflow.CopyMockFilesStep(distFolder, "${{ github.workspace }}/${{ inputs.plugin-directory }}/dist/"),
		))
		require.NoError(t, testAndBuild.RemoveStep("backend"))
	}
}

// WithMockedPackagedDistArtifacts modifies the workflow to mock the steps that create
// the packaged dist artifacts (ZIP files) in the test-and-build job to copy pre-packaged ZIP files.
// It also modifies the workflow to mock the dist files using WithMockedDist.
// This way if any further steps need the dist files (e.g. for extracting metadata from plugin.json), they are present.
// The distFolder parameter is the name of the plugin folder inside `mockdata` that contains the dist files (js + assets, etc), not the ZIP file.
// The packagedFolder parameter is the name of the folder inside `mockdata` that contains the pre-packaged ZIP files.
// Both folders are relative to tests/act/mockdata (e.g.: `dist/simple-frontend` and `dist-artifacts-unsigned/simple-frontend`).
// Both folders should use slashes as path separators.
// The function will convert them to the correct OS-specific separators when needed.
// The specified mock folders are sanity-checked to ensure they contain valid data.
func WithMockedPackagedDistArtifacts(t *testing.T, distFolder string, packagedFolder string) WorkflowOption {
	return func(w *Workflow) {
		// Sanity check that the packaged folder contains ZIP files
		packagedFolder = filepath.FromSlash(packagedFolder)
		entries, err := os.ReadDir(workflow.LocalMockdataPath(packagedFolder))
		if err != nil {
			require.FailNowf(t, "malformed packaged dist folder", "could not read the specified packaged dist folder %q", packagedFolder)
		}
		hasZip := false
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".zip") {
				hasZip = true
				break
			}
		}
		if !hasZip {
			require.FailNowf(t, "the specified packaged dist folder %q doesn't seem to contain any ZIP files", packagedFolder)
		}

		// Mock dist files as well (unpackaged plugin files), so if any steps require the dist files, they are present
		WithMockedDist(t, distFolder)(w)

		testAndBuild := w.CIWorkflow().BaseWorkflow.Jobs["test-and-build"]
		// Remove unnecessary steps (those that build the plugin)
		for _, id := range []string{
			"setup",
			"replace-plugin-version",
		} {
			require.NoError(t, testAndBuild.RemoveStep(id))
		}

		// Mock package steps
		dest := "${{ github.workspace }}/${{ inputs.plugin-directory }}/dist-artifacts/"
		for i, id := range []string{
			"universal-zip",
			"os-arch-zips",
		} {
			mockStep := workflow.CopyMockFilesStep(packagedFolder, dest)
			// Set step output
			if i == 0 {
				// Universal
				mockStep.Run += "\n" + workflow.Commands{
					// Output ONE zip file, get the name by excluding file names that contain '_'
					// (which is used as a separator in os/arch zips)
					`echo zip=$(ls -1 ` + dest + `/*.zip | xargs -n 1 basename | grep -v '_') >> "${GITHUB_OUTPUT}"`,
				}.String()
			} else {
				// os/arch
				mockStep.Run += "\n" + workflow.Commands{
					// Output ALL ZIP files that contains an '_' (separator for os/arch in zip file names)
					// as a JSON array
					`echo zip=$(ls -1 ` + dest + `/*.zip | xargs -n 1 basename | grep '_' | jq -RncM '[inputs]') >> "${GITHUB_OUTPUT}"`,
				}.String()
			}
			require.NoError(t, testAndBuild.ReplaceStep(id, mockStep))
		}
	}
}

// workflowMutator is a helper to mutate the Workflow or its children workflows
// with options that are not specific to the Workflow itself, but rather to the testing workflow in general.
type workflowMutator struct {
	workflowGetter func(*Workflow) *workflow.TestingWorkflow
}

// MutateTestingWorkflow returns a workflowMutator that can be used to mutate the testing workflow.
func MutateTestingWorkflow() workflowMutator {
	return workflowMutator{
		workflowGetter: func(w *Workflow) *workflow.TestingWorkflow {
			return w.TestingWorkflow
		},
	}
}

// MutateCIWorkflow returns a workflowMutator that can be used to mutate the CI workflow
// (child of the testing workflow).
func MutateCIWorkflow() workflowMutator {
	return workflowMutator{
		workflowGetter: func(w *Workflow) *workflow.TestingWorkflow {
			return w.CIWorkflow()
		},
	}
}

// With applies the given options to the workflow returned by the workflowGetter function.
func (m workflowMutator) With(opts ...workflow.TestingWorkflowOption) WorkflowOption {
	return func(w *Workflow) {
		wf := m.workflowGetter(w)
		for _, opt := range opts {
			opt(wf)
		}
	}
}

// Context represents the mocked workflow context.
// It is the JSON payload returned by the "workflow-context" step.
type Context struct {
	IsTrusted bool `json:"isTrusted"`
	IsForkPR  bool `json:"isForkPR"`
}

// WithMockedWorkflowContext modifies the workflow to mock the "workflow-context" step
// to return the given mocked Context.
// This can be used to test behavior that depends on whether the workflow is running in a trusted context or not.
func WithMockedWorkflowContext(t *testing.T, ctx Context) WorkflowOption {
	return func(w *Workflow) {
		step, err := mockWorkflowContextStep(ctx)
		require.NoError(t, err)

		const stepID = "workflow-context"
		err = w.CIWorkflow().BaseWorkflow.Jobs["test-and-build"].ReplaceStep(stepID, step)
		require.NoError(t, err)
	}
}

// WithMockedGCS modifies the workflow to mock all GCS upload steps
// (which use the google-github-actions/upload-cloud-storage action)
// to instead copy files to a local folder mounted into the act container at /gcs.
// It also takes all google-github-actions/auth steps and removes them,
// as authentication is not needed for local file copy.
// This allows testing GCS upload functionality without actually accessing GCS.
// Since GCS is only used in trusted contexts, callers should most likely also use WithMockedWorkflowContext.
func WithMockedGCS(t *testing.T) WorkflowOption {
	return func(w *Workflow) {
		require.NoError(t, w.CIWorkflow().MockAllStepsUsingAction(workflow.GCSLoginAction, func(step workflow.Step) (workflow.Step, error) {
			return workflow.NoOpStep(step.ID), nil
		}))
		require.NoError(t, w.CIWorkflow().MockAllStepsUsingAction(workflow.GCSUploadAction, func(step workflow.Step) (workflow.Step, error) {
			return workflow.MockGCSUploadStep(step)
		}))
	}
}

// Static checks

var _ workflow.Workflow = Workflow{}
