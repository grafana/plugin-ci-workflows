package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
)

// Static tests for the test-and-build runner input and the inputs that skip
// frontend and backend checks a caller runs elsewhere. TestCDWorkflowContainsAllCIInputs
// and TestCDWorkflowCIJobPassesAllInputs cover the CD pass-through.

const (
	frontendActionPath = "actions/internal/plugins/frontend/action.yml"
	backendActionPath  = "actions/internal/plugins/backend/action.yml"
)

func TestTestAndBuildRunsOnInput(t *testing.T) {
	t.Parallel()

	ci := loadWorkflow(t, "ci.yml")
	in := ci.On.WorkflowCall.Inputs["test-and-build-runs-on"]
	require.Equal(t, workflow.WorkflowCallInputTypeString, in.Type)
	require.Equal(t, "ubuntu-x64-large", in.Default, "the default keeps today's runner")
	require.Equal(t, "${{ inputs.test-and-build-runs-on }}", ci.Jobs["test-and-build"].RunsOn)
}

type skippableCheck struct {
	ciInput, actionInput, stepName string
}

func TestFrontendChecksCanBeSkipped(t *testing.T) {
	t.Parallel()

	requireChecksCanBeSkipped(t, "frontend", frontendActionPath, []skippableCheck{
		{"run-frontend-lint", "run-lint", "Lint"},
		{"run-frontend-typecheck", "run-typecheck", "Typecheck"},
		{"run-frontend-tests", "run-tests", "Test"},
	})
}

func TestBackendChecksCanBeSkipped(t *testing.T) {
	t.Parallel()

	requireChecksCanBeSkipped(t, "backend", backendActionPath, []skippableCheck{
		{"run-backend-lint", "run-lint", "Lint"},
		{"run-backend-tests", "run-tests", "Test"},
	})
}

// requireChecksCanBeSkipped checks that each ci.yml input reaches the
// composite action, defaults to true, and guards its step, and that the
// install and build steps always run.
func requireChecksCanBeSkipped(t *testing.T, stepID, actionPath string, checks []skippableCheck) {
	t.Helper()

	ci := loadWorkflow(t, "ci.yml")
	step := ci.Jobs["test-and-build"].GetStep(stepID)
	require.NotNil(t, step)

	action, err := workflow.NewCompositeActionFromFile(actionPath)
	require.NoError(t, err)

	for _, tc := range checks {
		t.Run(tc.ciInput, func(t *testing.T) {
			t.Parallel()

			in := ci.On.WorkflowCall.Inputs[tc.ciInput]
			require.Equal(t, workflow.WorkflowCallInputTypeBoolean, in.Type)
			require.Equal(t, true, in.Default, "checks run unless the caller opts out")
			require.Equal(t, "${{ inputs."+tc.ciInput+" }}", step.With[tc.actionInput])

			require.Equal(t, "true", action.Inputs[tc.actionInput].Default)
			var found bool
			for _, s := range action.Runs.Steps {
				if s.Name == tc.stepName {
					found = true
					require.Equal(t, "inputs."+tc.actionInput+" == 'true'", s.If)
				}
			}
			require.True(t, found, "step %q not found", tc.stepName)
		})
	}

	for _, s := range action.Runs.Steps {
		if s.Name == "Install dependencies" || s.Name == "Build" {
			require.Empty(t, s.If, s.Name)
		}
	}
}
