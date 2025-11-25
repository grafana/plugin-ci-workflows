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

func TestSmoke(t *testing.T) {
	type testAndBuildOutput struct {
		ID         string `json:"id"`
		Version    string `json:"version"`
		HasBackend string `json:"has-backend"`
		Executable string `json:"executable"`
	}

	type cas struct {
		folder string

		exp testAndBuildOutput
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
		// "simple-frontend-yarn",
		// "simple-frontend-pnpm",
		// "simple-backend",
	} {
		t.Run(tc.folder, func(t *testing.T) {
			runner, err := act.NewRunner(t)
			require.NoError(t, err)
			runner.Verbose = true

			r, err := runner.Run(
				workflow.NewSimpleCI(
					workflow.WithPluginDirectory(filepath.Join("tests", tc.folder)),
					workflow.WithDistArtifactPrefix(tc.folder+"-"),
					workflow.WithPlaywright(false),
				),
				act.NewEmptyEventPayload(),
			)
			require.NoError(t, err)
			require.True(t, r.Success, "workflow should succeed")

			// Assert outputs
			var pluginOutput testAndBuildOutput
			err = json.Unmarshal([]byte(r.JobOutputs["test-and-build"]["plugin"]), &pluginOutput)
			require.NoError(t, err, "unmarshal plugin output JSON")
			require.Equal(t, tc.exp, pluginOutput)
		})
	}
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

func stringPointer(s string) *string {
	return &s
}
