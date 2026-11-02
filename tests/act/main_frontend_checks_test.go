package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
)

// Static tests for the test-and-build runner input and the inputs that skip
// frontend checks a caller runs elsewhere. TestCDWorkflowContainsAllCIInputs
// and TestCDWorkflowCIJobPassesAllInputs cover the CD pass-through.

const frontendActionPath = "actions/internal/plugins/frontend/action.yml"

func TestTestAndBuildRunsOnInput(t *testing.T) {
	t.Parallel()

	ci := loadWorkflow(t, "ci.yml")
	in := ci.On.WorkflowCall.Inputs["test-and-build-runs-on"]
	require.Equal(t, workflow.WorkflowCallInputTypeString, in.Type)
	require.Equal(t, "ubuntu-x64-large", in.Default, "the default keeps today's runner")
	require.Equal(t, "${{ inputs.test-and-build-runs-on }}", ci.Jobs["test-and-build"].RunsOn)
}

func TestFrontendChecksCanBeSkipped(t *testing.T) {
	t.Parallel()

	ci := loadWorkflow(t, "ci.yml")
	frontend := ci.Jobs["test-and-build"].GetStep("frontend")
	require.NotNil(t, frontend)

	action, err := workflow.NewCompositeActionFromFile(frontendActionPath)
	require.NoError(t, err)

	for _, tc := range []struct {
		ciInput, actionInput, stepName string
	}{
		{"run-frontend-lint", "run-lint", "Lint"},
		{"run-frontend-typecheck", "run-typecheck", "Typecheck"},
		{"run-frontend-tests", "run-tests", "Test"},
	} {
		t.Run(tc.ciInput, func(t *testing.T) {
			t.Parallel()

			in := ci.On.WorkflowCall.Inputs[tc.ciInput]
			require.Equal(t, workflow.WorkflowCallInputTypeBoolean, in.Type)
			require.Equal(t, true, in.Default, "checks run unless the caller opts out")
			require.Equal(t, "${{ inputs."+tc.ciInput+" }}", frontend.With[tc.actionInput])

			require.Equal(t, "true", action.Inputs[tc.actionInput].Default)
			var found bool
			for _, step := range action.Runs.Steps {
				if step.Name == tc.stepName {
					found = true
					require.Equal(t, "inputs."+tc.actionInput+" == 'true'", step.If)
				}
			}
			require.True(t, found, "step %q not found", tc.stepName)
		})
	}

	// Install and build always run.
	for _, step := range action.Runs.Steps {
		if step.Name == "Install dependencies" || step.Name == "Build" {
			require.Empty(t, step.If, step.Name)
		}
	}
}
