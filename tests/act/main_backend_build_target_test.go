package main

import (
	"path/filepath"
	"testing"

	"github.com/go-logfmt/logfmt"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow/ci"
	"github.com/stretchr/testify/require"
)

func TestBackendBuildTarget(t *testing.T) {
	baseInputs := ci.WorkflowInputs{
		PluginDirectory:     workflow.Input(filepath.Join("tests", "simple-backend")),
		DistArtifactsPrefix: workflow.Input("simple-backend-"),
		RunPlaywright:       workflow.Input(false),
	}

	t.Run("custom target succeeds and is invoked", func(t *testing.T) {
		t.Parallel()

		runner, err := act.NewRunner(t)
		require.NoError(t, err)

		inputs := baseInputs
		inputs.BackendBuildTarget = workflow.Input("buildCustom")

		wf, err := ci.NewWorkflow(ci.WithWorkflowInputs(inputs))
		require.NoError(t, err)

		r, err := runner.Run(wf, act.NewPushEventPayload("main"))
		require.NoError(t, err)
		require.True(t, r.Success, "workflow should succeed when using a custom build target")

		expMsg, err := logfmt.MarshalKeyvals("msg", "custom build target invoked")
		require.NoError(t, err)
		require.Contains(t, r.Annotations, act.Annotation{
			Level:   act.AnnotationLevelDebug,
			Message: string(expMsg),
		}, "custom build target should have printed its debug annotation")
	})

	t.Run("non-existent target fails", func(t *testing.T) {
		t.Parallel()

		runner, err := act.NewRunner(t)
		require.NoError(t, err)

		inputs := baseInputs
		inputs.BackendBuildTarget = workflow.Input("buildNonExistentTarget")

		wf, err := ci.NewWorkflow(ci.WithWorkflowInputs(inputs))
		require.NoError(t, err)

		r, err := runner.Run(wf, act.NewPushEventPayload("main"))
		require.NoError(t, err)
		require.False(t, r.Success, "workflow should fail when using a non-existent build target")
	})
}
