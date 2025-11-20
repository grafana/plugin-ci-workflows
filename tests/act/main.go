package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
)

const testWorkflowFile = `
name: act

on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main

permissions: {}

jobs:
  act:
    name: act
    uses: grafana/plugin-ci-workflows/.github/workflows/ci.yml@main
    permissions:
      contents: read
      id-token: write
    with:
      plugin-version-suffix: ${{ github.event_name == 'pull_request' && github.event.pull_request.head.sha || '' }}
      run-playwright: false
      testing: true
      testing-act: true
      plugin-directory: tests/simple-frontend
      dist-artifacts-prefix: simple-frontend-
    secrets:
      GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
`

func _main() error {
	// Go to the root of the repo
	root, err := getRepoRootAbsPath()
	if err != nil {
		return fmt.Errorf("determine repo root: %w", err)
	}
	if err := os.Chdir(root); err != nil {
		return fmt.Errorf("chdir to repo root: %w", err)
	}

	// Create a temporary workflow file to run the test
	fn := "act-" + uuid.NewString() + ".yml"
	fn = filepath.Join(".github", "workflows", fn)
	if err := os.WriteFile(fn, []byte(testWorkflowFile), 0o644); err != nil {
		return fmt.Errorf("write test workflow file: %w", err)
	}
	defer os.Remove(fn)

	runner, err := act.NewRunner()
	if err != nil {
		return err
	}
	if err := runner.Run(fn); err != nil {
		return err
	}
	return nil
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

func main() {
	if err := _main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
