package workflow

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	pciwfBaseRef = "grafana/plugin-ci-workflows/.github/workflows"
)

// SimpleCI is a predefined GitHub Actions workflow for testing plugins using act.
// It uses the plugin-ci-workflows CI workflow as a base, with sane default values
// and allows customization through options.
// It implements the Marshalable interface to allow conversion to YAML format.
// Instances must be created using NewSimpleCI.
type SimpleCI struct {
	*TestingWorkflow
}

// NewSimpleCI creates a new SimpleCI workflow instance with default settings.
// The caller can provide options to customize the workflow.
func NewSimpleCI(opts ...SimpleCIOption) (SimpleCI, error) {
	ciBaseWf := BaseWorkflow{
		Name: "CI",
		On: On{
			Push: OnPush{
				Branches: []string{"main"},
			},
			PullRequest: OnPullRequest{
				Branches: []string{"main"},
			},
		},
		Jobs: map[string]*Job{
			"ci": {
				Name: "CI",
				// This will be populated later with the child testing workflow reference
				// Uses: "..."
				Permissions: Permissions{
					"contents": "read",
					"id-token": "write",
				},
				With: map[string]any{
					"plugin-version-suffix": "${{ github.event_name == 'pull_request' && github.event.pull_request.head.sha || '' }}",
					"testing":               true,
				},
				Secrets: Secrets{
					"GITHUB_TOKEN": "${{ secrets.GITHUB_TOKEN }}",
				},
			},
		},
	}

	// We need to create a child testing workflow for the called "ci.yml" workflow, in order to mock jobs/steps in it
	// Read the base workflow from file to create the child BaseWorkflow
	childBaseWf, err := NewBaseWorkflowFromFile(filepath.Join(".github", "workflows", "ci.yml"))
	if err != nil {
		return SimpleCI{}, fmt.Errorf("new base workflow from file for child ci workflow: %w", err)
	}

	// Workflow to be returned
	testingWf := SimpleCI{NewTestingWorkflow("simple-ci", ciBaseWf)}

	// Add the child workflow (ci) now, so further customization can be done on it via opts.
	// Use the same UUID as the parent, so they have the same uuid in the file name
	// and it is easier to correlate them.
	childTestingWf := NewTestingWorkflow("ci", childBaseWf, withUUID(testingWf.uuid))
	testingWf.AddChild("ci", childTestingWf)

	// Change the parent workflow so it calls the mocked child workflow
	testingWf.Jobs["ci"].Uses = pciwfBaseRef + "/" + testingWf.GetChild("ci").FileName() + "@main"

	// Apply options to customize the SimpleCI instance.
	// These opts can also modify the child testing workflow.
	for _, opt := range opts {
		opt(&testingWf)
	}
	return testingWf, nil
}

// CIWorkflow returns the TestingWorkflow instance representing the "ci" child workflow.
// This can be used to further customize/mock steps and jobs in the child workflow.
func (w *SimpleCI) CIWorkflow() *TestingWorkflow {
	return w.GetChild("ci")
}

// SimpleCIOption is a function that modifies a SimpleCI instance during its construction.
type SimpleCIOption func(*SimpleCI)

// WithPluginDirectoryInput sets the plugin-directory input for the CI job in the SimpleCI workflow.
func WithPluginDirectoryInput(dir string) SimpleCIOption {
	return func(w *SimpleCI) {
		w.Jobs["ci"].With["plugin-directory"] = dir
	}
}

// WithDistArtifactPrefixInput sets the dist-artifacts-prefix input for the CI job in the SimpleCI workflow.
func WithDistArtifactPrefixInput(prefix string) SimpleCIOption {
	return func(w *SimpleCI) {
		w.Jobs["ci"].With["dist-artifacts-prefix"] = prefix
	}
}

// WithPlaywrightInput sets the run-playwright input for the CI job in the SimpleCI workflow.
func WithPlaywrightInput(enabled bool) SimpleCIOption {
	return func(w *SimpleCI) {
		w.Jobs["ci"].With["run-playwright"] = enabled
	}
}

// WithRunPluginValidatorInput sets the run-plugin-validator input for the CI job in the SimpleCI workflow.
func WithRunPluginValidatorInput(enabled bool) SimpleCIOption {
	return func(w *SimpleCI) {
		w.Jobs["ci"].With["run-plugin-validator"] = enabled
	}
}

// WithRunTruffleHogInput sets the run-trufflehog input for the CI job in the SimpleCI workflow.
func WithRunTruffleHogInput(enabled bool) SimpleCIOption {
	return func(w *SimpleCI) {
		w.Jobs["ci"].With["run-trufflehog"] = enabled
	}
}

// WithMockedDist modifies the SimpleCI workflow to mock the test-and-build job
// to copy pre-built dist files (js + assets + backend executable, NOT the ZIP files)
// from the tests/act/mockdata folder instead of building them.
// This can be used for tests that need to assert on side-effects of building the plugin,
// without actually building it, which saves execution time.
// The pluginFolder parameter is the name of the plugin folder inside tests/act/mockdata/dist.
func WithMockedDist(t *testing.T, pluginFolder string) SimpleCIOption {
	return func(w *SimpleCI) {
		testAndBuild := w.CIWorkflow().Jobs["test-and-build"]
		require.NoError(t, testAndBuild.RemoveStep("setup"))
		require.NoError(t, testAndBuild.ReplaceStep(
			"frontend",
			CopyMockFilesStep("dist/"+pluginFolder, "${{ github.workspace }}/${{ inputs.plugin-directory }}/dist/"),
		))
		require.NoError(t, testAndBuild.RemoveStep("backend"))
	}
}

// Static checks

var _ Workflow = SimpleCI{}
