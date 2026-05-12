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
// in actions/internal/plugins/frontend/action.yml by running the real frontend
// composite action against the test plugins and asserting on the ::act-debug::
// annotations it emits.
//
// The annotations are emitted directly by the production "Configure package manager
// script policy" step in the frontend action:
//
//	msg="install-cmd"               cmd=<the computed install command>
//	msg="npm-config-ignore-scripts" value=<true|not-set>
//	msg="yarn-enable-scripts"       value=<false|not-set>
//
// We use the three real test plugins so that package-manager-detect runs for real
// and produces the correct agent/frozenInstallCmd values:
//   - tests/simple-frontend      → npm  (npm ci)
//   - tests/simple-frontend-yarn → yarn classic (yarn install --frozen-lockfile)
//   - tests/simple-frontend-pnpm → pnpm (pnpm install --frozen-lockfile)
//
// For each plugin we test both allow-scripts=false (default, secure) and
// allow-scripts=true (escape-hatch).
func TestInstallScripts(t *testing.T) {
	type testCase struct {
		folder             string
		allowScripts       bool
		wantCmdContains    string
		wantCmdNotContains string
		wantNPMIgnore      bool // true → NPM_CONFIG_IGNORE_SCRIPTS=true emitted
		wantYarnDisable    bool // true → YARN_ENABLE_SCRIPTS=false emitted
	}

	for _, tc := range []testCase{
		// --- scripts disabled (default, secure) ---
		{
			folder:          "simple-frontend",
			allowScripts:    false,
			wantCmdContains: "--ignore-scripts",
			wantNPMIgnore:   true,
			wantYarnDisable: true,
		},
		{
			folder:          "simple-frontend-yarn",
			allowScripts:    false,
			wantCmdContains: "--ignore-scripts",
			wantNPMIgnore:   true,
			wantYarnDisable: true,
		},
		{
			folder:          "simple-frontend-pnpm",
			allowScripts:    false,
			wantCmdContains: "--ignore-scripts",
			wantNPMIgnore:   true,
			wantYarnDisable: true,
		},

		// --- scripts enabled (allow-scripts = true, escape-hatch) ---
		{
			folder:             "simple-frontend",
			allowScripts:       true,
			wantCmdNotContains: "--ignore-scripts",
			wantNPMIgnore:      false,
			wantYarnDisable:    false,
		},
		{
			folder:             "simple-frontend-yarn",
			allowScripts:       true,
			wantCmdNotContains: "--ignore-scripts",
			wantNPMIgnore:      false,
			wantYarnDisable:    false,
		},
		{
			folder:             "simple-frontend-pnpm",
			allowScripts:       true,
			wantCmdNotContains: "--ignore-scripts",
			wantNPMIgnore:      false,
			wantYarnDisable:    false,
		},
	} {
		name := fmt.Sprintf("%s/allow-scripts=%v", tc.folder, tc.allowScripts)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)

			wf, err := ci.NewWorkflow(
				ci.WithWorkflowInputs(ci.WorkflowInputs{
					PluginDirectory:                workflow.Input(filepath.Join("tests", tc.folder)),
					DistArtifactsPrefix:            workflow.Input(tc.folder + "-"),
					RunPlaywright:                  workflow.Input(false),
					NodePackageManagerAllowScripts: workflow.Input(tc.allowScripts),
				}),
				ci.MutateCIWorkflow().With(
					// Only run the job we care about.
					workflow.WithOnlyOneJob(t, "test-and-build", true),
					// Stop right after the frontend step completes — the ::act-debug::
					// annotations we assert on are emitted inside the frontend composite
					// action's "Configure package manager script policy" step. We don't
					// need the backend, packaging, or upload steps.
					workflow.WithRemoveAllStepsAfter(t, "test-and-build", "frontend"),
				),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewPushEventPayload("main"))
			require.NoError(t, err)
			require.True(t, r.Success, "workflow should succeed")

			// --- Assert install-cmd annotation ---
			// The frontend action emits: ::act-debug::msg="install-cmd" cmd=<value>
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
					"install command should contain %q", tc.wantCmdContains)
			}
			if tc.wantCmdNotContains != "" {
				require.NotContains(t, installCmdAnnotation.Message, tc.wantCmdNotContains,
					"install command should not contain %q", tc.wantCmdNotContains)
			}

			// --- Assert env var annotations ---
			// The frontend action emits one of:
			//   ::act-debug::msg="npm-config-ignore-scripts" value=true
			//   ::act-debug::msg="npm-config-ignore-scripts" value=not-set
			npmIgnoreVal := "not-set"
			if tc.wantNPMIgnore {
				npmIgnoreVal = "true"
			}
			require.Contains(t, r.Annotations, act.Annotation{
				Level:   act.AnnotationLevelDebug,
				Message: fmt.Sprintf(`msg="npm-config-ignore-scripts" value=%s`, npmIgnoreVal),
			}, "NPM_CONFIG_IGNORE_SCRIPTS annotation should have value=%s", npmIgnoreVal)

			//   ::act-debug::msg="yarn-enable-scripts" value=false
			//   ::act-debug::msg="yarn-enable-scripts" value=not-set
			yarnVal := "not-set"
			if tc.wantYarnDisable {
				yarnVal = "false"
			}
			require.Contains(t, r.Annotations, act.Annotation{
				Level:   act.AnnotationLevelDebug,
				Message: fmt.Sprintf(`msg="yarn-enable-scripts" value=%s`, yarnVal),
			}, "YARN_ENABLE_SCRIPTS annotation should have value=%s", yarnVal)
		})
	}
}
