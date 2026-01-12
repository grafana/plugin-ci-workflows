package cd

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
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
					"branch":      "main",
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
	ciGrandchildBaseWf, err := workflow.NewBaseWorkflowFromFile(filepath.Join(".github", "workflows", "ci.yml"))
	if err != nil {
		return Workflow{}, fmt.Errorf("new base workflow from file for grandchild ci workflow: %w", err)
	}

	// Create the parent workflow
	testingWf := Workflow{workflow.NewTestingWorkflow("simple-cd", cdBaseWf)}

	// Add the CD child workflow
	// Use the same UUID as the parent for correlation
	cdChildTestingWf := workflow.NewTestingWorkflow("cd", cdChildBaseWf, workflow.WithUUID(testingWf.UUID()))
	testingWf.AddChild("cd", cdChildTestingWf)

	// Add the CI grandchild workflow as a child of the CD workflow
	ciGrandchildTestingWf := workflow.NewTestingWorkflow("ci", ciGrandchildBaseWf, workflow.WithUUID(testingWf.UUID()))
	cdChildTestingWf.AddChild("ci", ciGrandchildTestingWf)

	// Update the parent workflow to call the mocked CD workflow
	testingWf.BaseWorkflow.Jobs["cd"].Uses = workflow.PCIWFBaseRef + "/" + testingWf.GetChild("cd").FileName() + "@main"

	// Update the CD workflow to call the mocked CI workflow
	// The CD workflow has a "ci" job that calls ci.yml (line 541 in cd.yml)
	if ciJob, ok := cdChildTestingWf.BaseWorkflow.Jobs["ci"]; ok {
		ciJob.Uses = workflow.PCIWFBaseRef + "/" + ciGrandchildTestingWf.FileName() + "@main"
	}

	// Add uuid to each job in the workflow and all its children in order to
	// make unique container names and allow tests to run in parallel, so that
	// container names created by act don't clash
	// TODO: move to TestingWorkflow instead?
	allWorkflows := []workflow.Workflow{testingWf.TestingWorkflow}
	allWorkflows = append(allWorkflows, testingWf.Children()...)
	// Also include grandchildren (CI workflow)
	for _, child := range testingWf.Children() {
		allWorkflows = append(allWorkflows, child.Children()...)
	}
	for _, wf := range allWorkflows {
		for _, j := range wf.Jobs() {
			if j.Name != "" {
				j.Name = j.Name + "-" + testingWf.UUID().String()
			} else {
				j.Name = testingWf.UUID().String()
			}
		}
	}

	// Apply options to customize the SimpleCD instance.
	// These opts can also modify the child and grandchild workflows.
	for _, opt := range opts {
		opt(&testingWf)
	}
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

// WithCDEnvironmentInput sets the environment input for the CD job in the workflow.
func WithCDEnvironmentInput(env string) WorkflowOption {
	return func(w *Workflow) {
		w.BaseWorkflow.Jobs["cd"].With["environment"] = env
	}
}

// WithCDBranchInput sets the branch input for the CD job in the workflow.
func WithCDBranchInput(branch string) WorkflowOption {
	return func(w *Workflow) {
		w.BaseWorkflow.Jobs["cd"].With["branch"] = branch
	}
}

// WithCDPluginDirectoryInput sets the plugin-directory input for the CD job in the workflow.
func WithCDPluginDirectoryInput(dir string) WorkflowOption {
	return func(w *Workflow) {
		w.BaseWorkflow.Jobs["cd"].With["plugin-directory"] = dir
	}
}

// WithCDScopesInput sets the scopes input for the CD job in the workflow.
func WithCDScopesInput(scopes string) WorkflowOption {
	return func(w *Workflow) {
		w.BaseWorkflow.Jobs["cd"].With["scopes"] = scopes
	}
}

// WithCDGrafanaCloudDeploymentTypeInput sets the grafana-cloud-deployment-type input for the CD job in the workflow.
func WithCDGrafanaCloudDeploymentTypeInput(deploymentType string) WorkflowOption {
	return func(w *Workflow) {
		w.BaseWorkflow.Jobs["cd"].With["grafana-cloud-deployment-type"] = deploymentType
	}
}

// WithCDTriggerArgoInput sets the trigger-argo input for the CD job in the workflow.
func WithCDTriggerArgoInput(triggerArgo bool) WorkflowOption {
	return func(w *Workflow) {
		w.BaseWorkflow.Jobs["cd"].With["trigger-argo"] = triggerArgo
	}
}

// WithCDMockedVault modifies the SimpleCD workflow to mock all Vault secrets steps
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
func WithCDMockedVault(t *testing.T, secrets workflow.VaultSecrets) WorkflowOption {
	return func(w *Workflow) {
		err := w.CDWorkflow().MockAllStepsUsingAction(workflow.VaultSecretsAction, func(step workflow.Step) (workflow.Step, error) {
			return workflow.MockVaultSecretsStep(step, secrets)
		})
		require.NoError(t, fmt.Errorf("mock vault secrets step: %w", err))
	}
}

// WithCDMockedArgoWorkflow modifies the SimpleCD workflow to mock the Argo Workflow trigger step
// (which uses the grafana/shared-workflows/actions/trigger-argo-workflow action)
// to instead return a mock URI.
// This allows testing CD workflows without actually triggering Argo Workflows.
func WithCDMockedArgoWorkflow(t *testing.T) WorkflowOption {
	return func(w *Workflow) {
		err := w.CDWorkflow().MockAllStepsUsingAction(workflow.ArgoWorkflowAction, func(step workflow.Step) (workflow.Step, error) {
			return workflow.MockArgoWorkflowStep(step)
		})
		require.NoError(t, fmt.Errorf("mock argo workflow step: %w", err))
	}
}

// simpleCDWorkflowMutator is a helper to mutate the SimpleCD workflow or its children workflows
// with options that are not specific to the SimpleCI workflow itself, but rather to the testing workflow in general.
type workflowMutator struct {
	workflowGetter func(*Workflow) *workflow.TestingWorkflow
}

// MutateTestingWorkflow returns a simpleCDWorkflowMutator that can be used to mutate the testing workflow.
func MutateTestingWorkflow() workflowMutator {
	return workflowMutator{
		workflowGetter: func(w *Workflow) *workflow.TestingWorkflow {
			return w.TestingWorkflow
		},
	}
}

// MutateCIWorkflow returns a simpleCDWorkflowMutator that can be used to mutate the CD workflow
// (child of the testing workflow).
func MutateCDWorkflow() workflowMutator {
	return workflowMutator{
		workflowGetter: func(w *Workflow) *workflow.TestingWorkflow {
			return w.CDWorkflow()
		},
	}
}

// CDMutateCIWorkflow returns a simpleCDWorkflowMutator that can be used to mutate the CI workflow
// (grandchild of the CD workflow).
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

// Static checks
var _ workflow.Workflow = Workflow{}
