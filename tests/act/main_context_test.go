package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/stretchr/testify/require"
)

func TestContext(t *testing.T) {
	const (
		testAndBuild    = "test-and-build"
		workflowContext = "workflow-context"
	)

	for _, tc := range []struct {
		actor        string
		testingInput bool
		expIsTrusted bool
	}{
		// Non-testing mode: trust only trusted actors
		{actor: "dependabot[bot]", testingInput: false, expIsTrusted: true},
		{actor: "renovate-sh-app[bot]", testingInput: false, expIsTrusted: true},
		{actor: "grafana-plugins-platform-bot[bot]", testingInput: false, expIsTrusted: true},
		{actor: "hacker[bot]", testingInput: false, expIsTrusted: false},

		// In testing mode, context is never trusted
		{actor: "dependabot[bot]", testingInput: true, expIsTrusted: false},
		{actor: "renovate-sh-app[bot]", testingInput: true, expIsTrusted: false},
		{actor: "grafana-plugins-platform-bot[bot]", testingInput: true, expIsTrusted: false},
		{actor: "hacker[bot]", testingInput: true, expIsTrusted: false},

		// TODO: pull_request event, fork and untrusted events test cases
	} {
		t.Run(fmt.Sprintf("%s testing=%t", tc.actor, tc.testingInput), func(t *testing.T) {
			t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)

			wf, err := workflow.NewSimpleCI(
				workflow.WithPluginDirectoryInput(filepath.Join("tests", "simple-frontend")),
				workflow.WithDistArtifactPrefixInput("simple-frontend-"),

				// Eventually disable testing mode, otherwise context is never trusted
				workflow.WithTestingInput(tc.testingInput),

				// Only run test-and-build job and stop after workflow-context step
				// (no need to build the plugin, etc, for this test)
				workflow.WithOnlyOneJob(t, testAndBuild),
				workflow.WithRemoveAllStepsAfter(t, testAndBuild, workflowContext),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewPushEventPayload("main", act.WithEventActor(tc.actor)))
			require.NoError(t, err)
			require.True(t, r.Success, "workflow should succeed")

			contextPayload, ok := r.Outputs.Get(testAndBuild, workflowContext, "result")
			require.True(t, ok, "output result should be present")
			var context struct {
				IsTrusted bool `json:"isTrusted"`
			}
			require.NoError(t, json.Unmarshal([]byte(contextPayload), &context))
			require.Equalf(t, tc.expIsTrusted, context.IsTrusted, "workflow should not be trusted for %q actor", tc.actor)
		})
	}
}
