package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/versions"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow/ci"
	"github.com/stretchr/testify/require"
)

const (
	// defaultGoVersion is the default Go version used by ci.yml when no go.mod is present.
	defaultGoVersion = "1.25"
)

// readNodeMajorFromNvmrc reads the major Node.js version from an .nvmrc file.
// It returns the major version string (e.g., "24").
func readNodeMajorFromNvmrc(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// readGoVersionFromGoMod reads the Go version from a go.mod file.
// It returns the version string (e.g., "1.25").
func readGoVersionFromGoMod(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	// Match "go X.Y" or "go X.Y.Z" lines
	goVersionRegex := regexp.MustCompile(`^go\s+(\d+\.\d+)`)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := goVersionRegex.FindStringSubmatch(line); len(matches) == 2 {
			return matches[1], nil
		}
	}
	return "", scanner.Err()
}

func TestToolingVersions(t *testing.T) {
	type toolingVersionsOutput struct {
		NodeVersion     string `json:"nodeVersion"`
		GoVersion       string `json:"goVersion"`
		NodeVersionFile string `json:"nodeVersionFile"`
		GoVersionFile   string `json:"goVersionFile"`
	}
	type setupVersions struct {
		nodeVersion               string
		nodePackageManagerName    string
		nodePackageManagerVersion string
		goVersion                 string
	}
	type tc struct {
		name   string
		folder string

		// tooling-versions output, requested versions to be installed
		expToolingVersions toolingVersionsOutput

		// setup output, versions that were actually installed
		expSetupVersions setupVersions
	}

	// Resolve versions dynamically from simple-backend (has both .nvmrc and go.mod)
	// All test plugins use the same Node.js and Go versions.
	simpleBackendNvmrc := filepath.Join("tests", "simple-backend", ".nvmrc")
	simpleBackendGoMod := filepath.Join("tests", "simple-backend", "go.mod")

	nodeMajor, err := readNodeMajorFromNvmrc(simpleBackendNvmrc)
	require.NoError(t, err, "read .nvmrc")
	goMajorMinor, err := readGoVersionFromGoMod(simpleBackendGoMod)
	require.NoError(t, err, "read go.mod")

	latestNode, err := versions.LatestNodeVersion(nodeMajor)
	require.NoError(t, err, "get latest Node.js version")
	latestGo, err := versions.LatestGoVersion(goMajorMinor)
	require.NoError(t, err, "get latest Go version")

	for _, tc := range []tc{
		{
			// simple-frontend has .nvmrc and no backend, so Go version is the default one.
			name:   "simple-frontend",
			folder: "simple-frontend",
			expToolingVersions: toolingVersionsOutput{
				NodeVersion:     "",
				GoVersion:       defaultGoVersion,
				NodeVersionFile: filepath.Join("tests", "simple-frontend", ".nvmrc"),
				GoVersionFile:   "",
			},
			expSetupVersions: setupVersions{
				nodeVersion:               latestNode,
				nodePackageManagerName:    "npm",
				nodePackageManagerVersion: "10.9.0",
				goVersion:                 latestGo,
			},
		},
		{
			// simple-backend has both .nvmrc and go.mod, so both versions should be the ones from the files.
			name:   "simple-backend",
			folder: "simple-backend",
			expToolingVersions: toolingVersionsOutput{
				NodeVersion:     "",
				GoVersion:       "",
				NodeVersionFile: simpleBackendNvmrc,
				GoVersionFile:   simpleBackendGoMod,
			},
			expSetupVersions: setupVersions{
				nodeVersion:               latestNode,
				nodePackageManagerName:    "npm",
				nodePackageManagerVersion: "10.9.0",
				goVersion:                 latestGo,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)

			wf, err := ci.NewWorkflow(
				ci.WithWorkflowInputs(ci.WorkflowInputs{
					PluginDirectory:     workflow.Input(filepath.Join("tests", tc.folder)),
					DistArtifactsPrefix: workflow.Input(tc.folder + "-"),
					RunPlaywright:       workflow.Input(false),
				}),
				ci.MutateCIWorkflow().With(
					workflow.WithOnlyOneJob(t, "test-and-build", true),
					workflow.WithRemoveAllStepsAfter(t, "test-and-build", "setup"),
				),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewPushEventPayload("main"))
			require.NoError(t, err)
			require.True(t, r.Success, "workflow should succeed")

			// Assert tooling-versions output (requested versions to be installed)
			out, ok := r.Outputs.Get("test-and-build", "tooling-versions", "result")
			require.True(t, ok, "output result should be present")
			var requestedToolingVersions toolingVersionsOutput
			require.NoError(t, json.Unmarshal([]byte(out), &requestedToolingVersions))
			require.Equal(t, tc.expToolingVersions.NodeVersion, requestedToolingVersions.NodeVersion, "requested node version should be the expected one")
			require.Equal(t, tc.expToolingVersions.GoVersion, requestedToolingVersions.GoVersion, "requested go version should be the expected one")
			require.Equal(t, tc.expToolingVersions.NodeVersionFile, requestedToolingVersions.NodeVersionFile, "requested node version file should be the expected one")
			require.Equal(t, tc.expToolingVersions.GoVersionFile, requestedToolingVersions.GoVersionFile, "requested go version file should be the expected one")

			// Assert setup output (versions that were actually installed)
			setupNodeVersion, ok := r.Outputs.Get("test-and-build", "setup", "node-version")
			require.True(t, ok, "node version should be present")
			setupNodePackageManagerName, ok := r.Outputs.Get("test-and-build", "setup", "node-package-manager-name")
			require.True(t, ok, "node package manager name should be present")
			setupNodePackageManagerVersion, ok := r.Outputs.Get("test-and-build", "setup", "node-package-manager-version")
			require.True(t, ok, "node package manager version should be present")
			setupGoVersion, ok := r.Outputs.Get("test-and-build", "setup", "go-version")
			require.True(t, ok, "go version should be present")
			require.Equal(t, tc.expSetupVersions.nodeVersion, setupNodeVersion, "installed node version should be the expected one")
			require.Equal(t, tc.expSetupVersions.nodePackageManagerName, setupNodePackageManagerName, "installed node package manager name should be the expected one")
			require.Equal(t, tc.expSetupVersions.nodePackageManagerVersion, setupNodePackageManagerVersion, "installed node package manager version should be the expected one")
			require.Equal(t, tc.expSetupVersions.goVersion, setupGoVersion, "installed go version should be the expected one")

		})
	}
}
