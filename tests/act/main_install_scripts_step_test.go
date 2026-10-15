package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/stretchr/testify/require"
)

func TestSetupDependencyCacheToggle(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("actions/internal/plugins/setup/action.yml")
	require.NoError(t, err)
	var setup struct {
		Inputs map[string]struct {
			Default string
		}
		Runs struct {
			Steps workflow.Steps
		}
	}
	require.NoError(t, yaml.Unmarshal(data, &setup))
	require.Contains(t, setup.Inputs, "node-setup-caching")
	require.Equal(t, "true", setup.Inputs["node-setup-caching"].Default)

	for _, name := range []string{"Node", "Locate the pnpm store", "Cache pnpm store"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var found bool
			for _, step := range setup.Runs.Steps {
				if step.Name != name {
					continue
				}
				found = true
				if name == "Node" {
					require.Contains(t, step.With["cache"], "inputs.node-setup-caching == 'true'")
					require.Equal(t, false, step.With["package-manager-cache"])
				} else {
					require.Contains(t, step.If, "inputs.node-setup-caching == 'true'")
				}
			}
			require.True(t, found, "cache step should exist")
		})
	}
}

// Run the real install-step shell without installing dependencies or starting Docker.
func TestInstallScriptsStep(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("actions/internal/plugins/frontend/action.yml")
	require.NoError(t, err)
	var definition struct {
		Runs struct {
			Steps workflow.Steps
		}
	}
	require.NoError(t, yaml.Unmarshal(data, &definition))
	steps := definition.Runs.Steps
	var installStep workflow.Step
	for _, step := range steps {
		if step.Name == "Install dependencies" {
			installStep = step
			break
		}
	}
	require.NotEmpty(t, installStep.Run)

	for _, tc := range []struct {
		name       string
		policy     string
		command    string
		wantOutput string
		wantError  bool
	}{
		{name: "none", policy: "false", command: "pm-probe install --immutable --mode=skip-build", wantOutput: "YARN_ENABLE_SCRIPTS=false\n[install]\n[--immutable]\n[--mode=skip-build]\n"},
		{name: "allowlist", policy: "false", command: "pm-probe install --immutable", wantOutput: "YARN_ENABLE_SCRIPTS=false\n[install]\n[--immutable]\n"},
		{name: "all", policy: "true", command: "pm-probe install --immutable", wantOutput: "YARN_ENABLE_SCRIPTS=true\n[install]\n[--immutable]\n"},
		{name: "not yarn", command: "pm-probe ci --ignore-scripts", wantOutput: "[ci]\n[--ignore-scripts]\n"},
		{name: "empty command", wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			binDir := t.TempDir()
			probe := "#!/bin/bash\n/usr/bin/env | /usr/bin/sort | /usr/bin/grep '^YARN_' || true\nprintf '[%s]\\n' \"$@\"\n"
			require.NoError(t, os.WriteFile(filepath.Join(binDir, "pm-probe"), []byte(probe), 0o755))

			scriptPath := filepath.Join(binDir, "install.sh")
			require.NoError(t, os.WriteFile(scriptPath, []byte(installStep.Run), 0o600))
			cmd := exec.Command("bash", "--noprofile", "--norc", "-e", "-o", "pipefail", scriptPath)
			cmd.Env = []string{"PATH=" + binDir + ":/usr/bin:/bin", "INSTALL_CMD=" + tc.command}
			var policyKey string
			for key, expression := range installStep.Env {
				if strings.Contains(expression, "yarn-enable-scripts") {
					policyKey = key
					cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, tc.policy))
				}
			}
			require.NotEmpty(t, policyKey)
			output, err := cmd.CombinedOutput()
			if tc.wantError {
				require.Error(t, err)
				require.Contains(t, string(output), "::error::")
			} else {
				require.NoError(t, err, string(output))
				// Any extra YARN_* helper variable is interpreted as an unknown setting by Yarn.
				require.Equal(t, tc.wantOutput, string(output))
			}
		})
	}
}

func stepIndex(steps workflow.Steps, match func(workflow.Step) bool) int {
	for i, step := range steps {
		if match(step) {
			return i
		}
	}
	return -1
}

// The Playwright job must install dependencies through the frontend action in install-only mode.
func TestPlaywrightUsesFrontendInstallOnly(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(".github/workflows/playwright.yml")
	require.NoError(t, err)
	var wf struct {
		Jobs map[string]struct {
			Steps workflow.Steps
		}
	}
	require.NoError(t, yaml.Unmarshal(data, &wf))
	require.Contains(t, wf.Jobs, "playwright-tests")
	steps := wf.Jobs["playwright-tests"].Steps

	// The ref is rewritten to a release tag by release automation; only match the action path.
	install := stepIndex(steps, func(s workflow.Step) bool {
		path, _, _ := strings.Cut(s.Uses, "@")
		return path == "grafana/plugin-ci-workflows/actions/internal/plugins/frontend"
	})
	require.NotEqual(t, -1, install, "frontend action step should exist")
	setup := stepIndex(steps, func(s workflow.Step) bool { return s.ID == "setup" })
	version := stepIndex(steps, func(s workflow.Step) bool { return s.Name == "Get Playwright version" })
	require.NotEqual(t, -1, setup)
	require.NotEqual(t, -1, version)
	require.Less(t, setup, install)
	require.Less(t, install, version)

	with := steps[install].With
	for key, want := range map[string]any{
		"install-only":        "true",
		"install-cmd":         "${{ steps.setup.outputs.node-package-manager-install-cmd }}",
		"yarn-enable-scripts": "${{ steps.setup.outputs.node-package-manager-yarn-enable-scripts }}",
		"npm-registry-auth":   "${{ inputs.npm-registry-auth }}",
	} {
		require.Equal(t, want, with[key], key)
	}
	require.NotContains(t, with, "secrets")
}

func TestFrontendActionInstallOnly(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("actions/internal/plugins/frontend/action.yml")
	require.NoError(t, err)
	var action struct {
		Inputs map[string]struct {
			Default string
		}
		Runs struct {
			Steps workflow.Steps
		}
	}
	require.NoError(t, yaml.Unmarshal(data, &action))
	require.Contains(t, action.Inputs, "install-only")
	require.Equal(t, "false", action.Inputs["install-only"].Default)

	for _, tc := range []struct {
		step  string
		gated bool
	}{
		{step: "Detect package manager", gated: true},
		{step: "Lint", gated: true},
		{step: "Typecheck", gated: true},
		{step: "Test", gated: true},
		{step: "Build", gated: true},
		{step: "Login to Google Cloud"},
		{step: "NPM registry auth"},
		{step: "Install dependencies"},
	} {
		t.Run(tc.step, func(t *testing.T) {
			t.Parallel()

			i := stepIndex(action.Runs.Steps, func(s workflow.Step) bool { return s.Name == tc.step })
			require.NotEqual(t, -1, i, "step should exist")
			if tc.gated {
				require.Contains(t, action.Runs.Steps[i].If, "inputs.install-only != 'true'")
			} else {
				require.NotContains(t, action.Runs.Steps[i].If, "install-only")
			}
		})
	}
}
