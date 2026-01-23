package main

import (
	"path/filepath"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/stretchr/testify/require"
)

func TestCDOutputs_HasCIOutputs(t *testing.T) {
	// Ensures that the outputs of the CD workflow are in-sync with the ones from the CI workflow
	cdWf, err := workflow.NewBaseWorkflowFromFile(filepath.Join(".github", "workflows", "cd.yml"))
	require.NoError(t, err)

	ciWf, err := workflow.NewBaseWorkflowFromFile(filepath.Join(".github", "workflows", "ci.yml"))
	require.NoError(t, err)

	ciOutputs := ciWf.On.WorkflowCall.Outputs
	require.NotEmpty(t, ciOutputs, "ci.yml outputs should not be empty")

	cdOutputs := cdWf.On.WorkflowCall.Outputs
	require.NotEmpty(t, cdOutputs, "cd.yml outputs should not be empty")

	for outputName := range ciOutputs {
		cdOutputValue, exists := cdOutputs[outputName]
		require.True(t, exists, "output %q from ci.yml not found in cd.yml", outputName)
		require.Equal(t, "${{ jobs.ci.outputs."+outputName+" }}", cdOutputValue.Value, "output %q value in cd.yml does not match expected value", outputName)
	}
}
