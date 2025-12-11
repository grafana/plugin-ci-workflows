package main

import (
	"path/filepath"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/stretchr/testify/require"
)

func TestGCS(t *testing.T) {
	const folder = "simple-frontend"

	runner, err := act.NewRunner(t)
	require.NoError(t, err)

	wf, err := workflow.NewSimpleCI(
		workflow.WithPluginDirectoryInput(filepath.Join("tests", folder)),
		workflow.WithDistArtifactPrefixInput(folder+"-"),
		workflow.WithPlaywrightInput(false),

		// Mock dist so we don't spend time building it
		workflow.WithMockedDist(t, folder),
		// Mock a trusted context to enable GCS upload
		workflow.WithMockedWorkflowContext(t, workflow.Context{
			IsTrusted: true,
		}),
		// Mock all GCS access
		workflow.WithMockedGCS(t),

		// No-op steps that are normally executed in a trusted context
		// but are not relevant for this test and would error out otherwise.
		workflow.WithNoOpStep(t, "get-secrets"),
		workflow.WithNoOpStep(t, "generate-github-token"),
	)
	require.NoError(t, err)

	r, err := runner.Run(wf, act.NewEmptyEventPayload())
	require.NoError(t, err)
	require.True(t, r.Success, "workflow should succeed")

	// Assert files uploaded to GCS
	files, err := runner.GCS.List("integration-artifacts/grafana-simplefrontend-panel/1.0.0")
	require.NoError(t, err)
	t.Log(files)
}
