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
// in the frontend action against the three real test plugins (npm, yarn, pnpm),
// with allow-scripts both true and false.
//
// TODO: add a simple-frontend-yarn-berry test plugin and cover the yarnBerry agent path,
// where --ignore-scripts is NOT appended to the CLI (Berry rejects it) and scripts are
// suppressed solely via the YARN_ENABLE_SCRIPTS env var.
func TestInstallScripts(t *testing.T) {
	type testCase struct {
		folder             string
		allowScripts       bool
		wantCmdContains    string
		wantCmdNotContains string
		wantNPMIgnore      bool
		wantYarnDisable    bool
	}

	for _, tc := range []testCase{
		{folder: "simple-frontend", allowScripts: false, wantCmdContains: "--ignore-scripts", wantNPMIgnore: true, wantYarnDisable: true},
		{folder: "simple-frontend-yarn", allowScripts: false, wantCmdContains: "--ignore-scripts", wantNPMIgnore: true, wantYarnDisable: true},
		{folder: "simple-frontend-pnpm", allowScripts: false, wantCmdContains: "--ignore-scripts", wantNPMIgnore: true, wantYarnDisable: true},
		{folder: "simple-frontend", allowScripts: true, wantCmdNotContains: "--ignore-scripts", wantNPMIgnore: false, wantYarnDisable: false},
		{folder: "simple-frontend-yarn", allowScripts: true, wantCmdNotContains: "--ignore-scripts", wantNPMIgnore: false, wantYarnDisable: false},
		{folder: "simple-frontend-pnpm", allowScripts: true, wantCmdNotContains: "--ignore-scripts", wantNPMIgnore: false, wantYarnDisable: false},
	} {
		t.Run(fmt.Sprintf("%s/allow-scripts=%v", tc.folder, tc.allowScripts), func(t *testing.T) {
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
					workflow.WithOnlyOneJob(t, "test-and-build", true),
					workflow.WithRemoveAllStepsAfter(t, "test-and-build", "frontend"),
				),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewPushEventPayload("main"))
			require.NoError(t, err)
			require.True(t, r.Success, "workflow should succeed")

			var installCmdAnnotation *act.Annotation
			for i, ann := range r.Annotations {
				if ann.Level == act.AnnotationLevelDebug && strings.Contains(ann.Message, `msg="install-cmd"`) {
					installCmdAnnotation = &r.Annotations[i]
					break
				}
			}
			require.NotNil(t, installCmdAnnotation, "install-cmd annotation should be present")
			if tc.wantCmdContains != "" {
				require.Contains(t, installCmdAnnotation.Message, tc.wantCmdContains)
			}
			if tc.wantCmdNotContains != "" {
				require.NotContains(t, installCmdAnnotation.Message, tc.wantCmdNotContains)
			}

			npmIgnoreVal := "not-set"
			if tc.wantNPMIgnore {
				npmIgnoreVal = "true"
			}
			require.Contains(t, r.Annotations, act.Annotation{
				Level:   act.AnnotationLevelDebug,
				Message: fmt.Sprintf(`msg="npm-config-ignore-scripts" value=%s`, npmIgnoreVal),
			})

			yarnVal := "not-set"
			if tc.wantYarnDisable {
				yarnVal = "false"
			}
			require.Contains(t, r.Annotations, act.Annotation{
				Level:   act.AnnotationLevelDebug,
				Message: fmt.Sprintf(`msg="yarn-enable-scripts" value=%s`, yarnVal),
			})
		})
	}
}
