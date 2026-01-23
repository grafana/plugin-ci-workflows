package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/versions"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow/ci"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var goVersionRegex = regexp.MustCompile(`^go\s+(\d+\.\d+)`)

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

		wfInputs *ci.WorkflowInputs

		// expToolingVersions contains the expected tooling-versions output, requested versions to be installed
		expToolingVersions toolingVersionsOutput

		// expSetupVersions contains the expected setup output, versions that were actually installed
		expSetupVersions setupVersions
	}

	// Read default Node/Go versions from ci.yml
	ciWf, err := workflow.NewBaseWorkflowFromFile(filepath.Join(".github", "workflows", "ci.yml"))
	require.NoError(t, err)
	defaultGoVersion := ciWf.Env["DEFAULT_GO_VERSION"]
	defaultNodeVersion := ciWf.Env["DEFAULT_NODE_VERSION"]
	require.NotEmpty(t, defaultGoVersion, "DEFAULT_GO_VERSION is not set in ci.yml")
	require.NotEmpty(t, defaultNodeVersion, "DEFAULT_NODE_VERSION is not set in ci.yml")

	// Resolve versions dynamically from simple-backend (has both .nvmrc and go.mod)
	// All test plugins use the same Node.js and Go versions.
	simpleBackendNvmrc := filepath.Join("tests", "simple-backend", ".nvmrc")
	simpleBackendGoMod := filepath.Join("tests", "simple-backend", "go.mod")

	nodeMajor, err := readNodeMajorFromNvmrc(simpleBackendNvmrc)
	require.NoError(t, err, "read .nvmrc")
	goMajorMinor, err := readGoVersionFromGoMod(simpleBackendGoMod)
	require.NoError(t, err, "read go.mod")

	// All examples should have the same Node/Go versions as the default ones in ci.yml.
	// Make sure this is true before running the tests.
	require.Equal(t, defaultGoVersion, goMajorMinor, "go version in go.mod should be the default one")
	require.Equal(t, defaultNodeVersion, nodeMajor, "node version in .nvmrc should be the default one")

	// Fetch the latest minor+patch for Node.js and Go. Done in parallel.
	// In the goroutines, use assert.* rather than require.*
	// because require.* is not safe for concurrent use.

	var latestNodeVersion, latestGoVersion string
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		latestNodeVersion, err = versions.LatestNodeVersion(nodeMajor)
		assert.NoError(t, err, "get latest Node.js version")
	}()
	go func() {
		defer wg.Done()
		latestGoVersion, err = versions.LatestGoVersion(goMajorMinor)
		assert.NoError(t, err, "get latest Go version")
	}()

	var npmVersion, yarnVersion, pnpmVersion string
	go func() {
		defer wg.Done()
		var pm string
		var err error

		pm, npmVersion, err = getPackageManagerAndVersion("tests/simple-frontend")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "npm", pm, "npm should be the package manager")
		assert.NotEmpty(t, npmVersion, "npm version should be present")

		pm, yarnVersion, err = getPackageManagerAndVersion("tests/simple-frontend-yarn")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "yarn", pm, "yarn should be the package manager")
		assert.NotEmpty(t, yarnVersion, "yarn version should be present")

		pm, pnpmVersion, err = getPackageManagerAndVersion("tests/simple-frontend-pnpm")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "pnpm", pm, "pnpm should be the package manager")
		assert.NotEmpty(t, pnpmVersion, "pnpm version should be present")
	}()
	wg.Wait()

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
				nodeVersion:               latestNodeVersion,
				nodePackageManagerName:    "npm",
				nodePackageManagerVersion: npmVersion,
				goVersion:                 latestGoVersion,
			},
		},
		{
			// same as simple-frontend, but with yarn
			name:   "simple-frontend-yarn",
			folder: "simple-frontend-yarn",
			expToolingVersions: toolingVersionsOutput{
				NodeVersion:     "",
				GoVersion:       defaultGoVersion,
				NodeVersionFile: filepath.Join("tests", "simple-frontend-yarn", ".nvmrc"),
				GoVersionFile:   "",
			},
			expSetupVersions: setupVersions{
				nodeVersion:               latestNodeVersion,
				nodePackageManagerName:    "yarn",
				nodePackageManagerVersion: yarnVersion,
				goVersion:                 latestGoVersion,
			},
		},
		{
			// same as simple-frontend, but with pnpm
			name:   "simple-frontend-pnpm",
			folder: "simple-frontend-pnpm",
			expToolingVersions: toolingVersionsOutput{
				NodeVersion:     "",
				GoVersion:       defaultGoVersion,
				NodeVersionFile: filepath.Join("tests", "simple-frontend-pnpm", ".nvmrc"),
				GoVersionFile:   "",
			},
			expSetupVersions: setupVersions{
				nodeVersion:               latestNodeVersion,
				nodePackageManagerName:    "pnpm",
				nodePackageManagerVersion: pnpmVersion,
				goVersion:                 latestGoVersion,
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
				nodeVersion:               latestNodeVersion,
				nodePackageManagerName:    "npm",
				nodePackageManagerVersion: npmVersion,
				goVersion:                 latestGoVersion,
			},
		},
		{
			// explicit versions via inputs has higher priority than the files or default ones
			// use simple-backend as folder because it has both .nvmrc and go.mod
			name:   "explicit versions via inputs",
			folder: "simple-backend",
			wfInputs: &ci.WorkflowInputs{
				GoVersion:   workflow.Input("1.25.5"),
				NodeVersion: workflow.Input("24.12.0"),
			},
			expToolingVersions: toolingVersionsOutput{
				NodeVersion:     "24.12.0",
				GoVersion:       "1.25.5",
				NodeVersionFile: "",
				GoVersionFile:   "",
			},
			expSetupVersions: setupVersions{
				nodeVersion:               "v24.12.0",
				nodePackageManagerName:    "npm",
				nodePackageManagerVersion: npmVersion,
				goVersion:                 "1.25.5",
			},
		},
		{
			// no version files and no inputs, so fall back to default versions
			name:   "no version files and no inputs",
			folder: "simple-frontend-no-nvmrc",
			expToolingVersions: toolingVersionsOutput{
				NodeVersion:     defaultNodeVersion,
				GoVersion:       defaultGoVersion,
				NodeVersionFile: "",
				GoVersionFile:   "",
			},
			expSetupVersions: setupVersions{
				nodeVersion:               latestNodeVersion,
				nodePackageManagerName:    "npm",
				nodePackageManagerVersion: npmVersion,
				goVersion:                 latestGoVersion,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)

			var wfInputs ci.WorkflowInputs
			if tc.wfInputs != nil {
				// Apply custom inputs if provided in the test case
				wfInputs = *tc.wfInputs
			}
			// Override the plugin directory input with the actual plugin folder
			wfInputs.PluginDirectory = workflow.Input(filepath.Join("tests", tc.folder))

			wf, err := ci.NewWorkflow(
				ci.WithWorkflowInputs(wfInputs),
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

func getPackageManagerAndVersion(pluginFolder string) (string, string, error) {
	packageJsonPath := filepath.Join(pluginFolder, "package.json")
	data, err := os.Open(packageJsonPath)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = data.Close() }()

	var packageJson struct {
		PackageManager string `json:"packageManager"`
	}
	err = json.NewDecoder(data).Decode(&packageJson)
	if err != nil {
		return "", "", err
	}
	pmParts := strings.Split(packageJson.PackageManager, "@")
	if len(pmParts) != 2 {
		return "", "", errors.New("package manager should be in the format <name>@<version>")
	}
	pmName := pmParts[0]
	pmVersion := pmParts[1]
	return pmName, pmVersion, nil
}
