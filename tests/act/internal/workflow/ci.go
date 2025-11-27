package workflow

import (
	"fmt"
	"path/filepath"
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
	childTestingWf := NewTestingWorkflow("ci", childBaseWf, WithUUID(testingWf.uuid))
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

// SimpleCIOption is a function that modifies a SimpleCI instance during its construction.
type SimpleCIOption func(*SimpleCI)

// WithPluginDirectory sets the plugin-directory input for the CI job in the SimpleCI workflow.
func WithPluginDirectory(dir string) SimpleCIOption {
	return func(w *SimpleCI) {
		w.Jobs["ci"].With["plugin-directory"] = dir
	}
}

// WithDistArtifactPrefix sets the dist-artifacts-prefix input for the CI job in the SimpleCI workflow.
func WithDistArtifactPrefix(prefix string) SimpleCIOption {
	return func(w *SimpleCI) {
		w.Jobs["ci"].With["dist-artifacts-prefix"] = prefix
	}
}

// WithPlaywright sets the run-playwright input for the CI job in the SimpleCI workflow.
func WithPlaywright(enabled bool) SimpleCIOption {
	return func(w *SimpleCI) {
		w.Jobs["ci"].With["run-playwright"] = enabled
	}
}

// Static checks

var _ Workflow = SimpleCI{}
