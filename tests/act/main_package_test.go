package main

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow/ci"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// TestPackage tests the packaging step of the CI workflow.
func TestPackage(t *testing.T) {

	for _, tc := range []struct {
		folder           string
		expPluginID      string
		expPluginVersion string
		expBackend       bool
		expPluginType    string
		// expExecutables lists each backend executable's path prefix inside
		// the plugin folder, without the os/arch suffix. Per-platform zips
		// must contain exactly one binary per entry.
		expExecutables []string
		// expExtraDistFiles lists non-executable dist files beyond the common
		// base files (e.g. a nested datasource's own plugin files).
		expExtraDistFiles []string
	}{
		{
			folder:           "simple-frontend",
			expPluginID:      "grafana-simplefrontend-panel",
			expPluginVersion: "1.0.0",
			expBackend:       false,
			expPluginType:    "panel",
		},
		{
			folder:           "simple-backend",
			expPluginID:      "grafana-simplebackend-datasource",
			expPluginVersion: "1.0.0",
			expBackend:       true,
			expPluginType:    "datasource",
			expExecutables:   []string{"gpx_simple_backend"},
		},
		{
			// App with its own backend AND a nested datasource backend:
			// per-platform zips must filter the binaries of both backends
			// (https://github.com/grafana/plugin-ci-workflows/pull/882)
			folder:           "simple-app-nested-backend",
			expPluginID:      "grafana-simpleappnested-app",
			expPluginVersion: "1.0.0",
			expBackend:       true,
			expPluginType:    "app",
			expExecutables: []string{
				"gpx_simpleappnested_app",
				"datasource/gpx_simpleappnested_datasource",
			},
			expExtraDistFiles: []string{
				"datasource/plugin.json",
				"datasource/module.js",
			},
		},
	} {
		t.Run(tc.folder, func(t *testing.T) {
			t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)
			wf, err := ci.NewWorkflow(
				ci.WithWorkflowInputs(ci.WorkflowInputs{
					PluginDirectory:     workflow.Input(filepath.Join("tests", tc.folder)),
					DistArtifactsPrefix: workflow.Input(tc.folder + "-"),
					RunPlaywright:       workflow.Input(false),
					RunTruffleHog:       workflow.Input(false),
				}),
				// Mock the test-and-build job to copy pre-built dist files
				ci.WithMockedDist(t, "dist/"+tc.folder),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewPushEventPayload("main"))
			require.NoError(t, err)
			require.True(t, r.Success, "workflow should succeed")

			// Inspect the artifact and assert its contents
			runID, err := r.GetTestingWorkflowRunID()
			require.NoError(t, err)
			distArtifacts, err := runner.ArtifactsStorage.GetFolder(runID, tc.folder+"-dist-artifacts")
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, distArtifacts.Close()) })

			// Expect the "any" zip file + their hashes
			anyZipFn := anyZipFileName(tc.expPluginID, tc.expPluginVersion)
			expArtifactFiles := []string{
				anyZipFn,
				anyZipFn + ".md5",
				anyZipFn + ".sha1",
			}
			if tc.expBackend {
				// Expect the os/arch backend zips + their hashes
				for _, osArch := range osArchCombos {
					osArchZipFn := osArchZipFileName(tc.expPluginID, tc.expPluginVersion, osArch)
					expArtifactFiles = append(
						expArtifactFiles,
						osArchZipFn,
						osArchZipFn+".md5",
						osArchZipFn+".sha1",
					)
				}
			}
			require.NoError(t, checkFilesExist(distArtifacts.Fs, expArtifactFiles, checkFilesExistOptions{strict: true}))

			// Check the checksum files
			checkChecksumFiles := func(fn string) {
				zipFileContent, err := distArtifacts.ReadFile(fn)
				require.NoError(t, err)

				md5, err := distArtifacts.ReadFile(fn + ".md5")
				require.NoError(t, err)
				require.Equal(t, md5Hash(zipFileContent), string(md5), "wrong md5 checksum")

				sha1, err := distArtifacts.ReadFile(fn + ".sha1")
				require.NoError(t, err)
				require.Equal(t, sha1Hash(zipFileContent), string(sha1), "wrong sha1 checksum")
			}
			checkChecksumFiles(anyZipFn)

			// Check os/arch zip checksums
			if tc.expBackend {
				for _, osArch := range osArchCombos {
					checkChecksumFiles(osArchZipFileName(tc.expPluginID, tc.expPluginVersion, osArch))
				}
			}

			// Check the nested plugin ZIP artifact for the "any" zip and then for os/arch zips

			// Start from the "any" zip
			basePluginFiles := []string{
				filepath.Join(tc.expPluginID, "CHANGELOG.md"),
				filepath.Join(tc.expPluginID, "LICENSE"),
				filepath.Join(tc.expPluginID, "module.js"),
				filepath.Join(tc.expPluginID, "module.js.map"),
				filepath.Join(tc.expPluginID, "plugin.json"),
				filepath.Join(tc.expPluginID, "README.md"),
				filepath.Join(tc.expPluginID, "img/logo.svg"),
			}
			for _, extraFile := range tc.expExtraDistFiles {
				basePluginFiles = append(basePluginFiles, filepath.Join(tc.expPluginID, extraFile))
			}
			anyPluginZIP, err := distArtifacts.OpenZIP(anyZipFn)
			require.NoError(t, err)
			expBasePluginZipFiles := make([]string, len(basePluginFiles))
			copy(expBasePluginZipFiles, basePluginFiles[:])
			if tc.expBackend {
				// Additional backend files for the "any" zip (all os+arch executables)
				// copy basePluginFiles
				expBasePluginZipFiles = append(
					expBasePluginZipFiles,
					filepath.Join(tc.expPluginID, "go_plugin_build_manifest"),
				)
				for _, exe := range tc.expExecutables {
					for _, osArch := range osArchCombos {
						suffix := osArch.String()
						if osArch.os == "windows" {
							suffix += ".exe"
						}
						expBasePluginZipFiles = append(
							expBasePluginZipFiles,
							filepath.Join(tc.expPluginID, exe+"_"+suffix),
						)
					}
				}
			}
			require.NoError(t, checkFilesExist(anyPluginZIP, expBasePluginZipFiles, checkFilesExistOptions{strict: true}))

			// plugin.json exists, check its content
			checkPluginJSON := func(zf afero.Fs) {
				pluginJSONFile, err := zf.Open(filepath.Join(tc.expPluginID, "plugin.json"))
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, pluginJSONFile.Close()) })

				var pluginJSON struct {
					ID      string `json:"id"`
					Type    string `json:"type"`
					Backend bool   `json:"backend"`
					Name    string `json:"name"`
					Info    struct {
						Version string `json:"version"`
					} `json:"info"`
				}
				require.NoError(t, json.NewDecoder(pluginJSONFile).Decode(&pluginJSON))
				require.Equal(t, tc.expPluginID, pluginJSON.ID)
				require.Equal(t, tc.expPluginVersion, pluginJSON.Info.Version)
				require.Equal(t, tc.expPluginType, pluginJSON.Type)
				require.Equal(t, tc.expBackend, pluginJSON.Backend)
			}
			checkPluginJSON(anyPluginZIP)

			// Check ZIP content for os/arch combos zips
			if tc.expBackend {
				// Base files should be present
				expBasePluginZipFiles = make([]string, len(basePluginFiles))
				copy(expBasePluginZipFiles, basePluginFiles[:])

				// Backend manifest should be present
				expBasePluginZipFiles = append(expBasePluginZipFiles, filepath.Join(tc.expPluginID, "go_plugin_build_manifest"))

				for _, osArch := range osArchCombos {
					// Create a copy of the expected base files for each zip file we check
					expPluginZipFiles := make([]string, len(expBasePluginZipFiles))
					copy(expPluginZipFiles, expBasePluginZipFiles[:])
					// Expect each backend's executable for this os/arch, and
					// nothing else: the strict check below fails on any other
					// backend binary leaking into this platform's zip.
					for _, exe := range tc.expExecutables {
						backendExeFn := exe + "_" + osArch.String()
						if osArch.os == "windows" {
							backendExeFn += ".exe"
						}
						expPluginZipFiles = append(expPluginZipFiles, filepath.Join(tc.expPluginID, backendExeFn))
					}

					// Check that all files exist
					osArchPluginZIP, err := distArtifacts.OpenZIP(osArchZipFileName(tc.expPluginID, tc.expPluginVersion, osArch))
					require.NoError(t, err)
					require.NoError(t, checkFilesExist(osArchPluginZIP, expPluginZipFiles, checkFilesExistOptions{strict: true}))

					// Check plugin.json content rather than just file existence
					checkPluginJSON(osArchPluginZIP)
				}
			}
		})
	}
}
