package cd

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow/ci"
	"github.com/stretchr/testify/require"
)

// Workflow is a predefined GitHub Actions workflow for testing the CD workflow using act.
// It uses the plugin-ci-workflows CD workflow as a base, with sane default values
// and allows customization through options.
// It implements the Workflow interface to allow conversion to YAML format.
// Instances must be created using NewSimpleCD.
//
// The workflow has a nested structure:
//   - Parent workflow (simple-cd): calls cd.yml
//   - CD child workflow (cd): the mocked cd.yml, calls ci.yml
//   - CI grandchild workflow (ci): the mocked ci.yml
type Workflow struct {
	*workflow.TestingWorkflow

	ciWorkflow *ci.Workflow
}

// NewWorkflow creates a new SimpleCD workflow instance with default settings.
// The caller can provide options to customize the workflow.
func NewWorkflow(opts ...WorkflowOption) (Workflow, error) {
	cdBaseWf := workflow.BaseWorkflow{
		Name: "CD",
		On: workflow.On{
			Push: workflow.OnPush{
				Branches: []string{"main"},
			},
		},
		Jobs: map[string]*workflow.Job{
			"cd": {
				Name: "CD",
				// This will be populated later with the child testing workflow reference
				// Uses: "..."
				Permissions: workflow.Permissions{
					"contents":      "write",
					"id-token":      "write",
					"attestations":  "write",
					"pull-requests": "read",
				},
				With: map[string]any{
					"environment": "dev",
					"branch":      "${{ github.event_name == 'push' && github.ref_name || github.ref }}",
				},
			},
		},
	}

	// Read cd.yml to create the CD child workflow
	cdChildBaseWf, err := workflow.NewBaseWorkflowFromFile(filepath.Join(".github", "workflows", "cd.yml"))
	if err != nil {
		return Workflow{}, fmt.Errorf("new base workflow from file for child cd workflow: %w", err)
	}

	// Read ci.yml to create the CI grandchild workflow (cd.yml calls ci.yml)
	/* ciGrandchildBaseWf, err := workflow.NewBaseWorkflowFromFile(filepath.Join(".github", "workflows", "ci.yml"))
	if err != nil {
		return Workflow{}, fmt.Errorf("new base workflow from file for grandchild ci workflow: %w", err)
	} */

	// Create the parent workflow
	testingWf := Workflow{TestingWorkflow: workflow.NewTestingWorkflow("simple-cd", cdBaseWf)}

	// Add the CD child workflow
	// Use the same UUID as the parent for correlation
	cdChildTestingWf := workflow.NewTestingWorkflow("cd", cdChildBaseWf)
	testingWf.AddChild("cd", cdChildTestingWf)

	// Add the CI grandchild workflow as a child of the CD workflow
	// ciGrandchildTestingWf := workflow.NewTestingWorkflow("ci", ciGrandchildBaseWf, workflow.WithUUID(testingWf.UUID()))
	// cdChildTestingWf.AddChild("ci", ciGrandchildTestingWf)
	ciGrandchildWf, err := ci.NewWorkflow()
	if err != nil {
		return Workflow{}, fmt.Errorf("new ci grandchild workflow: %w", err)
	}
	cdChildTestingWf.AddChild("ci", ciGrandchildWf.TestingWorkflow)
	testingWf.ciWorkflow = &ciGrandchildWf

	// Update the parent workflow to call the mocked CD workflow
	testingWf.BaseWorkflow.Jobs["cd"].Uses = workflow.PCIWFBaseRef + "/" + testingWf.GetChild("cd").FileName() + "@main"

	// Update the CD workflow to call the mocked CI workflow
	// The CD workflow has a "ci" job that calls ci.yml (line 541 in cd.yml)
	if ciJob, ok := cdChildTestingWf.BaseWorkflow.Jobs["ci"]; ok {
		ciJob.Uses = workflow.PCIWFBaseRef + "/" + ciGrandchildWf.FileName() + "@main"
	}

	// Apply options to customize the SimpleCD instance.
	// These opts can also modify the child and grandchild workflows.
	for _, opt := range opts {
		opt(&testingWf)
	}
	testingWf.AddUUIDToAllJobsRecursive()
	return testingWf, nil
}

// CDWorkflow returns the TestingWorkflow instance representing the "cd" child workflow.
// This can be used to further customize/mock steps and jobs in the CD workflow.
func (w *Workflow) CDWorkflow() *workflow.TestingWorkflow {
	return w.GetChild("cd")
}

// CIWorkflow returns the TestingWorkflow instance representing the "ci" grandchild workflow.
// This can be used to further customize/mock steps and jobs in the CI workflow
// that is called by the CD workflow.
func (w *Workflow) CIWorkflow() *workflow.TestingWorkflow {
	return w.CDWorkflow().GetChild("ci")
}

// WorkflowOption is a function that modifies a Workflow instance during its construction.
type WorkflowOption func(*Workflow)

// WorkflowInputs are the inputs for the CD workflow.
// They are used to customize the CD workflow.
type WorkflowInputs struct {
	// CI options (shared between CI and CD)
	CI ci.WorkflowInputs

	Environment                *string
	Branch                     *string
	Scopes                     *string
	GrafanaCloudDeploymentType *string
	DisableDocsPublishing      *bool
	DisableGitHubRelease       *bool
	TriggerArgo                *bool

	// GCOMApiURL overrides the GCOM API URL for testing with mock servers.
	// Use GCOMMock.DockerAccessibleURL() to get a Docker-accessible URL.
	GCOMApiURL *string
}

// WithWorkflowInputs sets the inputs for the CD workflow.
func WithWorkflowInputs(inputs WorkflowInputs) WorkflowOption {
	return func(w *Workflow) {
		job := w.BaseWorkflow.Jobs["cd"]
		ci.SetCIInputs(job, inputs.CI)
		workflow.SetJobInput(job, "environment", inputs.Environment)
		workflow.SetJobInput(job, "branch", inputs.Branch)
		workflow.SetJobInput(job, "scopes", inputs.Scopes)
		workflow.SetJobInput(job, "grafana-cloud-deployment-type", inputs.GrafanaCloudDeploymentType)
		workflow.SetJobInput(job, "disable-docs-publishing", inputs.DisableDocsPublishing)
		workflow.SetJobInput(job, "disable-github-release", inputs.DisableGitHubRelease)
		workflow.SetJobInput(job, "trigger-argo", inputs.TriggerArgo)
		workflow.SetJobInput(job, "DO-NOT-USE-gcom-api-url", inputs.GCOMApiURL)
	}
}

func WithCIOptions(opts ...ci.WorkflowOption) WorkflowOption {
	return func(w *Workflow) {
		for _, opt := range opts {
			opt(w.ciWorkflow)
		}
	}
}

// WithMockedArgoWorkflows modifies the SimpleCD workflow to mock the Argo Workflow trigger step
// (which uses the grafana/shared-workflows/actions/trigger-argo-workflow action)
// to instead return a mock URI.
// This allows testing CD workflows without actually triggering Argo Workflows.
func WithMockedArgoWorkflows(t *testing.T) WorkflowOption {
	return func(w *Workflow) {
		err := w.CDWorkflow().MockAllStepsUsingAction(workflow.ArgoWorkflowAction, func(step workflow.Step) (workflow.Step, error) {
			return workflow.MockArgoWorkflowStep(step)
		})
		require.NoError(t, fmt.Errorf("mock argo workflow step: %w", err))
	}
}

// WithMockedGCOM configures the workflow to use the provided GCOM mock server.
// This sets the GCOM API URL to the mock server's Docker-accessible URL.
func WithMockedGCOM(mock *act.GCOM) WorkflowOption {
	return func(w *Workflow) {
		url := mock.DockerAccessibleURL()
		job := w.BaseWorkflow.Jobs["cd"]
		workflow.SetJobInput(job, "DO-NOT-USE-gcom-api-url", &url)
	}
}

type workflowGetter func(*Workflow) *workflow.TestingWorkflow

// simpleCDWorkflowMutator is a helper to mutate the SimpleCD workflow or its children workflows
// with options that are not specific to the SimpleCI workflow itself, but rather to the testing workflow in general.
type workflowMutator struct {
	workflowGetters []workflowGetter
}

// MutateTestingWorkflow returns a simpleCDWorkflowMutator that can be used to mutate the testing workflow.
func MutateTestingWorkflow() workflowMutator {
	return workflowMutator{
		workflowGetters: []workflowGetter{func(w *Workflow) *workflow.TestingWorkflow {
			return w.TestingWorkflow
		}},
	}
}

// MutateCIWorkflow returns a simpleCDWorkflowMutator that can be used to mutate the CD workflow
// (child of the testing workflow).
func MutateCDWorkflow() workflowMutator {
	return workflowMutator{
		workflowGetters: []workflowGetter{func(w *Workflow) *workflow.TestingWorkflow {
			return w.CDWorkflow()
		}},
	}
}

// CDMutateCIWorkflow returns a simpleCDWorkflowMutator that can be used to mutate the CI workflow
// (grandchild of the CD workflow).
func MutateCIWorkflow() workflowMutator {
	return workflowMutator{
		workflowGetters: []workflowGetter{func(w *Workflow) *workflow.TestingWorkflow {
			return w.CIWorkflow()
		}},
	}
}

func MutateAllWorkflows() workflowMutator {
	return workflowMutator{
		workflowGetters: []workflowGetter{
			func(w *Workflow) *workflow.TestingWorkflow {
				return w.TestingWorkflow
			},
			func(w *Workflow) *workflow.TestingWorkflow {
				return w.CDWorkflow()
			},
			func(w *Workflow) *workflow.TestingWorkflow {
				return w.CIWorkflow()
			},
		},
	}
}

// With applies the given options to the workflow returned by the workflowGetter function.
func (m workflowMutator) With(opts ...workflow.TestingWorkflowOption) WorkflowOption {
	return func(w *Workflow) {
		for i, getter := range m.workflowGetters {
			wf := getter(w)
			if wf == nil {
				fmt.Println("PANIC!!", i)
			}
			for _, opt := range opts {
				opt(wf)
			}
		}
	}
}

// Static checks
var _ workflow.Workflow = Workflow{}
