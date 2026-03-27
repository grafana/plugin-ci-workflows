package main

import (
	"path/filepath"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/stretchr/testify/require"
)

// newPlaywrightCloudWorkflow creates a testing workflow that wraps playwright-cloud.yml,
// applies the given mutations to the child workflow, and passes through any caller-provided inputs.
func newPlaywrightCloudWorkflow(t *testing.T, inputs map[string]any, childOpts ...workflow.TestingWorkflowOption) (*workflow.TestingWorkflow, error) {
	t.Helper()

	childBaseWf, err := workflow.NewBaseWorkflowFromFile(filepath.Join(".github", "workflows", "playwright-cloud.yml"))
	if err != nil {
		return nil, err
	}

	// Parent workflow: push-triggered, single job that calls playwright-cloud.yml.
	parentBaseWf := workflow.BaseWorkflow{
		Name: "Playwright Cloud",
		On: workflow.On{
			Push: workflow.OnPush{
				Branches: []string{"main"},
			},
		},
		Jobs: map[string]*workflow.Job{
			"playwright-cloud": {
				Name: "Playwright Cloud",
				Permissions: workflow.Permissions{
					"contents": "read",
					"id-token": "write",
				},
				With: inputs,
			},
		},
	}

	parentWf := workflow.NewTestingWorkflow("simple-playwright-cloud", parentBaseWf)
	childWf := workflow.NewTestingWorkflow("playwright-cloud", childBaseWf, childOpts...)
	parentWf.AddChild("playwright-cloud", childWf)

	// Point the parent job at the mocked child workflow.
	parentWf.Jobs()["playwright-cloud"].Uses = workflow.PCIWFBaseRef + "/" + childWf.FileName() + "@main"

	parentWf.AddUUIDToAllJobsRecursive()
	return parentWf, nil
}

// mockVaultForPlaywrightCloud returns mock Vault secrets that satisfy all
// common_secrets declared in the playwright-cloud.yml get-vault-secrets step.
func mockVaultForPlaywrightCloud() workflow.VaultSecrets {
	return workflow.VaultSecrets{
		CommonSecrets: map[string]string{
			"data-sources/e2e:grafana-pw":       "mock-grafana-password",
			"data-sources/e2e:grafana-username": "mock-grafana-user",
			"grafana-bench:prometheus_token":    "mock-prom-token",
			"grafana-bench:prometheus_url":      "http://mock-prometheus",
			"grafana-bench:prometheus_user":     "mock-prom-user",
		},
	}
}

func TestPlaywrightCloud(t *testing.T) {
	type tc struct {
		name        string
		inputs      map[string]any
		event       act.Event
		expectSkip  bool
	}

	for _, tc := range []tc{
		{
			name:       "bench_tests_run_in_grafana_org",
			inputs:     map[string]any{},
			event:      act.NewPushEventPayload("main"),
			expectSkip: false,
		},
		{
			name:       "bench_tests_skipped_outside_grafana_org",
			inputs:     map[string]any{},
			event:      act.NewPushEventPayload("main", act.WithNonGrafanaOwner()),
			expectSkip: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)

			wf, err := newPlaywrightCloudWorkflow(t, tc.inputs,
				// Mock Vault so the get-secrets step doesn't need real OIDC.
				workflow.WithMockedVault(t, mockVaultForPlaywrightCloud()),
				// No-op the wait-for-grafana action: no live Cloud instance in tests.
				workflow.WithNoOpStep(t, "bench-tests", "wait-for-grafana"),
				// No-op the docker run step: no Bench image or real Playwright tests.
				workflow.WithNoOpStep(t, "bench-tests", "run-bench-tests"),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, tc.event)
			require.NoError(t, err)

			if tc.expectSkip {
				// When the job is skipped the overall workflow still succeeds.
				require.True(t, r.Success, "workflow should succeed even when bench-tests job is skipped")
			} else {
				require.True(t, r.Success, "workflow should succeed")
			}
		})
	}
}
