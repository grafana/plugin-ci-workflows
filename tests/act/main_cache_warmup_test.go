package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/stretchr/testify/require"
)

func TestCacheWarmup_GoNodeVersions(t *testing.T) {
	// Ensure that all test cases use the same Go and Node versions as the default versions in ci.yml.
	// These versions are used to warm up the act cache, so if they are not the same, the cache will not be warmup correctly.
	// If the cache is not warmed up correctly, Go/Node will be downloaded and installed again for each test case, which:
	// - Takes longer to run the tests
	// - Causes the tests to be flaky because the cache directory used for download is shared between test cases, and is not cleaned up between test cases.
	// This test ensures that all test cases are using the same Go and Node versions as the default versions in ci.yml,
	// so that the act cache warmup is effective.

	ciWf, err := workflow.NewBaseWorkflowFromFile(filepath.Join(".github", "workflows", "ci.yml"))
	require.NoError(t, err)

	defaultGoVersion := ciWf.Env["DEFAULT_GO_VERSION"]
	defaultNodeVersion := ciWf.Env["DEFAULT_NODE_VERSION"]
	require.NotEmpty(t, defaultGoVersion, "DEFAULT_GO_VERSION is not set in ci.yml")
	require.NotEmpty(t, defaultNodeVersion, "DEFAULT_NODE_VERSION is not set in ci.yml")

	var examplesErr error
	var scannedExamples int
	walkErr := filepath.WalkDir(filepath.Join("tests"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip files and "act" directory itself
		if !d.IsDir() || d.Name() == "act" {
			return nil
		}

		nvmrcPath := filepath.Join(path, ".nvmrc")
		if _, err := os.Stat(nvmrcPath); os.IsNotExist(err) {
			return nil
		}

		scannedExamples++
		t.Logf("checking example %q", path)

		// Check Node version
		nvmrcContent, err := os.ReadFile(nvmrcPath)
		if err != nil {
			return err
		}
		if strings.TrimSpace(string(nvmrcContent)) != defaultNodeVersion {
			examplesErr = fmt.Errorf("example %q: node version in .nvmrc is not the default version: expected %q, got %q", path, defaultNodeVersion, string(nvmrcContent))
			// Do not return the error to WalkDir so we continue scanning other examples.
			return nil
		}

		// Check Go version
		// This can be in either "toolchain" or "go" line in go.mod
		// Toolchain has higher priority for the compiler version.
		goModPath := filepath.Join(path, "go.mod")
		if _, err := os.Stat(goModPath); os.IsNotExist(err) {
			return nil
		}
		f, err := os.Open(goModPath)
		if err != nil {
			return err
		}
		defer func() {
			require.NoError(t, f.Close())
		}()
		var goVersion string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			isToolchain := strings.HasPrefix(line, "toolchain")
			isGo := strings.HasPrefix(line, "go")
			if isToolchain || isGo {
				parts := strings.Split(line, " ")
				require.Len(t, parts, 2, "expected 2 parts in go.mod for 'go' or 'toolchain' directive")
				if len(parts) > 1 {
					goVersion = parts[1]
				}
				// Toolchain has higher priority, so if we find it, break immediately.
				if isToolchain {
					break
				}
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}
		if goVersion != defaultGoVersion {
			examplesErr = fmt.Errorf("example %q: go version in go.mod is not the default version: expected %q, got %q", path, defaultGoVersion, goVersion)
			// Do not return the error to WalkDir so we continue scanning other examples.
			return nil
		}
		return nil
	})
	require.NoError(t, walkErr, "error while walking directory")
	require.NotZero(t, scannedExamples, "no examples found")
	require.NoError(t, examplesErr, "some examples are not using the default Go or Node versions: this will cause issues with the act cache warmup.")
}
