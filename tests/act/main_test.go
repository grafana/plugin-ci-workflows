package main

import (
	"encoding/json"
	"fmt"
	"os"
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
			runner, err := act.NewRunner(t)
			require.NoError(t, err)
			// runner.Verbose = true

			wf, err := workflow.NewSimpleCI(
				workflow.WithPluginDirectory(filepath.Join("tests", tc.folder)),
				workflow.WithDistArtifactPrefix(tc.folder+"-"),
				workflow.WithPlaywright(false),
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
		})
	}
}

func TestPackage(t *testing.T) {
	runner, err := act.NewRunner(t)
	require.NoError(t, err)

	const folder = "simple-frontend"

	wf, err := workflow.NewSimpleCI(
		workflow.WithPluginDirectory(filepath.Join("tests", folder)),
		workflow.WithDistArtifactPrefix(folder+"-"),
		workflow.WithPlaywright(false),
		workflow.WithRunTruffleHog(false),
		// Mock the test-and-build job to copy pre-built dist files
		func(w *workflow.SimpleCI) {
			testAndBuild := w.CIWorkflow().Jobs["test-and-build"]
			require.NoError(t, testAndBuild.RemoveStep("setup"))
			require.NoError(t, testAndBuild.ReplaceStep("frontend", workflow.Step{
				Name: "Copy mock dist files",
				Run: workflow.Commands{
					"set -x",
					"mkdir -p ${{ github.workspace }}/${{ inputs.plugin-directory }}/dist",
					"cp -r /mockdata/dist/" + folder + "/. ${{ github.workspace }}/${{ inputs.plugin-directory }}/dist/",
					"cd ${{ github.workspace }}/${{ inputs.plugin-directory }}/dist",
					"ls -la",
				}.String(),
				Shell: "bash",
			}))
			require.NoError(t, testAndBuild.RemoveStep("backend"))
		},
	)
	require.NoError(t, err)

	r, err := runner.Run(wf, act.NewEmptyEventPayload())
	require.NoError(t, err)
	require.True(t, r.Success, "workflow should succeed")

	runID, err := r.GetTestingWorkflowRunID()
	require.NoError(t, err)
	t.Logf("gha run id is: %s", runID)
}

// TestMain sets up the test environment before running the tests.
func TestMain(m *testing.M) {
	fmt.Println("preparing test environment")

	// Go to the root of the repo
	root, err := getRepoRootAbsPath()
	if err != nil {
		panic(err)
	}
	if err := os.Chdir(root); err != nil {
		panic(err)
	}

	// Clean up old temp workflow files
	if err := act.CleanupTempWorkflowFiles(); err != nil {
		panic(err)
	}

	fmt.Println("test environment ready")

	// Run the tests
	os.Exit(m.Run())
}

// getRepoRootAbsPath returns the absolute path of the root of the git repository.
// This is the root directory for the plugin-ci-workflows repo.
// If the repo root is not found the function returns an error.
func getRepoRootAbsPath() (string, error) {
	// Start from the current working directory and look for ".git" folder.
	// If not found, move one level up and repeat until the root is reached.
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get current working directory: %w", err)
	}
	for {
		gitPath := filepath.Join(dir, ".git")
		info, err := os.Stat(gitPath)
		if err == nil && info.IsDir() {
			return dir, nil
		}
		if os.IsNotExist(err) {
			parentDir := filepath.Dir(dir)
			if parentDir == dir {
				break // Reached the root directory
			}
			dir = parentDir
			continue
		}
		return "", fmt.Errorf("stat .git directory: %w", err)
	}
	return "", fmt.Errorf(".git directory not found in any parent directories")
}
