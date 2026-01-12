package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
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
		},
	} {
		t.Run(tc.folder, func(t *testing.T) {
			t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)
			wf, err := ci.NewWorkflow(
				// CI workflow options
				ci.WithPluginDirectoryInput(filepath.Join("tests", tc.folder)),
				ci.WithDistArtifactPrefixInput(tc.folder+"-"),
				ci.WithPlaywrightInput(false),
				ci.WithRunTruffleHogInput(false),

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
			basePluginFiles := [...]string{
				filepath.Join(tc.expPluginID, "CHANGELOG.md"),
				filepath.Join(tc.expPluginID, "LICENSE"),
				filepath.Join(tc.expPluginID, "module.js"),
				filepath.Join(tc.expPluginID, "module.js.map"),
				filepath.Join(tc.expPluginID, "plugin.json"),
				filepath.Join(tc.expPluginID, "README.md"),
				filepath.Join(tc.expPluginID, "img/logo.svg"),
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
				for _, osArch := range osArchCombos {
					if strings.Contains(osArch, "windows") {
						osArch += ".exe"
					}
					expBasePluginZipFiles = append(
						expBasePluginZipFiles,
						filepath.Join(tc.expPluginID, "gpx_simple_backend_"+osArch),
					)
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
					t.Logf("checking plugin ZIP for %s", osArch)

					// Create a copy of the expected base files for each zip file we check
					expPluginZipFiles := make([]string, len(expBasePluginZipFiles))
					copy(expPluginZipFiles, expBasePluginZipFiles[:])
					backendExeFn := "gpx_simple_backend_" + osArch
					if strings.Contains(osArch, "windows") {
						backendExeFn += ".exe"
					}
					// Expect the backend executable for this os/arch
					expPluginZipFiles = append(expPluginZipFiles, filepath.Join(tc.expPluginID, backendExeFn))

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
