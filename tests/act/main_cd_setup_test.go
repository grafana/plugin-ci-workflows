package main

import (
	"path/filepath"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow/cd"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow/ci"
	"github.com/stretchr/testify/require"
)

func TestCD_Setup(t *testing.T) {
	gitSha, err := getGitCommitSHA()
	require.NoError(t, err)

	// Same expressions as in the auto-cd-example workflow, which is widely used for provisioned plugins
	autoCDExamplePushInputs := cd.WorkflowInputs{
		CI: ci.WorkflowInputs{
			PluginVersionSuffix: workflow.Input("${{ github.event_name == 'push' && github.sha || github.event.pull_request.head.sha }}"),
		},
		Environment: workflow.Input("${{ (github.event_name == 'push' && github.ref_name == 'main') && 'dev' || 'none' }}"),
	}
	autoCDExamplePublishInputs := cd.WorkflowInputs{
		Branch:      workflow.Input("${{ github.event.inputs.branch }}"),
		Environment: workflow.Input("${{ github.event.inputs.environment }}"),
	}

	type testCase struct {
		name              string
		inputs            cd.WorkflowInputs
		workflowOptions   []cd.WorkflowOption
		expOutputs        map[string]string
		triggerEvent      *act.Event
		expFailureMessage string
	}

	type testSuite struct {
		name      string
		testCases []testCase
	}

	for _, ts := range []testSuite{
		{
			name: "simple",
			testCases: []testCase{
				{
					name: "simple-frontend normal dev deployment",
					inputs: cd.WorkflowInputs{
						CI: ci.WorkflowInputs{
							PluginDirectory: workflow.Input(filepath.Join("tests", "simple-frontend")),
						},
						Environment: workflow.Input("dev"),
					},
					expOutputs: map[string]string{
						"environments":          `["dev"]`,
						"publish-docs":          "false",
						"plugin-version-suffix": "",
						// Should have only "any" platform because it has no backend
						"platforms": `["any"]`,
					},
				},
				{
					name: "simple-backend normal dev deployment",
					inputs: cd.WorkflowInputs{
						CI: ci.WorkflowInputs{
							PluginDirectory: workflow.Input(filepath.Join("tests", "simple-backend")),
						},
						Environment: workflow.Input("dev"),
					},
					expOutputs: map[string]string{
						"environments":          `["dev"]`,
						"publish-docs":          "false",
						"plugin-version-suffix": "",
						// Should have all platforms because it has a backend
						"platforms": `["linux","darwin","windows","any"]`,
					},
				},
			},
		},
		{
			name: "environments",
			testCases: []testCase{
				{
					name: "ops deploys to ops only",
					inputs: cd.WorkflowInputs{
						Environment: workflow.Input("ops"),
					},
					expOutputs: map[string]string{
						"environments": `["ops"]`,
					},
				},
				{
					name: "staging is an alias for ops",
					inputs: cd.WorkflowInputs{
						Environment: workflow.Input("staging"),
					},
					expOutputs: map[string]string{
						"environments": `["ops"]`,
					},
				},
				{
					name: "prod-canary deploys to prod-canary only",
					inputs: cd.WorkflowInputs{
						Environment: workflow.Input("prod-canary"),
					},
					expOutputs: map[string]string{
						"environments": `["prod-canary"]`,
					},
				},
				{
					name: "prod deploys to all environments",
					inputs: cd.WorkflowInputs{
						Environment: workflow.Input("prod"),
					},
					expOutputs: map[string]string{
						// (prod-canary is implicitly included in prod)
						"environments": `["dev","ops","prod"]`,
					},
				},
				{
					name: "can target multiple environments",
					inputs: cd.WorkflowInputs{
						Environment: workflow.Input("dev,ops"),
					},
					expOutputs: map[string]string{
						"environments": `["dev","ops"]`,
					},
				},
				{
					name: "none does not deploy anything",
					inputs: cd.WorkflowInputs{
						Environment: workflow.Input("none"),
					},
					expOutputs: map[string]string{
						"environments": `[]`,
					},
				},
				{
					name: "unsupported environments are filtered out",
					inputs: cd.WorkflowInputs{
						Environment: workflow.Input("dev,unsupported"),
					},
					expOutputs: map[string]string{
						"environments": `["dev"]`,
					},
				},
				{
					name: "no valid environments return an error",
					inputs: cd.WorkflowInputs{
						Environment: workflow.Input("a,b"),
					},
					expFailureMessage: `Invalid environment(s): a,b`,
				},
			},
		},
		{
			name: "deployments blocking",
			testCases: []testCase{
				{
					name:         "prod deployments are blocked if not a release reference",
					triggerEvent: newPointer(act.NewPushEventPayload("feature-branch")),
					inputs: cd.WorkflowInputs{
						// Default value for ReleaseReferenceRegex should be "main".
						Environment: workflow.Input("prod"),
					},
					expFailureMessage: `The reference 'feature-branch' is not a release reference. Deploying to 'prod' is only allowed from release reference.`,
				},
				{
					name:         "release reference regex can be set via inputs",
					triggerEvent: newPointer(act.NewPushEventPayload("release/1.2.0")),
					inputs: cd.WorkflowInputs{
						Environment:           workflow.Input("prod"),
						ReleaseReferenceRegex: workflow.Input("release/.*"),
					},
					expOutputs: map[string]string{
						"environments": `["dev","ops","prod"]`,
					},
				},
			},
		},
		{
			name: "automatic plugin version suffix",
			testCases: []testCase{
				{
					name:         "plugin version suffix is set if deploying to dev and not a release reference",
					triggerEvent: newPointer(act.NewPushEventPayload("feature-branch")),
					inputs: cd.WorkflowInputs{
						Environment: workflow.Input("dev"),
					},
					expOutputs: map[string]string{
						"plugin-version-suffix": gitSha,
					},
				},
				{
					name:         "plugin version suffix is set if deploying to prod-canary from a release reference",
					triggerEvent: newPointer(act.NewPushEventPayload("main")),
					inputs: cd.WorkflowInputs{
						Environment: workflow.Input("prod-canary"),
						// (default value, but let's be explicit)
						ReleaseReferenceRegex: workflow.Input("main"),
					},
					expOutputs: map[string]string{
						"environments":          `["prod-canary"]`,
						"plugin-version-suffix": gitSha,
					},
				},
			},
		}, {
			// examples/base/provisioned-plugin-auto-cd/push.yaml: triggered on every push to main and PRs
			name: "auto-cd-example:push",
			testCases: []testCase{
				{
					name:         "has plugin version suffix when deploying to dev",
					triggerEvent: newPointer(act.NewPushEventPayload("main")),
					inputs:       autoCDExamplePushInputs,
					expOutputs: map[string]string{
						"environments":          `["dev"]`,
						"plugin-version-suffix": gitSha,
					},
				},
			},
		}, {
			// same as above, but for pull_request trigger
			name: "auto-cd-example:pull_request",
			testCases: []testCase{
				{
					name:         "does not deploy pull requests automatically but adds plugin version suffix to ci build",
					triggerEvent: newPointer(act.NewPullRequestEventPayload("feature-branch")),
					inputs:       autoCDExamplePushInputs,
					expOutputs: map[string]string{
						"environments":          `[]`,
						"plugin-version-suffix": gitSha,
					},
				},
			},
		}, {
			// examples/base/provisioned-plugin-auto-cd/publish.yaml: triggered manually from the UI (for cutting releases and deploying)
			name: "auto-cd-example:workflow_dispatch",
			testCases: []testCase{
				{
					name: "does not have plugin version suffix when deploying to prod",
					triggerEvent: newPointer(
						act.NewWorkflowDispatchEventPayload(map[string]any{
							"branch":      "main",
							"environment": "prod",
						}),
					),
					inputs: autoCDExamplePublishInputs,
					expOutputs: map[string]string{
						"environments":          `["dev","ops","prod"]`,
						"plugin-version-suffix": "",
					},
				},
				{
					name: "does not have plugin version suffix when deploying to dev",
					triggerEvent: newPointer(
						act.NewWorkflowDispatchEventPayload(map[string]any{
							"branch":      "main",
							"environment": "dev",
						}),
					),
					inputs: autoCDExamplePublishInputs,
					expOutputs: map[string]string{
						"environments":          `["dev"]`,
						"plugin-version-suffix": "",
					},
				},
				{
					name: "has plugin version suffix when deploying non-main to dev",
					triggerEvent: newPointer(
						act.NewWorkflowDispatchEventPayload(map[string]any{
							"branch":      "feature-branch",
							"environment": "dev",
						}),
					),
					inputs: autoCDExamplePublishInputs,
					expOutputs: map[string]string{
						"environments":          `["dev"]`,
						"plugin-version-suffix": gitSha,
					},
				},
				{
					name: "can deploy branches that match release reference regex to prod if there is no PR",
					triggerEvent: newPointer(
						act.NewWorkflowDispatchEventPayload(map[string]any{}),
					),
					inputs: cd.WorkflowInputs{
						Branch:                workflow.Input("release/1.2.0"),
						ReleaseReferenceRegex: workflow.Input("release/.*"),
						Environment:           workflow.Input("prod"),
					},
					workflowOptions: []cd.WorkflowOption{
						cd.MutateCDWorkflow().With(
							// Mock GitHub API response (no PR)
							workflow.WithEnvironment(t, "setup", "vars", map[string]string{
								"ACT_MOCK_PRS": `[]`,
							}),
						),
					},
					expOutputs: map[string]string{
						"environments":          `["dev","ops","prod"]`,
						"plugin-version-suffix": "",
					},
				},
				{
					name: "cannot deploy PRs to prod",
					triggerEvent: newPointer(
						act.NewWorkflowDispatchEventPayload(map[string]any{
							"branch":      "feature-branch",
							"environment": "prod",
						}),
					),
					inputs:            autoCDExamplePublishInputs,
					expFailureMessage: `is not a release reference. Deploying to 'prod' is only allowed from release reference.`,
				},
				{
					name: "cannot deploy branches that match release reference regex to prod if there's an open PR",
					triggerEvent: newPointer(
						act.NewWorkflowDispatchEventPayload(map[string]any{}),
					),
					inputs: cd.WorkflowInputs{
						Branch:                workflow.Input("release/1.2.0"),
						ReleaseReferenceRegex: workflow.Input("release/.*"),
						Environment:           workflow.Input("prod"),
					},
					workflowOptions: []cd.WorkflowOption{
						cd.MutateCDWorkflow().With(
							// Mock GitHub API response (PR open and unmerged)
							workflow.WithEnvironment(t, "setup", "vars", map[string]string{
								"ACT_MOCK_PRS": `[{ "merged_at": null }]`,
							}),
						),
					},
					expFailureMessage: `is a release branch but it has an open PR.`,
				},
				{
					name: "can deploy branches that match release reference regex to prod if the PR has been merged",
					triggerEvent: newPointer(
						act.NewWorkflowDispatchEventPayload(map[string]any{}),
					),
					inputs: cd.WorkflowInputs{
						Branch:                workflow.Input("release/1.2.0"),
						ReleaseReferenceRegex: workflow.Input("release/.*"),
						Environment:           workflow.Input("prod"),
					},
					workflowOptions: []cd.WorkflowOption{
						cd.MutateCDWorkflow().With(
							// Mock GitHub API response (PR closed and merged)
							workflow.WithEnvironment(t, "setup", "vars", map[string]string{
								"ACT_MOCK_PRS": `[{ "merged_at": "2026-01-16T12:00:00Z" }]`,
							}),
						),
					},
					expOutputs: map[string]string{
						"environments":          `["dev","ops","prod"]`,
						"plugin-version-suffix": "",
					},
				},
				{
					name: "can force deploy a release branch to prod if there's an open PR and allow-publishing-prs-to-prod is true",
					triggerEvent: newPointer(
						act.NewWorkflowDispatchEventPayload(map[string]any{}),
					),
					inputs: cd.WorkflowInputs{
						Branch:                   workflow.Input("release/1.2.0"),
						ReleaseReferenceRegex:    workflow.Input("release/.*"),
						Environment:              workflow.Input("prod"),
						AllowPublishingPRsToProd: workflow.Input(true),
					},
					workflowOptions: []cd.WorkflowOption{
						cd.MutateCDWorkflow().With(
							// Mock GitHub API response (PR open and unmerged)
							workflow.WithEnvironment(t, "setup", "vars", map[string]string{
								"ACT_MOCK_PRS": `[{ "merged_at": null }]`,
							}),
						),
					},
					expOutputs: map[string]string{
						"environments":          `["dev","ops","prod"]`,
						"plugin-version-suffix": "",
					},
				},
			},
		},
		{
			name: "docs publishing",
			testCases: []testCase{
				{
					name: "published when targeting prod",
					triggerEvent: newPointer(
						act.NewWorkflowDispatchEventPayload(map[string]any{
							"branch":      "main",
							"environment": "prod",
						}),
					),
					inputs: autoCDExamplePublishInputs,
					expOutputs: map[string]string{
						"publish-docs": "true",
					},
				},
				{
					name: "not published when not targeting prod",
					triggerEvent: newPointer(
						act.NewWorkflowDispatchEventPayload(map[string]any{
							"branch":      "main",
							"environment": "dev",
						}),
					),
					inputs: autoCDExamplePublishInputs,
					expOutputs: map[string]string{
						"publish-docs": "false",
					},
				},
				{
					name:         "not published if disabled via input and targeting prod",
					triggerEvent: newPointer(act.NewWorkflowDispatchEventPayload(map[string]any{})),
					inputs: cd.WorkflowInputs{
						DisableDocsPublishing: workflow.Input(true),

						Branch:      workflow.Input("main"),
						Environment: workflow.Input("prod"),
					},
					expOutputs: map[string]string{
						"environments": `["dev","ops","prod"]`,
						"publish-docs": "false",
					},
				},
				{
					name:         "docs-only does not publish the plugin",
					triggerEvent: newPointer(act.NewWorkflowDispatchEventPayload(map[string]any{})),
					inputs: cd.WorkflowInputs{
						DocsOnly:    workflow.Input(true),
						Branch:      workflow.Input("main"),
						Environment: workflow.Input("prod"),
					},
					expOutputs: map[string]string{
						"environments": `[]`,
						"publish-docs": "true",
					},
				},
				{
					name:         "docs-only fails if not targeting prod",
					triggerEvent: newPointer(act.NewWorkflowDispatchEventPayload(map[string]any{})),
					inputs: cd.WorkflowInputs{
						DocsOnly:    workflow.Input(true),
						Branch:      workflow.Input("main"),
						Environment: workflow.Input("ops"),
					},
					expFailureMessage: `Only 'prod' environment is allowed for docs publishing.`,
				},
				{
					name:         "docs-only fails if targeting prod and a non-release reference",
					triggerEvent: newPointer(act.NewWorkflowDispatchEventPayload(map[string]any{})),
					inputs: cd.WorkflowInputs{
						DocsOnly:              workflow.Input(true),
						Branch:                workflow.Input("feature-1"),
						ReleaseReferenceRegex: workflow.Input("main|release/.*"),
						Environment:           workflow.Input("prod"),
					},
					expFailureMessage: `Non-release references cannot be used for docs publishing.`,
				},
			},
		},
	} {
		t.Run(ts.name, func(t *testing.T) {
			t.Parallel()
			for _, tc := range ts.testCases {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					runner, err := act.NewRunner(t)
					require.NoError(t, err)

					// Create trigger event, defaulting to a push to main
					var triggerEvent act.Event
					if tc.triggerEvent != nil {
						triggerEvent = *tc.triggerEvent
					} else {
						triggerEvent = act.NewPushEventPayload("main")
					}

					// Create testing workflow
					opts := []cd.WorkflowOption{
						cd.WithWorkflowInputs(tc.inputs),
						// Only run the setup job, not CI or the rest of CD.
						cd.MutateCDWorkflow().With(
							workflow.WithOnlyOneJob(t, "setup", false),
						),
					}
					// Adjust the testing workflow depending on the trigger event, otherwise act doesn't run it
					switch triggerEvent.Kind {
					case act.EventKindWorkflowDispatch:
						opts = append(opts, cd.MutateTestingWorkflow().With(
							workflow.WithWorkflowDispatchTrigger(map[string]workflow.WorkflowCallInput{
								"branch": {
									Type:    workflow.WorkflowCallInputTypeString,
									Default: "main",
								},
								"environment": {
									Type:     workflow.WorkflowCallInputTypeChoice,
									Required: true,
									Options:  []any{"dev", "ops", "prod-canary", "prod"},
								},
								"docs-only": {
									Type:     workflow.WorkflowCallInputTypeBoolean,
									Required: false,
									Default:  false,
								},
							})),
						)
					case act.EventKindPullRequest:
						opts = append(opts, cd.MutateTestingWorkflow().With(
							workflow.WithPullRequestTrigger([]string{"main"}),
						))
					}
					// Additional user-provided options
					opts = append(opts, tc.workflowOptions...)

					wf, err := cd.NewWorkflow(opts...)
					require.NoError(t, err)

					// Run the workflow
					r, err := runner.Run(wf, triggerEvent)
					require.NoError(t, err)

					// Ensure the test case makes sense
					if tc.expFailureMessage != "" && len(tc.expOutputs) > 0 {
						require.FailNow(t, "expFailureMessage and expOutputs cannot be set at the same time")
					}

					// Assert
					if tc.expFailureMessage != "" {
						require.False(t, r.Success, "workflow should fail")
						require.Len(t, r.Annotations, 1, "should have exactly one annotation")
						require.Equal(t, act.AnnotationLevelError, r.Annotations[0].Level, "annotation level should be error")
						require.Contains(t, r.Annotations[0].Message, tc.expFailureMessage, "annotation message should contain failure message")
					} else {
						require.True(t, r.Success, "workflow should succeed")
						for k, exp := range tc.expOutputs {
							o, ok := r.Outputs.Get("setup", "vars", k)
							require.True(t, ok)
							require.Equalf(t, exp, o, "output %q should be %q", k, exp)
						}
					}
				})
			}
		})
	}
}
