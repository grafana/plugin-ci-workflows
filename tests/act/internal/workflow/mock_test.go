package workflow

import (
	"strings"
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
	require.Contains(t, mockedStep.Run, `echo "uploaded=$files" >> "$GITHUB_OUTPUT"`, "should contain echo to output")
	require.Contains(t, mockedStep.Run, `mkdir -p /gcs/${DEST_PATH}`, "should contain bucket folder creation")
	require.Contains(t, mockedStep.Run, `  cp -r "${SRC_PATH}" /gcs/${DEST_PATH}`, "should contain cp command")
	require.Equal(t, map[string]string{
		"DEST_PATH": "integration-artifacts/grafana-foo-plugin/folder",
		"SRC_PATH":  "/tmp/dist-artifacts",
	}, mockedStep.Env, "should have correct env vars")
}

func TestMockVaultSecretsStep(t *testing.T) {
	const (
		vaultAction = "grafana/shared-workflows/actions/get-vault-secrets@a37de51f3d713a30a9e4b21bcdfbd38170020593"
		bashOutput  = `echo "secrets=${SECRETS_JSON}" >> "$GITHUB_OUTPUT"`
	)
	vault := VaultSecrets{
		CommonSecrets: map[string]string{
			"common_secret_1:a": "value1",
			"common_secret_2:b": "value2",
		},
		RepoSecrets: map[string]string{
			"repo_secret_1:c": "value3",
			"repo_secret_2:d": "value4",
		},
	}

	t.Run("common secrets", func(t *testing.T) {
		step := Step{
			Name: "Get Vault Secrets",
			Uses: vaultAction,
			With: map[string]any{
				"common_secrets": strings.Join([]string{
					"SECRET1=common_secret_1:a",
					"SECRET2=common_secret_2:b",
				}, "\n"),
				"export_env": false,
			},
		}
		mockedStep, err := MockVaultSecretsStep(step, vault)
		require.NoError(t, err)
		require.Equal(t, "Get Vault Secrets (mocked)", mockedStep.Name)
		require.Contains(t, mockedStep.Run, bashOutput)
		exp := `{"SECRET1":"value1","SECRET2":"value2"}`
		require.Equal(t, exp, mockedStep.Env["SECRETS_JSON"])
	})

	t.Run("repo secrets", func(t *testing.T) {
		step := Step{
			Name: "Get Vault Secrets",
			Uses: vaultAction,
			With: map[string]any{
				"repo_secrets": strings.Join([]string{
					"C=repo_secret_1:c",
					"D=repo_secret_2:d",
				}, "\n"),
				"export_env": false,
			},
		}
		mockedStep, err := MockVaultSecretsStep(step, vault)
		require.NoError(t, err)
		require.Equal(t, "Get Vault Secrets (mocked)", mockedStep.Name)
		require.Contains(t, mockedStep.Run, bashOutput)
		exp := `{"C":"value3","D":"value4"}`
		require.Equal(t, exp, mockedStep.Env["SECRETS_JSON"])
	})

	t.Run("common + repo secrets", func(t *testing.T) {
		step := Step{
			Name: "Get Vault Secrets",
			Uses: vaultAction,
			With: map[string]any{
				"common_secrets": strings.Join([]string{
					"SECRET1=common_secret_1:a",
					"SECRET2=common_secret_2:b",
				}, "\n"),
				"repo_secrets": strings.Join([]string{
					"C=repo_secret_1:c",
					"D=repo_secret_2:d",
				}, "\n"),
				"export_env": false,
			},
		}
		mockedStep, err := MockVaultSecretsStep(step, vault)
		require.NoError(t, err)
		require.Equal(t, "Get Vault Secrets (mocked)", mockedStep.Name)
		require.Contains(t, mockedStep.Run, bashOutput)
		exp := `{"C":"value3","D":"value4","SECRET1":"value1","SECRET2":"value2"}`
		require.Equal(t, exp, mockedStep.Env["SECRETS_JSON"])
	})

	t.Run("unexisting secret", func(t *testing.T) {
		step := Step{
			Name: "Get Vault Secrets",
			Uses: vaultAction,
			With: map[string]any{
				"common_secrets": strings.Join([]string{
					"SECRET1=this_secret_does_not_exist:a",
				}, "\n"),
				"export_env": false,
			},
		}
		_, err := MockVaultSecretsStep(step, vault)
		require.ErrorContains(t, err, "this_secret_does_not_exist")
		require.ErrorContains(t, err, "not found in provided fake secrets")
	})

	t.Run("unexisting secret with default value", func(t *testing.T) {
		step := Step{
			Name: "Get Vault Secrets",
			Uses: vaultAction,
			With: map[string]any{
				"common_secrets": strings.Join([]string{
					"SECRET1=a:b",
					"SECRET2=this_secret_does_not_exist:a",
				}, "\n"),
				"export_env": false,
			},
		}
		defaultValue := "foo"
		mockedStep, err := MockVaultSecretsStep(step, VaultSecrets{
			CommonSecrets: map[string]string{
				"a:b": "c",
			},
			DefaultValue: &defaultValue,
		})
		require.NoError(t, err)
		require.Equal(t, `{"SECRET1":"c","SECRET2":"foo"}`, mockedStep.Env["SECRETS_JSON"])
	})
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
