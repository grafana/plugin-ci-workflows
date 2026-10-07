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

// TestCatalogDocs tests that multi-page catalog docs are validated and packaged when
// plugin.json sets docsPath, and that nothing happens when it doesn't.
func TestCatalogDocs(t *testing.T) {
	const (
		folder        = "simple-frontend"
		pluginID      = "grafana-simplefrontend-panel"
		pluginVersion = "1.0.0"
	)

	inputs := ci.WorkflowInputs{
		PluginDirectory:     workflow.Input(filepath.Join("tests", folder)),
		DistArtifactsPrefix: workflow.Input(folder + "-"),
		RunPlaywright:       workflow.Input(false),
		RunTruffleHog:       workflow.Input(false),
	}

	expDebugMsg, err := logfmt.MarshalKeyvals("msg", "Building catalog docs", "docsPath", "docs")
	require.NoError(t, err)
	expDebugAnnotation := act.Annotation{
		Level:   act.AnnotationLevelDebug,
		Message: string(expDebugMsg),
	}

	t.Run("valid docs are packaged", func(t *testing.T) {
		t.Parallel()

		runner, err := act.NewRunner(t)
		require.NoError(t, err)
		wf, err := ci.NewWorkflow(
			ci.WithWorkflowInputs(inputs),
			ci.WithMockedDist(t, "dist/"+folder),
			withCatalogDocsFixture(t, "valid"),
		)
		require.NoError(t, err)

		r, err := runner.Run(wf, act.NewPushEventPayload("main"))
		require.NoError(t, err)
		require.True(t, r.Success, "workflow should succeed")
		require.Contains(t, r.Annotations, expDebugAnnotation)

		runID, err := r.GetTestingWorkflowRunID()
		require.NoError(t, err)
		distArtifacts, err := runner.ArtifactsStorage.GetFolder(runID, folder+"-dist-artifacts")
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, distArtifacts.Close()) })

		pluginZIP, err := distArtifacts.OpenZIP(anyZipFileName(pluginID, pluginVersion))
		require.NoError(t, err)
		require.NoError(t, checkFilesExist(pluginZIP, []string{
			filepath.Join(pluginID, "docs", "manifest.json"),
			filepath.Join(pluginID, "docs", "index.md"),
			filepath.Join(pluginID, "docs", "options.md"),
			filepath.Join(pluginID, "docs", "data-formats.md"),
			filepath.Join(pluginID, "docs", "img", "panel.png"),
		}))
	})

	t.Run("invalid docs fail the build", func(t *testing.T) {
		t.Parallel()

		runner, err := act.NewRunner(t)
		require.NoError(t, err)
		wf, err := ci.NewWorkflow(
			ci.WithWorkflowInputs(inputs),
			ci.WithMockedDist(t, "dist/"+folder),
			withCatalogDocsFixture(t, "invalid"),
			ci.MutateCIWorkflow().With(
				workflow.WithOnlyOneJob(t, "test-and-build", true),
				workflow.WithRemoveAllStepsAfter(t, "test-and-build", "catalog-docs"),
			),
		)
		require.NoError(t, err)

		r, err := runner.Run(wf, act.NewPushEventPayload("main"))
		require.NoError(t, err)
		require.False(t, r.Success, "workflow should fail when catalog docs are invalid")
		require.Contains(t, r.Annotations, expDebugAnnotation)
	})

	t.Run("no docsPath skips catalog docs", func(t *testing.T) {
		t.Parallel()

		runner, err := act.NewRunner(t)
		require.NoError(t, err)
		wf, err := ci.NewWorkflow(
			ci.WithWorkflowInputs(inputs),
			ci.WithMockedDist(t, "dist/"+folder),
			ci.MutateCIWorkflow().With(
				workflow.WithOnlyOneJob(t, "test-and-build", true),
				workflow.WithRemoveAllStepsAfter(t, "test-and-build", "catalog-docs"),
			),
		)
		require.NoError(t, err)

		r, err := runner.Run(wf, act.NewPushEventPayload("main"))
		require.NoError(t, err)
		require.True(t, r.Success, "workflow should succeed")
		require.NotContains(t, r.Annotations, expDebugAnnotation)
	})
}

// withCatalogDocsFixture copies a docs folder from mockdata/catalog-docs/<fixture> into the
// plugin directory and sets docsPath in both src/plugin.json and the mocked dist/plugin.json.
func withCatalogDocsFixture(t *testing.T, fixture string) ci.WorkflowOption {
	return ci.MutateCIWorkflow().With(
		workflow.WithInjectedSteps(t, "test-and-build", workflow.InjectedStepsOptions{
			Position:        workflow.InjectedStepsOptionsPositionBefore,
			InjectionStepID: "catalog-docs",
			Steps: workflow.Steps{
				{
					Name: "Add catalog docs fixture (mock)",
					Run: workflow.Commands{
						"set -x",
						"cp -r /mockdata/catalog-docs/" + fixture + "/. .",
						`for f in src/plugin.json dist/plugin.json; do jq '.docsPath = "docs"' "$f" > "$f.tmp" && mv "$f.tmp" "$f"; done`,
					}.String(),
					Shell:            "bash",
					WorkingDirectory: "${{ inputs.plugin-directory }}",
				},
			},
		}),
	)
}
