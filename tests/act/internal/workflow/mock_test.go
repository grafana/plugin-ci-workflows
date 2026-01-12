package workflow

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMockGCSUploadStep(t *testing.T) {
	step := Step{
		Name: "Upload GCS",
		Uses: "google-github-actions/upload-cloud-storage@6397bd7208e18d13ba2619ee21b9873edc94427a",
		With: map[string]any{
			"path":        "/tmp/dist-artifacts",
			"destination": "integration-artifacts/grafana-foo-plugin/folder",
		},
	}
	mockedStep, err := MockGCSUploadStep(step)
	require.NoError(t, err)
	require.Equal(t, "Upload GCS (mocked)", mockedStep.Name)
	require.Contains(t, mockedStep.Run, `echo "uploaded=$files" >> "$GITHUB_OUTPUT"`)
	require.Contains(t, mockedStep.Run, `mkdir -p /gcs/integration-artifacts/grafana-foo-plugin/folder`)
	require.Contains(t, mockedStep.Run, `cp -r /tmp/dist-artifacts /gcs/integration-artifacts/grafana-foo-plugin/folder`)
}

func TestMockVaultSecretsStep(t *testing.T) {
	step := Step{
		Name: "Get Vault Secrets",
		Uses: "grafana/shared-workflows/actions/get-vault-secrets@a37de51f3d713a30a9e4b21bcdfbd38170020593",
	}
	mockedStep, err := MockVaultSecretsStep(step, VaultSecrets{"secret1": "value1", "secret2": "value2"})
	require.NoError(t, err)
	require.Equal(t, "Get Vault Secrets (mocked)", mockedStep.Name)
	require.Contains(t, mockedStep.Run, `echo "secrets=$SECRETS_JSON" >> "$GITHUB_OUTPUT"`)
	require.Equal(t, `{"secret1":"value1","secret2":"value2"}`, mockedStep.Env["SECRETS_JSON"])
}

func TestMockArgoWorkflowStep(t *testing.T) {
	step := Step{
		Name: "Trigger Argo Workflow",
		Uses: "grafana/shared-workflows/actions/trigger-argo-workflow@e100806688f1209051080dfea5719fbbd1d18cc0",
	}
	mockedStep, err := MockArgoWorkflowStep(step)
	require.NoError(t, err)
	require.Equal(t, "Trigger Argo Workflow (mocked)", mockedStep.Name)
	require.Contains(t, mockedStep.Run, `echo "uri=https://mock-argo-workflows.example.com/workflows/grafana-plugins-cd/mock-workflow-id" >> "$GITHUB_OUTPUT"`)
}
