package main

import (
	"path/filepath"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow/ci"
	"github.com/stretchr/testify/require"
)

func TestDistArtifactsUnavailable(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name                       string
		distArtifactsRetentionDays *int
		expAnnotationMessage       string
	}{
		{
			name:                       "default retention days",
			distArtifactsRetentionDays: nil,
			expAnnotationMessage:       "The dist-artifacts artifact could not be downloaded. It may have expired (retention period: 10 days). Please re-run the entire workflow from the beginning to rebuild the plugin.",
		},
		{
			name:                       "custom retention days",
			distArtifactsRetentionDays: workflow.Input(30),
			expAnnotationMessage:       "The dist-artifacts artifact could not be downloaded. It may have expired (retention period: 30 days). Please re-run the entire workflow from the beginning to rebuild the plugin.",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)

			wf, err := ci.NewWorkflow(
				ci.WithWorkflowInputs(ci.WorkflowInputs{
					PluginDirectory:            workflow.Input(filepath.Join("tests", "simple-frontend")),
					AllowUnsigned:              workflow.Input(true),
					RunPlaywright:              workflow.Input(false),
					RunTruffleHog:              workflow.Input(false),
					RunPluginValidator:         workflow.Input(false),
					DistArtifactsRetentionDays: tc.distArtifactsRetentionDays,
				}),

				ci.MutateCIWorkflow().With(
					// Keep only upload-to-gcs and its dependencies.
					// Dependencies must be kept (removeDependencies=false) so that the
					// needs.test-and-build.outputs.workflow-context expression in the
					// upload-to-gcs `if:` condition can be evaluated.
					workflow.WithOnlyOneJob(t, "upload-to-gcs", false),

					// No-op check-for-release-channel (indirect dependency via test-and-build).
					workflow.WithNoOpJobWithOutputs(t, "check-for-release-channel", map[string]string{}),

					// Replace test-and-build with a no-op that sets the required outputs.
					workflow.WithNoOpJobWithOutputs(t, "test-and-build", map[string]string{
						"workflow-context": `{"isTrusted":true}`,
						"plugin":           `{"id":"test-plugin","version":"1.0.0"}`,
					}),

					// Upload a placeholder artifact so act initializes the run directory.
					// act's artifact server panics ("no such file or directory") when
					// listArtifacts is called on a run that has never uploaded anything,
					// causing ECONNRESET before continue-on-error can be recorded.
					// Any artifact (even a different name) creates the directory so
					// download-artifact fails cleanly instead.
					workflow.WithInjectedSteps(t, "test-and-build", workflow.InjectedStepsOptions{
						Position:        workflow.InjectedStepsOptionsPositionAfter,
						InjectionStepID: "set-outputs",
						Steps: workflow.Steps{
							{
								Run:   `mkdir -p /tmp/placeholder-artifact && echo placeholder > /tmp/placeholder-artifact/placeholder.txt`,
								Shell: "bash",
							},
							{
								Name: "Upload placeholder artifact (act workaround)",
								Uses: "actions/upload-artifact@330a01c490aca151604b8cf639adc76d48f6c5d4", // v5.0.0
								With: map[string]any{
									"name": "placeholder-artifact",
									"path": "/tmp/placeholder-artifact/",
								},
							},
						},
					}),

					// Mock GCS auth and upload steps. The workflow fails before reaching them,
					// but mocking avoids authentication errors if the setup changes.
					workflow.WithMockedGCS(t),
				),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewPushEventPayload("main"))
			require.NoError(t, err)

			require.False(t, r.Success, "workflow should fail when dist-artifacts are unavailable")
			require.Contains(t, r.Annotations, act.Annotation{
				Level:   act.AnnotationLevelError,
				Message: tc.expAnnotationMessage,
			})
		})
	}
}

// When ci.yml and cd.yml run in the same workflow run without different dist-artifacts-prefix values,
// both upload an artifact with the same name and a job can download the one from the other build.
func TestDistArtifactsFromAnotherBuild(t *testing.T) {
	t.Parallel()

	const builtZip = "test-plugin-1.0.0+aaaaaaa.zip"
	for _, tc := range []struct {
		name        string
		artifactZip string
		expSuccess  bool
	}{
		{name: "artifact from this build", artifactZip: builtZip, expSuccess: true},
		{name: "artifact from another build", artifactZip: "test-plugin-1.0.0+bbbbbbb.zip", expSuccess: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)

			wf, err := ci.NewWorkflow(
				ci.WithWorkflowInputs(ci.WorkflowInputs{
					PluginDirectory:    workflow.Input(filepath.Join("tests", "simple-frontend")),
					AllowUnsigned:      workflow.Input(true),
					RunPlaywright:      workflow.Input(false),
					RunTruffleHog:      workflow.Input(false),
					RunPluginValidator: workflow.Input(false),
				}),

				ci.MutateCIWorkflow().With(
					// Keep the dependencies so the upload-to-gcs `if:` condition can be evaluated.
					workflow.WithOnlyOneJob(t, "upload-to-gcs", false),
					workflow.WithNoOpJobWithOutputs(t, "check-for-release-channel", map[string]string{}),
					workflow.WithNoOpJobWithOutputs(t, "test-and-build", map[string]string{
						"workflow-context": `{"isTrusted":true}`,
						"plugin":           `{"id":"test-plugin","version":"1.0.0+aaaaaaa"}`,
						"universal-zip":    builtZip,
					}),

					// Upload a dist-artifacts artifact that holds tc.artifactZip.
					workflow.WithInjectedSteps(t, "test-and-build", workflow.InjectedStepsOptions{
						Position:        workflow.InjectedStepsOptionsPositionAfter,
						InjectionStepID: "set-outputs",
						Steps: workflow.Steps{
							{
								Run:   `mkdir -p /tmp/dist-artifacts && echo zip > "/tmp/dist-artifacts/${ZIP}"`,
								Shell: "bash",
								Env:   map[string]string{"ZIP": tc.artifactZip},
							},
							{
								Name: "Upload dist-artifacts",
								Uses: "actions/upload-artifact@330a01c490aca151604b8cf639adc76d48f6c5d4", // v5.0.0
								With: map[string]any{
									"name": "dist-artifacts",
									"path": "/tmp/dist-artifacts/",
								},
							},
						},
					}),

					workflow.WithMockedGCS(t),
				),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewPushEventPayload("main"))
			require.NoError(t, err)

			mismatch := act.Annotation{
				Level: act.AnnotationLevelError,
				Title: "Wrong dist-artifacts",
				Message: "The dist-artifacts artifact does not contain " + builtZip + ". " +
					"Another ci.yml or cd.yml call in this workflow run probably uploaded an artifact with the same name. " +
					"Set a different dist-artifacts-prefix input on each call.",
			}
			require.Equal(t, tc.expSuccess, r.Success)
			if tc.expSuccess {
				require.NotContains(t, r.Annotations, mismatch)
			} else {
				require.Contains(t, r.Annotations, mismatch)
			}
		})
	}
}
