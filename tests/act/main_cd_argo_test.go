package main

import (
	"strings"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow/cd"
	"github.com/stretchr/testify/require"
)

func TestCD_Argo(t *testing.T) {
	gitSha, err := getGitCommitSHA()
	require.NoError(t, err)

	type tc struct {
		name string

		inputs                   cd.WorkflowInputs
		expArgoShouldBeTriggered bool
		expArgoInputs            map[string]string
	}

	for _, tc := range []tc{
		// Provisioned plugins deployment
		{
			name: "provisioned plugin dev deployment with defaults",
			inputs: cd.WorkflowInputs{
				TriggerArgo:                workflow.Input(true),
				GrafanaCloudDeploymentType: workflow.Input("provisioned"),
				Environment:                workflow.Input("dev"),
				ArgoWorkflowSlackChannel:   workflow.Input("#some-slack-channel"),
			},
			expArgoShouldBeTriggered: true,
			expArgoInputs: map[string]string{
				"slug":                    "simple-frontend",
				"version":                 "1.0.0",
				"environment":             "dev",
				"slack_channel":           "#some-slack-channel",
				"commit":                  gitSha,
				"commit_link":             "https://github.com/grafana/plugin-ci-workflows/commit/" + gitSha,
				"auto_merge_environments": "dev+ops+prod-canary+prod",
				"auto_approve_durations":  `{"dev":0,"ops":0,"prod-canary":null,"prod":null}`,
				"prod_targets_all":        "true",
			},
		},
		{
			name: "provisioned plugin ops deployment without auto-merge",
			inputs: cd.WorkflowInputs{
				TriggerArgo:                workflow.Input(true),
				GrafanaCloudDeploymentType: workflow.Input("provisioned"),
				Environment:                workflow.Input("ops"),
				AutoMergeEnvironments:      workflow.Input("dev"),
				ArgoWorkflowSlackChannel:   workflow.Input("#some-slack-channel"),
			},
			expArgoShouldBeTriggered: true,
			expArgoInputs: map[string]string{
				"slug":                    "simple-frontend",
				"version":                 "1.0.0",
				"environment":             "ops",
				"slack_channel":           "#some-slack-channel",
				"commit":                  gitSha,
				"commit_link":             "https://github.com/grafana/plugin-ci-workflows/commit/" + gitSha,
				"auto_merge_environments": "dev",
				"auto_approve_durations":  `{"dev":0,"ops":0,"prod-canary":null,"prod":null}`,
				"prod_targets_all":        "true",
			},
		},
		{
			name: "provisioned plugin dev+ops deployment with dev auto-merge",
			inputs: cd.WorkflowInputs{
				TriggerArgo:                workflow.Input(true),
				GrafanaCloudDeploymentType: workflow.Input("provisioned"),
				Environment:                workflow.Input("dev,ops"),
				AutoMergeEnvironments:      workflow.Input("dev"),
				ArgoWorkflowSlackChannel:   workflow.Input("#some-slack-channel"),
			},
			expArgoShouldBeTriggered: true,
			expArgoInputs: map[string]string{
				"slug":                    "simple-frontend",
				"version":                 "1.0.0",
				"environment":             "dev+ops",
				"slack_channel":           "#some-slack-channel",
				"commit":                  gitSha,
				"commit_link":             "https://github.com/grafana/plugin-ci-workflows/commit/" + gitSha,
				"auto_merge_environments": "dev",
				"auto_approve_durations":  `{"dev":0,"ops":0,"prod-canary":null,"prod":null}`,
				"prod_targets_all":        "true",
			},
		},
		{
			name: "provisioned plugin prod deployment with dev+ops auto-merge",
			inputs: cd.WorkflowInputs{
				TriggerArgo:                workflow.Input(true),
				GrafanaCloudDeploymentType: workflow.Input("provisioned"),
				Environment:                workflow.Input("prod"),
				AutoMergeEnvironments:      workflow.Input("dev,ops"),
				ArgoWorkflowSlackChannel:   workflow.Input("#some-slack-channel"),
			},
			expArgoShouldBeTriggered: true,
			expArgoInputs: map[string]string{
				"slug":                    "simple-frontend",
				"version":                 "1.0.0",
				"environment":             "prod",
				"slack_channel":           "#some-slack-channel",
				"commit":                  gitSha,
				"commit_link":             "https://github.com/grafana/plugin-ci-workflows/commit/" + gitSha,
				"auto_merge_environments": "dev+ops",
				"auto_approve_durations":  `{"dev":0,"ops":0,"prod-canary":null,"prod":null}`,
				"prod_targets_all":        "true",
			},
		},
		{
			name: "provisioned plugin prod deployment prod_targets_all=false",
			inputs: cd.WorkflowInputs{
				TriggerArgo:                workflow.Input(true),
				GrafanaCloudDeploymentType: workflow.Input("provisioned"),
				Environment:                workflow.Input("prod"),
				ArgoWorkflowSlackChannel:   workflow.Input("#some-slack-channel"),
				ProdTargetsAll:             workflow.Input(false),
			},
			expArgoShouldBeTriggered: true,
			expArgoInputs: map[string]string{
				"slug":                    "simple-frontend",
				"version":                 "1.0.0",
				"environment":             "prod",
				"slack_channel":           "#some-slack-channel",
				"commit":                  gitSha,
				"commit_link":             "https://github.com/grafana/plugin-ci-workflows/commit/" + gitSha,
				"auto_merge_environments": "dev+ops+prod-canary+prod",
				"auto_approve_durations":  `{"dev":0,"ops":0,"prod-canary":null,"prod":null}`,
				"prod_targets_all":        "false",
			},
		},
		{
			name: "provisioned plugin with custom approval durations and auto-merge",
			inputs: cd.WorkflowInputs{
				TriggerArgo:                workflow.Input(true),
				GrafanaCloudDeploymentType: workflow.Input("provisioned"),
				Environment:                workflow.Input("prod"),
				ArgoWorkflowSlackChannel:   workflow.Input("#some-slack-channel"),
				AutoMergeEnvironments:      workflow.Input("dev,ops"),
				AutoApproveDurations:       workflow.Input(`{"dev":0,"ops":"1h","prod-canary":"24h","prod":"72h"}`),
				ProdTargetsAll:             workflow.Input(false),
			},
			expArgoShouldBeTriggered: true,
			expArgoInputs: map[string]string{
				"slug":                    "simple-frontend",
				"version":                 "1.0.0",
				"environment":             "prod",
				"slack_channel":           "#some-slack-channel",
				"commit":                  gitSha,
				"commit_link":             "https://github.com/grafana/plugin-ci-workflows/commit/" + gitSha,
				"auto_merge_environments": "dev+ops",
				"auto_approve_durations":  `{"dev":0,"ops":"1h","prod-canary":"24h","prod":"72h"}`,
				"prod_targets_all":        "false",
			},
		},

		// Special cases that don't trigger Argo Workflow
		{
			name: "trigger-argo=false doesn't trigger argo",
			inputs: cd.WorkflowInputs{
				TriggerArgo:                workflow.Input(false),
				GrafanaCloudDeploymentType: workflow.Input("provisioned"),
				Environment:                workflow.Input("dev"),
				AutoMergeEnvironments:      workflow.Input("dev"),
				ArgoWorkflowSlackChannel:   workflow.Input("#some-slack-channel"),
			},
			expArgoShouldBeTriggered: false,
		},
		{
			name: "environment=none doesn't trigger argo",
			inputs: cd.WorkflowInputs{
				TriggerArgo:                workflow.Input(true),
				GrafanaCloudDeploymentType: workflow.Input("provisioned"),
				Environment:                workflow.Input("none"),
				AutoMergeEnvironments:      workflow.Input("dev"),
				ArgoWorkflowSlackChannel:   workflow.Input("#some-slack-channel"),
			},
			expArgoShouldBeTriggered: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)
			wf, err := cd.NewWorkflow(
				cd.WithWorkflowInputs(tc.inputs),
				// Only run Argo Workflow trigger job, no CI or CD
				// Mock the CI output because the Argo Workflow trigger job needs its output.
				cd.MutateCDWorkflow().With(
					workflow.WithOnlyOneJob(t, "trigger-argo-workflow", false),
					// Mock direct dependencies of trigger-argo-workflow:
					workflow.WithNoOpJobWithOutputs(t, "publish-to-catalog", map[string]string{}),
					workflow.WithNoOpJobWithOutputs(t, "ci", map[string]string{
						// Simplified CI output for testing
						"plugin": `{"id": "simple-frontend", "version": "1.0.0"}`,
					}),
					// Indirect dependency: mock it so it doesn't run.
					workflow.WithNoOpJobWithOutputs(t, "upload-to-gcs-release", map[string]string{}),
				),
				// Mock Argo Workflow trigger using the runner's HTTPSpy
				cd.WithMockedArgoWorkflows(t, runner.Argo),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewPushEventPayload("main"))
			require.NoError(t, err)
			require.True(t, r.Success, "Argo Workflow trigger job should succeed")

			calls := runner.Argo.GetCalls()
			if !tc.expArgoShouldBeTriggered {
				// Argo should not be triggered
				require.Empty(t, calls, "expected no Argo Workflow trigger calls")
				require.Empty(t, r.Summary, "expected no summary")
				return
			}

			// Argo should be triggered
			// Verify the Argo Workflow trigger step received the expected inputs
			require.Len(t, calls, 1, "expected exactly one Argo Workflow trigger call")
			inputs := calls[0].Inputs
			require.Equal(t, "grafana-plugins-cd", inputs["namespace"])
			require.Equal(t, "grafana-plugins-deploy", inputs["workflow_template"])
			// Verify the "parameters" input (new-line-separated key=value pairs that are then passed to the Argo Workflow)
			argoInputs := make(map[string]string, len(tc.expArgoInputs))
			for _, line := range strings.Split(inputs["parameters"].(string), "\n") {
				key, value, _ := strings.Cut(line, "=")
				if key == "" {
					continue
				}
				argoInputs[key] = value
			}
			require.Equal(t, tc.expArgoInputs, argoInputs, "wrong argo inputs provided to argo workflow trigger step")

			// Verify summary
			require.Len(t, r.Summary, 1, "should have exactly one summary")
			require.Contains(t, r.Summary[0], "A deployment to Grafana Cloud via the plugins CD Argo Workflow has successfully been triggered.")
			require.Contains(t, r.Summary[0], "Plugin Version: `1.0.0`")
			require.Contains(t, r.Summary[0], "Environment(s): `"+*tc.inputs.Environment+"`")
			require.Contains(t, r.Summary[0], "**👉 You can follow the deployment [here](https://mock-argo-workflows.example.com/workflows/grafana-plugins-cd/mock-workflow-id)**")
		})
	}

	t.Run("unsupported grafana cloud deployment type doesn't trigger argo", func(t *testing.T) {
		t.Parallel()

		runner, err := act.NewRunner(t)
		require.NoError(t, err)
		wf, err := cd.NewWorkflow(
			cd.WithWorkflowInputs(cd.WorkflowInputs{
				TriggerArgo:                workflow.Input(true),
				GrafanaCloudDeploymentType: workflow.Input("foo bar baz"),
				Environment:                workflow.Input("dev"),
				AutoMergeEnvironments:      workflow.Input("dev"),
				ArgoWorkflowSlackChannel:   workflow.Input("#some-slack-channel"),
			}),
			// Only run Argo Workflow trigger job, no CI or CD
			cd.MutateCDWorkflow().With(
				workflow.WithOnlyOneJob(t, "trigger-argo-workflow", false),
				workflow.WithNoOpJobWithOutputs(t, "publish-to-catalog", map[string]string{}),
				workflow.WithNoOpJobWithOutputs(t, "ci", map[string]string{
					"plugin": `{"id": "simple-frontend", "version": "1.0.0"}`,
				}),
				workflow.WithNoOpJobWithOutputs(t, "upload-to-gcs-release", map[string]string{}),
			),
			cd.WithMockedArgoWorkflows(t, runner.Argo),
		)
		require.NoError(t, err)

		r, err := runner.Run(wf, act.NewPushEventPayload("main"))
		require.NoError(t, err)
		require.False(t, r.Success, "workflow should fail")
		require.Empty(t, runner.Argo.GetCalls(), "expected no Argo Workflow trigger calls")
		require.Empty(t, r.Summary, "expected no summary")
	})
}
