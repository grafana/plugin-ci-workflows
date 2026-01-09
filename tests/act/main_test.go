package main

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/spf13/afero"
)

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

	// Read ci.yml to get the default tooling versions, so we can warm up the cache
	ciWf, err := workflow.NewBaseWorkflowFromFile(filepath.Join(".github", "workflows", "ci.yml"))
	if err != nil {
		panic(err)
	}

	// Warm up act-toolcache volume, otherwise we get weird errors
	// when running the "setup/*" actions in parallel tests since they
	// all share the same act-toolcache volume.
	runner, err := act.NewRunner(&testing.T{}, act.WithName("toolcache-warmup"))
	if err != nil {
		panic(err)
	}
	cacheWarmupWf := workflow.BaseWorkflow{
		Name: "Act tool cache warm up",
		On: workflow.On{
			Push: workflow.OnPush{
				Branches: []string{"main"},
			},
		},
		Jobs: map[string]*workflow.Job{
			"warmup": {
				Name:   "Warm up tool cache",
				RunsOn: "ubuntu-arm64-small",
				Steps: []workflow.Step{
					{
						Name: "Warm up tooling",
						Uses: "grafana/plugin-ci-workflows/actions/internal/plugins/setup@main",
						With: map[string]any{
							"go-version":            ciWf.Env["DEFAULT_GO_VERSION"],
							"node-version":          ciWf.Env["DEFAULT_NODE_VERSION"],
							"golangci-lint-version": ciWf.Env["DEFAULT_GOLANGCI_LINT_VERSION"],
							"mage-version":          ciWf.Env["DEFAULT_MAGE_VERSION"],
							"act-cache-warmup":      "true",
						},
					},
					{
						Name: "Warm up Trufflehog",
						Uses: "grafana/plugin-ci-workflows/actions/internal/plugins/trufflehog@main",
						With: map[string]any{
							"trufflehog-version": ciWf.Env["DEFAULT_TRUFFLEHOG_VERSION"],
							"setup-only":         "true",
						},
					},
				},
			},
		},
	}
	r, err := runner.Run(
		workflow.NewTestingWorkflow("toolcache-warmup", cacheWarmupWf),
		act.NewPushEventPayload("main"),
	)
	if err != nil {
		panic(fmt.Errorf("warm up act toolcache: %w", err))
	}
	if !r.Success {
		panic("warm up act toolcache: workflow failed")
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

// Utilities for tests

// checkFilesExist checks that all expected files exist in the given afero.Fs.
// and they are not empty.
// It accepts an optional checkFilesExistOptions to customize the behavior.
// If more than one option is provided, an error is returned.
// If strict mode is enabled via the options, the function will also return
// an error if any unexpected files are found.
// Otherwise, unexpected files are allowed and won't cause the assertion to fail.
// The caller should assert on the returned error via testify, for example:
//
// ```go
//
//	err := checkFilesExist(fs, expectedFiles, checkFilesExistOptions{strict: true})
//	require.NoError(t, err)
//
// ```
func checkFilesExist(fs afero.Fs, exp []string, opt ...checkFilesExistOptions) error {
	var o checkFilesExistOptions
	if len(opt) == 1 {
		o = opt[0]
	} else if len(opt) == 0 {
		o = checkFilesExistOptions{}
	} else {
		return fmt.Errorf("only one option allowed, got %d", len(opt))
	}

	var finalErr error
	expectedFiles := aferoFilesMap(exp)
	if err := afero.Walk(fs, "/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip directory entries, only care about files
			return nil
		}
		if _, ok := expectedFiles[path]; ok {
			if info.Size() == 0 {
				finalErr = errors.Join(finalErr, fmt.Errorf("expected file %q is empty", path))
				return nil
			}
			delete(expectedFiles, path)
		} else if o.strict {
			finalErr = errors.Join(finalErr, fmt.Errorf("unexpected file %q found", path))
			return nil
		}
		return nil
	}); err != nil {
		return err
	}
	if len(expectedFiles) > 0 {
		finalErr = errors.Join(finalErr, fmt.Errorf("expected files not found: %v", expectedFiles))
	}
	return finalErr
}

// checkFilesExistOptions defines options for the checkFilesExist function.
type checkFilesExistOptions struct {
	// strict indicates whether to fail if unexpected files are found.
	// If true, unexpected files will cause an error.
	// If false, unexpected files are ignored and won't cause the assertion to fail.
	strict bool
}

// checkFilesDontExist checks that none of the files in the notExp slice exist in the given afero.Fs.
// If any of the files exist, an error is returned listing the unexpected files found.
// If none of the files exist, nil is returned.
// The caller should assert on the returned error via testify, for example:
//
// ```go
//
//	err := checkFilesDontExist(fs, unexpectedFiles)
//	require.NoError(t, err)
//
// ```
func checkFilesDontExist(fs afero.Fs, notExp []string) error {
	unexpectedFiles := aferoFilesMap(notExp)
	var finalErr error
	for fn := range unexpectedFiles {
		exists, err := afero.Exists(fs, fn)
		if err != nil {
			return fmt.Errorf("check existence of file %q: %w", fn, err)
		}
		if exists {
			finalErr = errors.Join(finalErr, fmt.Errorf("unexpected file %q found", fn))
		}
	}
	return finalErr
}

// aferoFilesMap converts a slice of file paths into a map for easy lookup.
func aferoFilesMap(files []string) map[string]struct{} {
	r := make(map[string]struct{}, len(files))
	for _, f := range files {
		// Add leading slash for consistency with afero.Walk paths
		if !strings.HasPrefix(f, "/") {
			f = "/" + f
		}
		r[f] = struct{}{}
	}
	return r
}

// md5Hash returns the MD5 hash of the given byte slice as a hexadecimal string.
func md5Hash(b []byte) string {
	h := md5.Sum(b)
	return hex.EncodeToString(h[:])
}

// sha1Hash returns the SHA1 hash of the given byte slice as a hexadecimal string.
func sha1Hash(b []byte) string {
	h := sha1.Sum(b)
	return hex.EncodeToString(h[:])
}

// anyZipFileName returns the file name for the "any" ZIP file of the given plugin ID and version.
func anyZipFileName(pluginID, version string) string {
	return pluginID + "-" + version + ".zip"
}

// osArchZipFileName returns the file name for the OS/Arch specific ZIP file
func osArchZipFileName(pluginID, version, osArch string) string {
	return pluginID + "-" + version + "." + osArch + ".zip"
}

// getGitCommitSHA returns the current git commit SHA of the repository in the current working directory.
// git must be installed.
func getGitCommitSHA() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// osArchCombos defines the supported OS/Arch combinations for plugin packaging.
var osArchCombos = [...]string{
	"darwin_amd64",
	"darwin_arm64",
	"linux_amd64",
	"linux_arm",
	"linux_arm64",
	"windows_amd64",
}
