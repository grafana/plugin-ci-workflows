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
	childTestingWf := workflow.NewTestingWorkflow("ci", childBaseWf)
	testingWf.AddChild("ci", childTestingWf)

	// Change the parent workflow so it calls the mocked child workflow
	testingWf.BaseWorkflow.Jobs["ci"].Uses = workflow.PCIWFBaseRef + "/" + testingWf.GetChild("ci").FileName() + "@main"

	// Apply options to customize the Workflow instance.
	// These opts can also modify the child testing workflow.
	for _, opt := range opts {
		opt(&testingWf)
	}
	testingWf.AddUUIDToAllJobsRecursive()
	return testingWf, nil
}

// CIWorkflow returns the TestingWorkflow instance representing the "ci" child workflow.
// This can be used to further customize/mock steps and jobs in the child workflow.
func (w *Workflow) CIWorkflow() *workflow.TestingWorkflow {
	return w.GetChild("ci")
}

// WorkflowOption is a function that modifies a Workflow instance during its construction.
type WorkflowOption func(*Workflow)

// WorkflowInputs are the inputs for the CI workflow.
// They are used to customize the CI workflow.
type WorkflowInputs struct {
	PluginDirectory     *string
	PluginVersionSuffix *string
	DistArtifactsPrefix *string

	GoVersion           *string
	NodeVersion         *string
	GolangciLintVersion *string
	MageVersion         *string
	TrufflehogVersion   *string

	RunPlaywright *bool

	RunPluginValidator    *bool
	PluginValidatorConfig *string

	RunTruffleHog *bool

	AllowUnsigned *bool
	Testing       *bool
}

// SetCIInputs sets the inputs for the CI workflow.
// This is a helper function to set the inputs for the CI workflow.
func SetCIInputs(dst *workflow.Job, inputs WorkflowInputs) {
	workflow.SetJobInput(dst, "plugin-directory", inputs.PluginDirectory)
	workflow.SetJobInput(dst, "plugin-version-suffix", inputs.PluginVersionSuffix)
	workflow.SetJobInput(dst, "dist-artifacts-prefix", inputs.DistArtifactsPrefix)

	workflow.SetJobInput(dst, "go-version", inputs.GoVersion)
	workflow.SetJobInput(dst, "node-version", inputs.NodeVersion)
	workflow.SetJobInput(dst, "golangci-lint-version", inputs.GolangciLintVersion)
	workflow.SetJobInput(dst, "mage-version", inputs.MageVersion)
	workflow.SetJobInput(dst, "trufflehog-version", inputs.TrufflehogVersion)

	workflow.SetJobInput(dst, "run-playwright", inputs.RunPlaywright)

	workflow.SetJobInput(dst, "run-plugin-validator", inputs.RunPluginValidator)
	workflow.SetJobInput(dst, "plugin-validator-config", inputs.PluginValidatorConfig)

	workflow.SetJobInput(dst, "run-trufflehog", inputs.RunTruffleHog)

	workflow.SetJobInput(dst, "allow-unsigned", inputs.AllowUnsigned)
	workflow.SetJobInput(dst, "testing", inputs.Testing)
}

// WithWorkflowInputs sets the inputs for the CI workflow.
func WithWorkflowInputs(inputs WorkflowInputs) WorkflowOption {
	return func(w *Workflow) {
		SetCIInputs(w.BaseWorkflow.Jobs["ci"], inputs)
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

// Static checks

var _ workflow.Workflow = Workflow{}
