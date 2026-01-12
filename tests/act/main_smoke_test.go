package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
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
		{
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
		},
	} {
		t.Run(tc.folder, func(t *testing.T) {
			t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)

			wf, err := workflow.NewSimpleCI(
				workflow.WithPluginDirectoryInput(filepath.Join("tests", tc.folder)),
				workflow.WithDistArtifactPrefixInput(tc.folder+"-"),
				workflow.WithPlaywrightInput(false),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewPushEventPayload("main"))
			require.NoError(t, err)
			require.True(t, r.Success, "workflow should succeed")

			// Assert outputs
			var pluginOutput testAndBuildOutput
			rawOutput, ok := r.Outputs.Get("test-and-build", "outputs", "plugin")
			require.True(t, ok, "plugin output should be present")
			err = json.Unmarshal([]byte(rawOutput), &pluginOutput)
			require.NoError(t, err, "unmarshal plugin output JSON")
			require.Equal(t, tc.exp, pluginOutput)

			// Sanity check the artifacts content (plugin ZIP files)
			hasBackend := tc.exp.HasBackend == "true"
			runID, err := r.GetTestingWorkflowRunID()
			require.NoError(t, err)
			distArtifacts, err := runner.ArtifactsStorage.GetFolder(runID, tc.folder+"-dist-artifacts")
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, distArtifacts.Close()) })

			anyZipFn := anyZipFileName(tc.exp.ID, tc.exp.Version)
			expFns := []string{
				anyZipFn,
				anyZipFn + ".md5",
				anyZipFn + ".sha1",
			}
			if hasBackend {
				// Additional zips for os/arch combos
				for _, osArch := range osArchCombos {
					osArchFn := osArchZipFileName(tc.exp.ID, tc.exp.Version, osArch)
					expFns = append(expFns, osArchFn, osArchFn+".md5", osArchFn+".sha1")
				}
			}
			require.NoError(t, checkFilesExist(distArtifacts.Fs, expFns, checkFilesExistOptions{strict: true}))

			// Sanity check the content of the "any" zip file
			zfs, err := distArtifacts.OpenZIP(anyZipFn)
			require.NoError(t, err)
			require.NoError(t, checkFilesExist(zfs, []string{
				filepath.Join(tc.exp.ID, "plugin.json"),
				filepath.Join(tc.exp.ID, "module.js"),
			}))
			if hasBackend {
				for _, osArch := range osArchCombos {
					if strings.Contains(osArch, "windows") {
						osArch = osArch + ".exe"
					}
					require.NoError(t, checkFilesExist(zfs, []string{
						filepath.Join(tc.exp.ID, tc.exp.Executable+"_"+osArch),
					}))
				}
			}
			require.NoError(
				t,
				checkFilesDontExist(zfs, []string{filepath.Join(tc.exp.ID, "MANIFEST.txt")}),
				"plugin should not be signed",
			)
		})
	}
}
