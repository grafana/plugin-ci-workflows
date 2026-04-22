package main

import (
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/stretchr/testify/require"
)

func TestCheckMatrixStatus(t *testing.T) {
	for _, tc := range []struct {
		name          string
		results       string
		expectSuccess bool
	}{
		{name: "success", results: "success", expectSuccess: true},
		{name: "failure", results: "failure", expectSuccess: false},
		{name: "cancelled", results: "cancelled", expectSuccess: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)

			wf := checkMatrixStatusWorkflow(tc.results)

			r, err := runner.Run(wf, act.NewPushEventPayload("main"))
			require.NoError(t, err)
			require.Equal(t, tc.expectSuccess, r.Success,
				"check-matrix-status with results=%q: expected success=%v", tc.results, tc.expectSuccess)
		})
	}
}

func checkMatrixStatusWorkflow(results string) *workflow.TestingWorkflow {
	baseWf := workflow.BaseWorkflow{
		Name: "Check Matrix Status Test",
		On: workflow.On{
			Push: workflow.OnPush{
				Branches: []string{"main"},
			},
		},
		Jobs: map[string]*workflow.Job{
			"check": {
				RunsOn: "ubuntu-arm64-small",
				Steps: workflow.Steps{
					{
						Name: "Check matrix status",
						Uses: "grafana/plugin-ci-workflows/actions/internal/check-matrix-status@main",
						With: map[string]any{
							"results": results,
						},
					},
				},
			},
		},
	}
	wf := workflow.NewTestingWorkflow("check-matrix-status", baseWf)
	wf.AddUUIDToAllJobsRecursive()
	return wf
}
