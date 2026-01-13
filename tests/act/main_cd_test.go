package main

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow/cd"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow/ci"
	"github.com/stretchr/testify/require"
)

func TestCD(t *testing.T) {
	runner, err := act.NewRunner(t)
	runner.GCOM.HandleFunc("GET /plugins/{pluginID}", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.WriteHeader(http.StatusOK)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"code": http.StatusOK,
		}))
	})
	require.NoError(t, err)
	wf, err := cd.NewWorkflow(
		cd.WithWorkflowInputs(cd.WorkflowInputs{
			CI: ci.WorkflowInputs{
				PluginDirectory:     workflow.Input(filepath.Join("tests", "simple-frontend")),
				DistArtifactsPrefix: workflow.Input("simple-frontend-"),
				Testing:             workflow.Input(false),
				AllowUnsigned:       workflow.Input(true),

				RunTruffleHog:      workflow.Input(false),
				RunPluginValidator: workflow.Input(false),
				RunPlaywright:      workflow.Input(false),
			},
			DisableDocsPublishing: workflow.Input(true),
			DisableGitHubRelease:  workflow.Input(true),
			Environment:           workflow.Input("dev"),
			// Branch:                     workflow.Input("main"),
			Scopes: workflow.Input("universal"),
			// GrafanaCloudDeploymentType: workflow.Input("provisioned"),
		}),
		cd.WithCIOptions(
			// TODO: also test with signature
			ci.WithMockedPackagedDistArtifacts(t, "dist/simple-frontend", "dist-artifacts-unsigned/simple-frontend"),
			ci.WithMockedWorkflowContext(t, ci.Context{IsTrusted: true}),
		),
		cd.WithMockedGCOM(runner.GCOM),
		cd.MutateAllWorkflows().With(
			workflow.WithMockedVault(t, workflow.VaultSecrets{}),
			workflow.WithMockedGitHubAppToken(t),
			workflow.WithMockedGCS(t),
		),
	)
	require.NoError(t, err)

	r, err := runner.Run(wf, act.NewPushEventPayload("main", act.WithEventActor("dependabot[bot]")))
	require.NoError(t, err)
	require.True(t, r.Success, "workflow should succeed")
}
