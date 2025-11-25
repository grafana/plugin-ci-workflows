package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmoke(t *testing.T) {
	type cas struct {
		folder string

		expPluginID      string
		expPluginVersion string
		expHasBackend    bool
		expExecutable    *string
	}

	for _, tc := range []cas{
		{
			folder:           "simple-frontend",
			expPluginID:      "grafana-simplefrontend-panel",
			expPluginVersion: "1.0.0",
			expHasBackend:    false,
			expExecutable:    nil,
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
			jobOutput := r.JobOutputs["test-and-build"]

			pluginOutputRaw := jobOutput["plugin"]
			t.Log(pluginOutputRaw)
			var pluginOutput map[string]any
			err = json.Unmarshal([]byte(pluginOutputRaw), &pluginOutput)
			require.NoError(t, err, "unmarshal plugin output JSON")

			assert.Equal(t, tc.expPluginID, pluginOutput["id"])
			assert.Equal(t, tc.expPluginVersion, pluginOutput["version"])
			backend, ok := pluginOutput["backend"].(string)
			require.True(t, ok, "backend field should be a boolean string")
			assert.Equal(t, tc.expHasBackend, backend == "true")
			if tc.expExecutable != nil {
				assert.Equal(t, backend, *tc.expExecutable)
			} else {
				assert.Nil(t, backend)
			}
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
