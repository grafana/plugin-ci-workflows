package main

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/stretchr/testify/require"
)

type testAndBuildOutput struct {
	ID         string `json:"id"`
	Version    string `json:"version"`
	HasBackend string `json:"has-backend"`
	Executable string `json:"executable"`
}

func TestSmoke(t *testing.T) {
	type cas struct {
		folder string
		exp    testAndBuildOutput
	}

	for _, tc := range []cas{
		{
			folder: "simple-frontend",
			exp: testAndBuildOutput{
				ID:         "grafana-simplefrontend-panel",
				Version:    "1.0.0",
				HasBackend: "false",
				Executable: "null",
			},
		},
		/* {
			folder: "simple-frontend-yarn",
			exp: testAndBuildOutput{
				ID:         "grafana-simplefrontendyarn-panel",
				Version:    "1.0.0",
				HasBackend: "false",
				Executable: "null",
			},
		},
		{
			folder: "simple-frontend-pnpm",
			exp: testAndBuildOutput{
				ID:         "grafana-simplefrontendpnpm-panel",
				Version:    "1.0.0",
				HasBackend: "false",
				Executable: "null",
			},
		},
		{
			folder: "simple-backend",
			exp: testAndBuildOutput{
				ID:         "grafana-simplebackend-datasource",
				Version:    "1.0.0",
				HasBackend: "true",
				Executable: "gpx_simple_backend",
			},
		}, */
	} {
		t.Run(tc.folder, func(t *testing.T) {
			runner, err := act.NewRunner(t)
			require.NoError(t, err)
			// runner.Verbose = true

			wf, err := workflow.NewSimpleCI(
				workflow.WithPluginDirectoryInput(filepath.Join("tests", tc.folder)),
				workflow.WithDistArtifactPrefixInput(tc.folder+"-"),
				workflow.WithPlaywrightInput(false),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewEmptyEventPayload())
			require.NoError(t, err)
			require.True(t, r.Success, "workflow should succeed")

			// Assert outputs
			var pluginOutput testAndBuildOutput
			rawOutput, ok := r.Outputs.Get("test-and-build", "outputs", "plugin")
			require.True(t, ok, "plugin output should be present")
			t.Log(rawOutput)
			err = json.Unmarshal([]byte(rawOutput), &pluginOutput)
			require.NoError(t, err, "unmarshal plugin output JSON")
			require.Equal(t, tc.exp, pluginOutput)

			// Sanity check the artifacts content (built plugin)
			if tc.exp.HasBackend == "true" {
				// TODO: ...
				return
			}
			runID, err := r.GetTestingWorkflowRunID()
			require.NoError(t, err)
			distArtifacts, err := runner.ArtifactsStorage.GetFolder(runID, tc.folder+"-dist-artifacts")
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, distArtifacts.Close()) })

			anyZipFn := anyZipFileName(tc.exp.ID, tc.exp.Version)
			require.NoError(t, checkFilesExist(distArtifacts.Fs, []string{
				anyZipFn,
				anyZipFn + ".md5",
				anyZipFn + ".sha1",
			}, checkFilesExistOptions{strict: true}))
			zfs, err := distArtifacts.OpenZIP(anyZipFn)
			require.NoError(t, err)
			require.NoError(t, checkFilesExist(zfs, []string{
				"plugin.json",
				"module.js",
			}))
			if tc.exp.HasBackend == "true" {
				require.NoError(t, checkFilesExist(zfs, []string{
					tc.exp.Executable,
				}))
			}
			require.NoError(t, checkFilesDontExist(zfs, []string{"MANIFEST.txt"}), "plugin should not be signed")
		})
	}
}
