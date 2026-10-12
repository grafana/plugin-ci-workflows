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
	"golang.org/x/mod/semver"
)

// TestInstallScripts verifies the install scripts policy computed by the setup action
// (node-package-manager-scripts input) against the three real test plugins (npm, yarn, pnpm).
//
// TODO: add a simple-frontend-yarn-berry test plugin and cover the yarnBerry agent path
// (--mode=skip-build for "none", YARN_ENABLE_SCRIPTS for "allowlist" and "all").
func TestInstallScripts(t *testing.T) {
	type testCase struct {
		folder string
		mode   string

		// wantEffective is the effective policy (allowlist falls back to none when unsupported).
		wantEffective string
		// wantCmdSuffix is the flag(s) appended to the frozen install command ("" for none).
		wantCmdSuffix string
		wantWarning   bool
		// npmAllFlagNeedsNPM1116 means wantCmdSuffix is derived from the npm runtime version
		// (the toolcache decides it): the flag is only expected on npm >= 11.16.0.
		npmAllFlagNeedsNPM1116 bool
	}

	for _, tc := range []testCase{
		// npm: simple-frontend has no allowScripts field
		{folder: "simple-frontend", mode: "none", wantEffective: "none", wantCmdSuffix: " --ignore-scripts"},
		{folder: "simple-frontend", mode: "allowlist", wantEffective: "none", wantCmdSuffix: " --ignore-scripts", wantWarning: true},
		{folder: "simple-frontend", mode: "all", wantEffective: "all", npmAllFlagNeedsNPM1116: true},
		// yarn classic: no allowlist support
		{folder: "simple-frontend-yarn", mode: "none", wantEffective: "none", wantCmdSuffix: " --ignore-scripts"},
		{folder: "simple-frontend-yarn", mode: "allowlist", wantEffective: "none", wantCmdSuffix: " --ignore-scripts", wantWarning: true},
		{folder: "simple-frontend-yarn", mode: "all", wantEffective: "all"},
		// pnpm 10: native allowlist
		{folder: "simple-frontend-pnpm", mode: "none", wantEffective: "none", wantCmdSuffix: " --ignore-scripts"},
		{folder: "simple-frontend-pnpm", mode: "allowlist", wantEffective: "allowlist"},
		{folder: "simple-frontend-pnpm", mode: "all", wantEffective: "all", wantCmdSuffix: " --config.dangerously-allow-all-builds=true"},
	} {
		t.Run(fmt.Sprintf("%s/%s", tc.folder, tc.mode), func(t *testing.T) {
			t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)

			wf, err := ci.NewWorkflow(
				ci.WithWorkflowInputs(ci.WorkflowInputs{
					PluginDirectory:           workflow.Input(filepath.Join("tests", tc.folder)),
					DistArtifactsPrefix:       workflow.Input(tc.folder + "-"),
					RunPlaywright:             workflow.Input(false),
					NodePackageManagerScripts: workflow.Input(tc.mode),
				}),
				ci.MutateCIWorkflow().With(
					workflow.WithOnlyOneJob(t, "test-and-build", true),
					workflow.WithRemoveAllStepsAfter(t, "test-and-build", "frontend"),
				),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewPushEventPayload("main"))
			require.NoError(t, err)
			require.True(t, r.Success, "workflow should succeed")

			require.Contains(t, r.Annotations, act.Annotation{
				Level:   act.AnnotationLevelDebug,
				Message: fmt.Sprintf(`msg="scripts-policy" requested=%s effective=%s`, tc.mode, tc.wantEffective),
			})
			require.Contains(t, r.Annotations, act.Annotation{
				Level:   act.AnnotationLevelDebug,
				Message: `msg="yarn-enable-scripts" value=not-set`,
			})

			// The frozen install command differs per package manager: check only the policy suffix.
			var installCmd string
			for _, ann := range r.Annotations {
				if ann.Level == act.AnnotationLevelDebug && strings.HasPrefix(ann.Message, `msg="install-cmd" cmd=`) {
					installCmd = strings.TrimPrefix(ann.Message, `msg="install-cmd" cmd=`)
					break
				}
			}
			require.NotEmpty(t, installCmd, "install-cmd annotation should be present")

			wantCmdSuffix := tc.wantCmdSuffix
			if tc.npmAllFlagNeedsNPM1116 {
				var npmVersion string
				for _, ann := range r.Annotations {
					if ann.Level == act.AnnotationLevelDebug && strings.HasPrefix(ann.Message, `msg="pm-runtime-version" value=`) {
						npmVersion = strings.TrimPrefix(ann.Message, `msg="pm-runtime-version" value=`)
						break
					}
				}
				require.NotEmpty(t, npmVersion, "pm-runtime-version annotation should be present")
				if semver.Compare("v"+npmVersion, "v11.16.0") >= 0 {
					wantCmdSuffix = " --dangerously-allow-all-scripts"
				}
			}
			if wantCmdSuffix == "" {
				for _, flag := range []string{"--ignore-scripts", "--dangerously-allow-all", "--strict-allow-scripts"} {
					require.NotContains(t, installCmd, flag)
				}
			} else {
				require.True(t, strings.HasSuffix(installCmd, wantCmdSuffix), "install cmd %q should end with %q", installCmd, wantCmdSuffix)
			}

			var hasWarning bool
			for _, ann := range r.Annotations {
				if ann.Level == act.AnnotationLevelWarning && ann.Title == "Install scripts disabled" {
					hasWarning = true
				}
			}
			require.Equal(t, tc.wantWarning, hasWarning, "unexpected warning annotation state")
		})
	}
}
