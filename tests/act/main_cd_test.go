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
	const pluginID = "simple-frontend"
	gitSha, err := getGitCommitSHA()
	require.NoError(t, err)

	assertHeadersAndAuth := func(t *testing.T, r *http.Request) {
		expHeaders := http.Header{
			"Accept":        []string{"application/json"},
			"User-Agent":    []string{"github-actions-shared-workflows:/plugins/publish"},
			"Authorization": []string{"Bearer dummy-gcom-api-key-dev"},
		}
		if r.Header.Get("X-Api-Key") != "" {
			expHeaders.Set("Authorization", "Bearer dummy-iap-token")
			expHeaders.Set("X-Api-Key", "dummy-gcom-api-key-dev")
		} else {
			expHeaders.Set("Authorization", "Bearer dummy-gcom-api-key-dev")
		}
		require.Subset(t, r.Header, expHeaders)
	}

	runner, err := act.NewRunner(t)
	runner.GCOM.HandleFunc("GET /api/plugins/{pluginID}", func(w http.ResponseWriter, r *http.Request) {
		assertHeadersAndAuth(t, r)

		w.WriteHeader(http.StatusOK)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"id":     1,
			"status": "active",
			"slug":   pluginID,
		}))
	})

	runner.GCOM.HandleFunc("POST /api/plugins", func(w http.ResponseWriter, r *http.Request) {
		assertHeadersAndAuth(t, r)

		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body), "should be able to decode json body")
		require.Equal(t, []any{"universal"}, body["scopes"], "should have correct scopes")
		require.Equal(t, false, body["pending"], "should have correct pending status")
		require.Equal(t, "https://github.com/grafana/plugin-ci-workflows", body["url"], "should have correct url")
		require.Equal(t, gitSha, body["commit"], "should have correct commit sha")
		require.Equal(t, map[string]any{
			"any": map[string]any{
				"url": "https://storage.googleapis.com/integration-artifacts/grafana-simplefrontend-panel/release/1.0.0/any/grafana-simplefrontend-panel-1.0.0.zip",
			},
		}, body["download"], "should have correct download URLs")

		w.WriteHeader(http.StatusOK)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"plugin": map[string]any{"id": 1337},
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

		// Mocks
		cd.WithMockedGCOM(runner.GCOM),
		cd.MutateAllWorkflows().With(
			workflow.WithMockedVault(t, workflow.VaultSecrets{
				DefaultValue: newPointer(""),
				CommonSecrets: map[string]string{
					"plugins/gcom-publish-token:dev":  "dummy-gcom-api-key-dev",
					"plugins/gcom-publish-token:ops":  "dummy-gcom-api-key-ops",
					"plugins/gcom-publish-token:prod": "dummy-gcom-api-key-prod",
				},
			}),
			workflow.WithMockedGitHubAppToken(t),
			workflow.WithMockedGCS(t),
		),

		// CD mutations to make it work in tests
		cd.MutateCDWorkflow().With(
			// Mock GCS artifacts exist safety check
			workflow.WithNoOpStep(t, "upload-to-gcs-release", "gcloud-sdk"),
			workflow.WithReplacedStep(t, "upload-to-gcs-release", "gcs_artifacts_exist", workflow.Step{
				Run:   workflow.Commands{`echo "gcs_artifacts_exist=false" >> $GITHUB_OUTPUT`}.String(),
				Shell: "bash",
			}),

			// Mock IAP token for GCOM API calls
			workflow.WithReplacedStep(t, "publish-to-catalog", "gcloud", workflow.Step{
				Run:   workflow.Commands{`echo "id_token=dummy-iap-token" >> $GITHUB_OUTPUT`}.String(),
				Shell: "bash",
			}),

			// TODO: done for simplicity now, remove later
			// workflow.WithNoOpStep(t, "publish-to-catalog", "check-and-create-stub"),

			// TODO: implement
			workflow.WithNoOpStep(t, "publish-to-catalog", "check-artifact-zips"),
		),
	)
	require.NoError(t, err)

	r, err := runner.Run(wf, act.NewPushEventPayload("main", act.WithEventActor("dependabot[bot]")))
	require.NoError(t, err)
	/* o, ok := r.Outputs.Get("setup", "vars", "environments")
	require.True(t, ok)
	t.Logf("the pif %+v", o) */
	require.True(t, r.Success, "workflow should succeed")
}
