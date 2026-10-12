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
// (node-package-manager-scripts input) against the three real test plugins (npm, yarn berry, pnpm).
func TestInstallScripts(t *testing.T) {
	type testCase struct {
		folder string
		mode   string

		// wantEffective is the effective policy (allowlist falls back to none when unsupported).
		// For npmRuntimeDependent allowlist cases it is derived from the npm runtime version.
		wantEffective string
		// wantCmdSuffix is the flag(s) appended to the frozen install command ("" for none).
		wantCmdSuffix string
		wantWarning   bool
		// wantYarnEnableScripts is the expected yarn-enable-scripts value ("" means not-set).
		wantYarnEnableScripts string
		// npmRuntimeDependent means the expectations are derived from the npm runtime version
		// (the toolcache decides it): allowlist is native on npm >= 12.0.0 (else falls back to
		// none with a warning), and the "all" flag is only expected on npm >= 11.16.0.
		npmRuntimeDependent bool
	}

	for _, tc := range []testCase{
		// npm: simple-frontend has no allowScripts field
		{folder: "simple-frontend", mode: "none", wantEffective: "none", wantCmdSuffix: " --ignore-scripts"},
		// allowlist: native on npm >= 12.0.0, otherwise falls back to none (see npmRuntimeDependent)
		{folder: "simple-frontend", mode: "allowlist", npmRuntimeDependent: true},
		{folder: "simple-frontend", mode: "all", wantEffective: "all", npmRuntimeDependent: true},
		// yarn berry (4.x): scripts controlled via --mode=skip-build and YARN_ENABLE_SCRIPTS
		{folder: "simple-frontend-yarn", mode: "none", wantEffective: "none", wantCmdSuffix: " --mode=skip-build", wantYarnEnableScripts: "false"},
		{folder: "simple-frontend-yarn", mode: "allowlist", wantEffective: "allowlist", wantYarnEnableScripts: "false"},
		{folder: "simple-frontend-yarn", mode: "all", wantEffective: "all", wantYarnEnableScripts: "true"},
		// pnpm 10+: native allowlist
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
					// Setup-only: no dependencies are installed, so there is no package manager cache to save.
					func(twf *workflow.TestingWorkflow) {
						setup := twf.BaseWorkflow.Jobs["test-and-build"].GetStep("setup")
						require.NotNil(t, setup)
						setup.With["node-setup-caching"] = false
					},
					workflow.WithRemoveAllStepsAfter(t, "test-and-build", "setup"),
				),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewPushEventPayload("main"))
			require.NoError(t, err)
			require.True(t, r.Success, "workflow should succeed")

			wantEffective, wantCmdSuffix, wantWarning := tc.wantEffective, tc.wantCmdSuffix, tc.wantWarning
			if tc.npmRuntimeDependent {
				var npmVersion string
				for _, ann := range r.Annotations {
					if ann.Level == act.AnnotationLevelDebug && strings.HasPrefix(ann.Message, `msg="pm-runtime-version" value=`) {
						npmVersion = strings.TrimPrefix(ann.Message, `msg="pm-runtime-version" value=`)
						break
					}
				}
				require.NotEmpty(t, npmVersion, "pm-runtime-version annotation should be present")
				npmVersion = "v" + npmVersion
				switch tc.mode {
				case "allowlist":
					if semver.Compare(npmVersion, "v12.0.0") >= 0 {
						wantEffective = "allowlist"
					} else {
						wantEffective, wantCmdSuffix, wantWarning = "none", " --ignore-scripts", true
					}
				case "all":
					if semver.Compare(npmVersion, "v11.16.0") >= 0 {
						wantCmdSuffix = " --dangerously-allow-all-scripts"
					}
				}
			}

			require.Contains(t, r.Annotations, act.Annotation{
				Level:   act.AnnotationLevelDebug,
				Message: fmt.Sprintf(`msg="scripts-policy" requested=%s effective=%s`, tc.mode, wantEffective),
			})
			wantYarnEnableScripts := tc.wantYarnEnableScripts
			if wantYarnEnableScripts == "" {
				wantYarnEnableScripts = "not-set"
			}
			require.Contains(t, r.Annotations, act.Annotation{
				Level:   act.AnnotationLevelDebug,
				Message: fmt.Sprintf(`msg="yarn-enable-scripts" value=%s`, wantYarnEnableScripts),
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

			if wantCmdSuffix == "" {
				for _, flag := range []string{"--ignore-scripts", "--dangerously-allow-all", "--strict-allow-scripts", "--mode=skip-build"} {
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
			require.Equal(t, wantWarning, hasWarning, "unexpected warning annotation state")
		})
	}
}
