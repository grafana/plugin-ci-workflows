package main

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/stretchr/testify/require"
)

func TestPackage(t *testing.T) {
	for _, tc := range []struct {
		folder        string
		expPluginID   string
		expBackend    bool
		expPluginType string
	}{
		{
			folder:        "simple-frontend",
			expPluginID:   "grafana-simplefrontend-panel",
			expBackend:    false,
			expPluginType: "panel",
		},
		{
			folder:        "simple-backend",
			expPluginID:   "grafana-simplebackend-datasource",
			expBackend:    true,
			expPluginType: "datasource",
		},
	} {
		t.Run(tc.folder, func(t *testing.T) {
			if !tc.expBackend {
				t.Skip()
			}
			runner, err := act.NewRunner(t)
			require.NoError(t, err)
			wf, err := workflow.NewSimpleCI(
				// CI workflow options
				workflow.WithPluginDirectoryInput(filepath.Join("tests", tc.folder)),
				workflow.WithDistArtifactPrefixInput(tc.folder+"-"),
				workflow.WithPlaywrightInput(false),
				workflow.WithRunTruffleHogInput(false),

				// Mock the test-and-build job to copy pre-built dist files
				workflow.WithMockedDist(t, tc.folder),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewEmptyEventPayload())
			require.NoError(t, err)
			require.True(t, r.Success, "workflow should succeed")

			// Inspect the artifact and assert its contents
			runID, err := r.GetTestingWorkflowRunID()
			require.NoError(t, err)
			distArtifacts, err := runner.ArtifactsStorage.GetFolder(runID, tc.folder+"-dist-artifacts")
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, distArtifacts.Close()) })

			// Expect the "any" zip file + their hashes
			anyZipFileName := tc.expPluginID + "-1.0.0.zip"
			expArtifactFiles := []string{
				anyZipFileName,
				anyZipFileName + ".md5",
				anyZipFileName + ".sha1",
			}
			if tc.expBackend {
				// Expect the os/arch backend zips + their hashes
				for _, osArch := range []string{
					"darwin_amd64",
					"darwin_arm64",
					"linux_amd64",
					"linux_arm",
					"linux_arm64",
					"windows_amd64",
				} {
					expArtifactFiles = append(
						expArtifactFiles,
						tc.expPluginID+"-1.0.0."+osArch+".zip",
						tc.expPluginID+"-1.0.0."+osArch+".zip.md5",
						tc.expPluginID+"-1.0.0."+osArch+".zip.sha1",
					)
				}
			}
			require.NoError(t, checkFilesExist(distArtifacts.Fs, expArtifactFiles, checkFilesExistOptions{strict: true}))

			// Check the checksum files (any)
			zipFileContent, err := distArtifacts.ReadFile(anyZipFileName)
			require.NoError(t, err)

			md5, err := distArtifacts.ReadFile(anyZipFileName + ".md5")
			require.NoError(t, err)
			require.Equal(t, md5Hash(zipFileContent), string(md5), "wrong md5 checksum")

			sha1, err := distArtifacts.ReadFile(anyZipFileName + ".sha1")
			require.NoError(t, err)
			require.Equal(t, sha1Hash(zipFileContent), string(sha1), "wrong sha1 checksum")

			// TODO: check checksums for os/arch backend zips

			// Check the nested plugin ZIP artifact (any)
			pluginZIP, err := distArtifacts.OpenZIP(anyZipFileName)
			require.NoError(t, err)
			expPluginZipFiles := []string{
				filepath.Join(tc.expPluginID, "CHANGELOG.md"),
				filepath.Join(tc.expPluginID, "LICENSE"),
				filepath.Join(tc.expPluginID, "module.js"),
				filepath.Join(tc.expPluginID, "module.js.map"),
				filepath.Join(tc.expPluginID, "plugin.json"),
				filepath.Join(tc.expPluginID, "README.md"),
				filepath.Join(tc.expPluginID, "img/logo.svg"),
			}
			if tc.expBackend {
				expPluginZipFiles = append(
					expPluginZipFiles,
					filepath.Join(tc.expPluginID, "go_plugin_build_manifest"),
				)
				for _, osArch := range []string{
					"darwin_amd64",
					"darwin_arm64",
					"linux_amd64",
					"linux_arm",
					"linux_arm64",
					"windows_amd64.exe",
				} {
					expPluginZipFiles = append(
						expPluginZipFiles,
						filepath.Join(tc.expPluginID, "gpx_simple_backend_"+osArch),
					)
				}
			}
			require.NoError(t, checkFilesExist(pluginZIP, expPluginZipFiles, checkFilesExistOptions{strict: true}))

			// Inspect plugin.json
			pluginJSONFile, err := pluginZIP.Open(filepath.Join(tc.expPluginID, "plugin.json"))
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
			require.Equal(t, "1.0.0", pluginJSON.Info.Version)
			require.Equal(t, tc.expPluginType, pluginJSON.Type)
			require.Equal(t, tc.expBackend, pluginJSON.Backend)

			// TODO: check the os/arch zips for their content
		})
	}
}
