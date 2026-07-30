package main

import (
	"archive/zip"
	"encoding/json"
	"os"
	"os/exec"
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
					suffix := osArch.String()
					if osArch.os == "windows" {
						suffix += ".exe"
					}
					expBasePluginZipFiles = append(
						expBasePluginZipFiles,
						filepath.Join(tc.expPluginID, "gpx_simple_backend_"+suffix),
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
					// Create a copy of the expected base files for each zip file we check
					expPluginZipFiles := make([]string, len(expBasePluginZipFiles))
					copy(expPluginZipFiles, expBasePluginZipFiles[:])
					backendExeFn := "gpx_simple_backend_" + osArch.String()
					if osArch.os == "windows" {
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

func TestPackageScriptMultipleBackendFamilies(t *testing.T) {
	t.Parallel()

	const (
		pluginID      = "grafana-multibackend-app"
		pluginVersion = "1.0.0"
	)

	tmp := t.TempDir()
	dist := filepath.Join(tmp, "dist")
	out := filepath.Join(tmp, "dist-artifacts")
	require.NoError(t, os.MkdirAll(out, 0o755))

	fixtureFiles := map[string]string{
		"plugin.json": `{
			"id": "grafana-multibackend-app",
			"type": "app",
			"executable": "bin/gpx_root",
			"info": {"version": "1.0.0"}
		}`,
		"module.js":                               "root frontend",
		"go_plugin_build_manifest":                "root manifest",
		"bin/gpx_root_linux_amd64":                "root linux executable",
		"bin/gpx_root_windows_amd64.exe":          "root windows executable",
		"resources/config.json":                   `{"preserve": true}`,
		"datasource/plugin.json":                  `{"id":"grafana-nested-datasource","type":"datasource","backend":true,"executable":"gpx_nested","info":{"version":"1.0.0"}}`,
		"datasource/module.js":                    "nested frontend",
		"datasource/go_plugin_build_manifest":     "nested manifest",
		"datasource/gpx_nested_linux_amd64":       "nested linux executable",
		"datasource/gpx_nested_windows_amd64.exe": "nested windows executable",
	}
	writePackageFixtureFiles(t, dist, fixtureFiles)

	output, err := runPackageScript(dist, out)
	require.NoError(t, err, string(output))
	output, err = runPackageScript("--universal", dist, out)
	require.NoError(t, err, string(output))

	baseFiles := []string{
		filepath.Join(pluginID, "plugin.json"),
		filepath.Join(pluginID, "module.js"),
		filepath.Join(pluginID, "go_plugin_build_manifest"),
		filepath.Join(pluginID, "resources/config.json"),
		filepath.Join(pluginID, "datasource/plugin.json"),
		filepath.Join(pluginID, "datasource/module.js"),
		filepath.Join(pluginID, "datasource/go_plugin_build_manifest"),
	}
	backendFiles := map[string][]string{
		"linux_amd64": {
			filepath.Join(pluginID, "bin/gpx_root_linux_amd64"),
			filepath.Join(pluginID, "datasource/gpx_nested_linux_amd64"),
		},
		"windows_amd64": {
			filepath.Join(pluginID, "bin/gpx_root_windows_amd64.exe"),
			filepath.Join(pluginID, "datasource/gpx_nested_windows_amd64.exe"),
		},
	}

	for _, tc := range []struct {
		name          string
		zipName       string
		expectedFiles []string
	}{
		{
			name:    "universal",
			zipName: anyZipFileName(pluginID, pluginVersion),
			expectedFiles: append(
				append(append([]string{}, baseFiles...), backendFiles["linux_amd64"]...),
				backendFiles["windows_amd64"]...,
			),
		},
		{
			name:          "linux amd64",
			zipName:       pluginID + "-" + pluginVersion + ".linux_amd64.zip",
			expectedFiles: append(append([]string{}, baseFiles...), backendFiles["linux_amd64"]...),
		},
		{
			name:          "windows amd64",
			zipName:       pluginID + "-" + pluginVersion + ".windows_amd64.zip",
			expectedFiles: append(append([]string{}, baseFiles...), backendFiles["windows_amd64"]...),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			zr, err := zip.OpenReader(filepath.Join(out, tc.zipName))
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, zr.Close()) })

			actualFiles := make([]string, 0, len(zr.File))
			for _, file := range zr.File {
				if !file.FileInfo().IsDir() {
					actualFiles = append(actualFiles, file.Name)
				}
			}
			require.ElementsMatch(t, tc.expectedFiles, actualFiles)
		})
	}
}

func TestPackageScriptRejectsMismatchedBackendPlatforms(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	dist := filepath.Join(tmp, "dist")
	out := filepath.Join(tmp, "dist-artifacts")
	require.NoError(t, os.MkdirAll(out, 0o755))

	writePackageFixtureFiles(t, dist, map[string]string{
		"plugin.json":                       `{"id":"grafana-multibackend-app","type":"app","executable":"gpx_root","info":{"version":"1.0.0"}}`,
		"gpx_root_linux_amd64":              "root linux executable",
		"gpx_root_windows_amd64.exe":        "root windows executable",
		"datasource/plugin.json":            `{"id":"grafana-nested-datasource","type":"datasource","backend":true,"executable":"gpx_nested","info":{"version":"1.0.0"}}`,
		"datasource/gpx_nested_linux_amd64": "nested linux executable",
	})

	output, err := runPackageScript(dist, out)
	require.Error(t, err)
	require.Contains(t, string(output), "Executable 'datasource/gpx_nested' does not provide windows_amd64, aborting.")

	entries, err := os.ReadDir(out)
	require.NoError(t, err)
	require.Empty(t, entries, "validation should fail before creating tailored archives")
}

func writePackageFixtureFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()

	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o755))
	}
}

func runPackageScript(args ...string) ([]byte, error) {
	cmd := exec.Command("bash", append([]string{"actions/internal/plugins/package/package.sh"}, args...)...)
	cmd.Env = append(os.Environ(), "GRAFANA_ACCESS_POLICY_TOKEN=", "SIGNATURE_TYPE=grafana")
	return cmd.CombinedOutput()
}
