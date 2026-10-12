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
					require.Contains(t, step.With["cache"], "inputs.node-setup-caching != 'false'")
					require.Equal(t, "${{ inputs.node-setup-caching != 'false' }}", step.With["package-manager-cache"])
				} else {
					require.Contains(t, step.If, "inputs.node-setup-caching != 'false'")
				}
			}
			require.True(t, found, "cache step should exist")
		})
	}
}

// Run the real install-step shell without installing dependencies or starting Docker.
func TestInstallScriptsStep(t *testing.T) {
	t.Parallel()

	for _, source := range []struct {
		path string
		job  string
		step string
	}{
		{path: "actions/internal/plugins/frontend/action.yml", step: "Install dependencies"},
		{path: ".github/workflows/playwright.yml", job: "playwright-tests", step: "Install npm dependencies"},
	} {
		t.Run(source.path, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(source.path)
			require.NoError(t, err)
			var definition struct {
				Runs struct {
					Steps workflow.Steps
				}
				Jobs map[string]struct {
					Steps workflow.Steps
				}
			}
			require.NoError(t, yaml.Unmarshal(data, &definition))
			steps := definition.Runs.Steps
			if source.job != "" {
				require.Contains(t, definition.Jobs, source.job)
				steps = definition.Jobs[source.job].Steps
			}
			var installStep workflow.Step
			for _, step := range steps {
				if step.Name == source.step {
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
		})
	}
}
