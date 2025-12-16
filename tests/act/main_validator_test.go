package main

import (
	"path/filepath"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/stretchr/testify/require"
)

func TestValidator(t *testing.T) {
	runner, err := act.NewRunner(t, act.WithLinuxAMD64ContainerArchitecture())
	require.NoError(t, err)

	wf, err := workflow.NewSimpleCI(
		workflow.WithPluginDirectoryInput(filepath.Join("tests", "simple-frontend")),
		workflow.WithDistArtifactPrefixInput("simple-frontend-"),

		// Disable some features to speed up the test
		workflow.WithPlaywrightInput(false),
		workflow.WithRunTruffleHogInput(false),

		// Enable the plugin validator (opt-in)
		workflow.WithRunPluginValidatorInput(true),

		// Mock dist so we don't spend time building the plugin
		workflow.WithMockedPackagedDistArtifacts(t, "simple-frontend", false),
	)
	require.NoError(t, err)

	r, err := runner.Run(wf, act.NewEmptyEventPayload())
	require.NoError(t, err)
	require.True(t, r.Success, "workflow should succeed")
}
