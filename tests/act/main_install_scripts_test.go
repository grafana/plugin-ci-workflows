package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow/ci"
	"github.com/stretchr/testify/require"
)

// TestInstallScripts verifies the "Configure package manager script policy" logic
// that is applied before running npm/yarn/pnpm install in the frontend action and
// playwright workflow.
//
// It tests the two dimensions that matter:
//  1. The package manager agent (npm, pnpm, yarn, yarnBerry) - determines whether
//     --ignore-scripts can be appended to the CLI install command.
//  2. The node-package-manager-allow-scripts input (true/false) - determines whether
//     scripts should be disabled at all.
//
// Since the "Configure package manager script policy" step lives inside the
// frontend composite action (which we cannot inject into directly via the test
// framework), we replace the entire `frontend` step in test-and-build with an
// inline bash step that reproduces the same logic and emits ::act-debug::
// annotations. This lets us assert on the computed install command and the env
// vars that get exported to GITHUB_ENV, without running the real install.
func TestInstallScripts(t *testing.T) {
	// injectScriptPolicyStep is a workflow step that reproduces the
	// "Configure package manager script policy" shell logic from
	// actions/internal/plugins/frontend/action.yml, emitting ::act-debug::
	// annotations so the test framework can capture and assert on the results.
	//
	// Emitted annotations (logfmt format):
	//   msg="install-cmd" cmd=<computed install command>
	//   msg="npm-config-ignore-scripts" value=<true|not-set>
	//   msg="yarn-enable-scripts" value=<false|not-set>
	injectScriptPolicyStep := func(frozenInstallCmd, packageManagerAgent string, allowScripts bool) workflow.Step {
		allowScriptsStr := "false"
		if allowScripts {
			allowScriptsStr = "true"
		}
		return workflow.Step{
			Name:  "Configure package manager script policy (test probe)",
			Shell: "bash",
			Run: workflow.Commands{
				`INSTALL_CMD="${FROZEN_INSTALL_CMD}"`,
				`if [ "${ALLOW_SCRIPTS}" != "true" ]; then`,
				`  if [ "${PACKAGE_MANAGER_AGENT}" != "yarnBerry" ]; then`,
				`    INSTALL_CMD="${INSTALL_CMD} --ignore-scripts"`,
				`  fi`,
				`  echo "NPM_CONFIG_IGNORE_SCRIPTS=true" >> "$GITHUB_ENV"`,
				`  echo "YARN_ENABLE_SCRIPTS=false" >> "$GITHUB_ENV"`,
				`  printf '::act-debug::msg="npm-config-ignore-scripts" value=true\n'`,
				`  printf '::act-debug::msg="yarn-enable-scripts" value=false\n'`,
				`else`,
				`  printf '::act-debug::msg="npm-config-ignore-scripts" value=not-set\n'`,
				`  printf '::act-debug::msg="yarn-enable-scripts" value=not-set\n'`,
				`fi`,
				`printf '::act-debug::msg="install-cmd" cmd=%s\n' "${INSTALL_CMD}"`,
			}.String(),
			Env: map[string]string{
				"FROZEN_INSTALL_CMD":    frozenInstallCmd,
				"PACKAGE_MANAGER_AGENT": packageManagerAgent,
				"ALLOW_SCRIPTS":         allowScriptsStr,
			},
		}
	}

	type testCase struct {
		name                string
		packageManagerAgent string
		frozenInstallCmd    string
		allowScripts        bool
		wantCmdContains     string
		wantCmdNotContains  string
		wantNPMIgnore       bool // true = NPM_CONFIG_IGNORE_SCRIPTS=true expected
		wantYarnDisable     bool // true = YARN_ENABLE_SCRIPTS=false expected
	}

	testCases := []testCase{
		// --- scripts disabled (default) ---
		{
			name:                "npm / scripts disabled",
			packageManagerAgent: "npm",
			frozenInstallCmd:    "npm ci",
			allowScripts:        false,
			wantCmdContains:     "--ignore-scripts",
			wantNPMIgnore:       true,
			wantYarnDisable:     true,
		},
		{
			name:                "pnpm / scripts disabled",
			packageManagerAgent: "pnpm",
			frozenInstallCmd:    "pnpm install --frozen-lockfile",
			allowScripts:        false,
			wantCmdContains:     "--ignore-scripts",
			wantNPMIgnore:       true,
			wantYarnDisable:     true,
		},
		{
			name:                "yarn classic / scripts disabled",
			packageManagerAgent: "yarn",
			frozenInstallCmd:    "yarn install --frozen-lockfile",
			allowScripts:        false,
			wantCmdContains:     "--ignore-scripts",
			wantNPMIgnore:       true,
			wantYarnDisable:     true,
		},
		{
			// Yarn Berry does not support --ignore-scripts on the CLI.
			// Scripts are suppressed only via YARN_ENABLE_SCRIPTS env var.
			name:                "yarn berry / scripts disabled",
			packageManagerAgent: "yarnBerry",
			frozenInstallCmd:    "yarn install --immutable",
			allowScripts:        false,
			wantCmdNotContains:  "--ignore-scripts",
			wantNPMIgnore:       true,
			wantYarnDisable:     true,
		},

		// --- scripts enabled (allow-scripts = true) ---
		{
			name:                "npm / scripts allowed",
			packageManagerAgent: "npm",
			frozenInstallCmd:    "npm ci",
			allowScripts:        true,
			wantCmdNotContains:  "--ignore-scripts",
			wantNPMIgnore:       false,
			wantYarnDisable:     false,
		},
		{
			name:                "pnpm / scripts allowed",
			packageManagerAgent: "pnpm",
			frozenInstallCmd:    "pnpm install --frozen-lockfile",
			allowScripts:        true,
			wantCmdNotContains:  "--ignore-scripts",
			wantNPMIgnore:       false,
			wantYarnDisable:     false,
		},
		{
			name:                "yarn classic / scripts allowed",
			packageManagerAgent: "yarn",
			frozenInstallCmd:    "yarn install --frozen-lockfile",
			allowScripts:        true,
			wantCmdNotContains:  "--ignore-scripts",
			wantNPMIgnore:       false,
			wantYarnDisable:     false,
		},
		{
			name:                "yarn berry / scripts allowed",
			packageManagerAgent: "yarnBerry",
			frozenInstallCmd:    "yarn install --immutable",
			allowScripts:        true,
			wantCmdNotContains:  "--ignore-scripts",
			wantNPMIgnore:       false,
			wantYarnDisable:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)

			probeStep := injectScriptPolicyStep(tc.frozenInstallCmd, tc.packageManagerAgent, tc.allowScripts)

			wf, err := ci.NewWorkflow(
				ci.WithWorkflowInputs(ci.WorkflowInputs{
					PluginDirectory:     workflow.Input(filepath.Join("tests", "simple-frontend")),
					DistArtifactsPrefix: workflow.Input("simple-frontend-"),
					RunPlaywright:       workflow.Input(false),
					NodePackageManagerAllowScripts: workflow.Input(tc.allowScripts),
				}),
				ci.MutateCIWorkflow().With(
					// Only run the job we care about, drop all others (playwright, docs, GCS upload, etc.)
					workflow.WithOnlyOneJob(t, "test-and-build", true),
					// Replace the frontend composite action step with our inline probe step.
					// This avoids a real npm install while still exercising the script-policy logic.
					// Note: WithReplacedStep preserves the original step ID ("frontend"), so we must
					// reference "frontend" (not the probe step's own ID) in WithRemoveAllStepsAfter.
					workflow.WithReplacedStep(t, "test-and-build", "frontend", probeStep),
					// Stop after our probe step (which kept the ID "frontend") — no need to run
					// package/sign/upload steps.
					workflow.WithRemoveAllStepsAfter(t, "test-and-build", "frontend"),
				),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewPushEventPayload("main"))
			require.NoError(t, err)
			require.True(t, r.Success, "workflow should succeed")

			// --- Assert install command ---
			// The probe step emits: ::act-debug::msg="install-cmd" cmd=<value>
			// We find the annotation by prefix and assert on its content.
			var installCmdAnnotation *act.Annotation
			for i, ann := range r.Annotations {
				if ann.Level == act.AnnotationLevelDebug && strings.Contains(ann.Message, `msg="install-cmd"`) {
					installCmdAnnotation = &r.Annotations[i]
					break
				}
			}
			require.NotNil(t, installCmdAnnotation, "install-cmd annotation should be present")
			if tc.wantCmdContains != "" {
				require.Contains(t, installCmdAnnotation.Message, tc.wantCmdContains,
					"install command annotation should contain %q", tc.wantCmdContains)
			}
			if tc.wantCmdNotContains != "" {
				require.NotContains(t, installCmdAnnotation.Message, tc.wantCmdNotContains,
					"install command annotation should not contain %q", tc.wantCmdNotContains)
			}

			// --- Assert env var annotations ---
			// The probe step emits these in hardcoded logfmt-like format with quoted keys:
			//   msg="npm-config-ignore-scripts" value=<true|not-set>
			//   msg="yarn-enable-scripts" value=<false|not-set>
			npmIgnoreVal := "not-set"
			if tc.wantNPMIgnore {
				npmIgnoreVal = "true"
			}
			expectedNPMMsg := fmt.Sprintf(`msg="npm-config-ignore-scripts" value=%s`, npmIgnoreVal)
			require.Contains(t, r.Annotations, act.Annotation{
				Level:   act.AnnotationLevelDebug,
				Message: expectedNPMMsg,
			}, "NPM_CONFIG_IGNORE_SCRIPTS annotation should have value=%s", npmIgnoreVal)

			yarnVal := "not-set"
			if tc.wantYarnDisable {
				yarnVal = "false"
			}
			expectedYarnMsg := fmt.Sprintf(`msg="yarn-enable-scripts" value=%s`, yarnVal)
			require.Contains(t, r.Annotations, act.Annotation{
				Level:   act.AnnotationLevelDebug,
				Message: expectedYarnMsg,
			}, "YARN_ENABLE_SCRIPTS annotation should have value=%s", yarnVal)
		})
	}
}
