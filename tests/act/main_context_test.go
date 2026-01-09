package main

import (
	"encoding/json"
	"fmt"
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
	} {
		t.Run(fmt.Sprintf("push event: %s testing=%t", tc.actor, tc.testingInput), func(t *testing.T) {
			t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)

			// Copy plugin to temp directory to avoid leftover files
			tempDir, err := act.CopyPluginToTemp(t, "simple-frontend")
			require.NoError(t, err)

			wf, err := workflow.NewSimpleCI(
				workflow.WithPluginDirectoryInput(tempDir),
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
				IsForkPR  bool `json:"isForkPR"`
			}
			require.NoError(t, json.Unmarshal([]byte(contextPayload), &context))
			require.Equalf(t, tc.expIsTrusted, context.IsTrusted, "workflow should not be trusted for %q actor", tc.actor)
			require.False(t, context.IsForkPR, "push event should not be a fork PR")
		})
	}

	// Test pull_request events (non-fork)
	for _, tc := range []struct {
		name         string
		actor        string
		testingInput bool
		expIsTrusted bool
		expIsForkPR  bool
	}{
		// Non-fork PR from regular user (not a bot) - should be trusted if not testing
		{name: "non-fork PR from regular user", actor: "regular-user", testingInput: false, expIsTrusted: true, expIsForkPR: false},
		{name: "non-fork PR from regular user (testing)", actor: "regular-user", testingInput: true, expIsTrusted: false, expIsForkPR: false},

		// Non-fork PR from trusted bot - should be trusted if not testing
		{name: "non-fork PR from trusted bot", actor: "dependabot[bot]", testingInput: false, expIsTrusted: true, expIsForkPR: false},
		{name: "non-fork PR from trusted bot (testing)", actor: "dependabot[bot]", testingInput: true, expIsTrusted: false, expIsForkPR: false},

		// Non-fork PR from untrusted bot - should NOT be trusted
		{name: "non-fork PR from untrusted bot", actor: "hacker[bot]", testingInput: false, expIsTrusted: false, expIsForkPR: false},
		{name: "non-fork PR from untrusted bot (testing)", actor: "hacker[bot]", testingInput: true, expIsTrusted: false, expIsForkPR: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)

			// Copy plugin to temp directory to avoid leftover files
			tempDir, err := act.CopyPluginToTemp(t, "simple-frontend")
			require.NoError(t, err)

			wf, err := workflow.NewSimpleCI(
				workflow.WithPluginDirectoryInput(tempDir),
				workflow.WithDistArtifactPrefixInput("simple-frontend-"),
				workflow.WithTestingInput(tc.testingInput),
				workflow.WithOnlyOneJob(t, testAndBuild),
				workflow.WithRemoveAllStepsAfter(t, testAndBuild, workflowContext),
			)
			require.NoError(t, err)

			// Create a non-fork PR event (head repo same as base repo)
			// NewPullRequestEventPayload already sets this up by default
			prEvent := act.NewPullRequestEventPayload("feature-branch", act.WithEventActor(tc.actor))

			r, err := runner.Run(wf, prEvent)
			require.NoError(t, err)
			require.True(t, r.Success, "workflow should succeed")

			contextPayload, ok := r.Outputs.Get(testAndBuild, workflowContext, "result")
			require.True(t, ok, "output result should be present")
			var context struct {
				IsTrusted bool `json:"isTrusted"`
				IsForkPR  bool `json:"isForkPR"`
			}
			require.NoError(t, json.Unmarshal([]byte(contextPayload), &context))
			require.Equalf(t, tc.expIsTrusted, context.IsTrusted, "workflow trust status mismatch")
			require.Equalf(t, tc.expIsForkPR, context.IsForkPR, "fork PR status mismatch")
		})
	}

	// Test fork PR events - should never be trusted
	for _, tc := range []struct {
		name         string
		actor        string
		testingInput bool
		expIsTrusted bool
		expIsForkPR  bool
	}{
		{name: "fork PR from regular user", actor: "fork-user", testingInput: false, expIsTrusted: false, expIsForkPR: true},
		{name: "fork PR from regular user (testing)", actor: "fork-user", testingInput: true, expIsTrusted: false, expIsForkPR: true},
		{name: "fork PR from trusted bot", actor: "dependabot[bot]", testingInput: false, expIsTrusted: false, expIsForkPR: true},
		{name: "fork PR from trusted bot (testing)", actor: "dependabot[bot]", testingInput: true, expIsTrusted: false, expIsForkPR: true},
		{name: "fork PR from untrusted bot", actor: "hacker[bot]", testingInput: false, expIsTrusted: false, expIsForkPR: true},
		{name: "fork PR from untrusted bot (testing)", actor: "hacker[bot]", testingInput: true, expIsTrusted: false, expIsForkPR: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)

			// Copy plugin to temp directory to avoid leftover files
			tempDir, err := act.CopyPluginToTemp(t, "simple-frontend")
			require.NoError(t, err)

			wf, err := workflow.NewSimpleCI(
				workflow.WithPluginDirectoryInput(tempDir),
				workflow.WithDistArtifactPrefixInput("simple-frontend-"),
				workflow.WithTestingInput(tc.testingInput),
				workflow.WithOnlyOneJob(t, testAndBuild),
				workflow.WithRemoveAllStepsAfter(t, testAndBuild, workflowContext),
			)
			require.NoError(t, err)

			// Create a fork PR event (head repo different from base repo)
			prEvent := act.NewPullRequestEventPayload("feature-branch", act.WithEventActor(tc.actor), act.WithForkPR())

			r, err := runner.Run(wf, prEvent)
			require.NoError(t, err)
			require.True(t, r.Success, "workflow should succeed")

			contextPayload, ok := r.Outputs.Get(testAndBuild, workflowContext, "result")
			require.True(t, ok, "output result should be present")
			var context struct {
				IsTrusted bool `json:"isTrusted"`
				IsForkPR  bool `json:"isForkPR"`
			}
			require.NoError(t, json.Unmarshal([]byte(contextPayload), &context))
			require.Equalf(t, tc.expIsTrusted, context.IsTrusted, "fork PR should never be trusted")
			require.Equalf(t, tc.expIsForkPR, context.IsForkPR, "should be detected as fork PR")
		})
	}

	// Test pull_request_target events - should never be trusted (not in trusted events list)
	for _, tc := range []struct {
		name         string
		actor        string
		testingInput bool
		expIsTrusted bool
		expIsForkPR  bool
	}{
		{name: "pull_request_target from regular user", actor: "regular-user", testingInput: false, expIsTrusted: false, expIsForkPR: false},
		{name: "pull_request_target from regular user (testing)", actor: "regular-user", testingInput: true, expIsTrusted: false, expIsForkPR: false},
		{name: "pull_request_target from trusted bot", actor: "dependabot[bot]", testingInput: false, expIsTrusted: false, expIsForkPR: false},
		{name: "pull_request_target from trusted bot (testing)", actor: "dependabot[bot]", testingInput: true, expIsTrusted: false, expIsForkPR: false},
		{name: "pull_request_target from untrusted bot", actor: "hacker[bot]", testingInput: false, expIsTrusted: false, expIsForkPR: false},
		{name: "pull_request_target from untrusted bot (testing)", actor: "hacker[bot]", testingInput: true, expIsTrusted: false, expIsForkPR: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)

			// Copy plugin to temp directory to avoid leftover files
			tempDir, err := act.CopyPluginToTemp(t, "simple-frontend")
			require.NoError(t, err)

			wf, err := workflow.NewSimpleCI(
				workflow.WithPluginDirectoryInput(tempDir),
				workflow.WithDistArtifactPrefixInput("simple-frontend-"),
				workflow.WithTestingInput(tc.testingInput),
				workflow.WithPullRequestTargetTrigger([]string{"main"}),
				workflow.WithOnlyOneJob(t, testAndBuild),
				workflow.WithRemoveAllStepsAfter(t, testAndBuild, workflowContext),
			)
			require.NoError(t, err)

			// Create a pull_request_target event (untrusted event type)
			prTargetEvent := act.NewEventPayload(act.EventKindPullRequestTarget, map[string]any{
				"action": "opened",
				"repository": map[string]any{
					"full_name": "grafana/plugin-ci-workflows",
				},
				"pull_request": map[string]any{
					"head": map[string]any{
						"ref": "feature-branch",
						"repo": map[string]any{
							"full_name": "grafana/plugin-ci-workflows",
						},
					},
					"base": map[string]any{
						"ref": "main",
					},
				},
			}, act.WithEventActor(tc.actor))

			r, err := runner.Run(wf, prTargetEvent)
			require.NoError(t, err)
			require.True(t, r.Success, "workflow should succeed")

			contextPayload, ok := r.Outputs.Get(testAndBuild, workflowContext, "result")
			require.True(t, ok, "output result should be present")
			var context struct {
				IsTrusted bool `json:"isTrusted"`
				IsForkPR  bool `json:"isForkPR"`
			}
			require.NoError(t, json.Unmarshal([]byte(contextPayload), &context))
			require.Equalf(t, tc.expIsTrusted, context.IsTrusted, "pull_request_target should never be trusted")
			require.Equalf(t, tc.expIsForkPR, context.IsForkPR, "pull_request_target should not be detected as fork PR")
		})
	}
}
