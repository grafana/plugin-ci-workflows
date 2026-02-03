package workflow

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMockGCSUploadStep(t *testing.T) {
	t.Run("basic upload without glob", func(t *testing.T) {
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
		require.NotContains(t, mockedStep.Run, "globstar", "should not use globstar without glob pattern")
		require.Equal(t, map[string]string{
			"DEST_PATH":    "integration-artifacts/grafana-foo-plugin/folder",
			"SRC_PATH":     "/tmp/dist-artifacts",
			"GLOB_PATTERN": "",
			"FOLDER_NAME":  "dist-artifacts/",
		}, mockedStep.Env, "should have correct env vars")
	})

	t.Run("upload with glob pattern", func(t *testing.T) {
		step := Step{
			Name: "Upload GCS",
			Uses: "google-github-actions/upload-cloud-storage@6397bd7208e18d13ba2619ee21b9873edc94427a",
			With: map[string]any{
				"path":        "/tmp/dist-artifacts",
				"destination": "integration-artifacts/grafana-foo-plugin/folder",
				"glob":        "**/*.zip",
			},
		}
		mockedStep, err := MockGCSUploadStep(step)
		require.NoError(t, err)
		require.Equal(t, "Upload GCS (mocked)", mockedStep.Name)
		require.Contains(t, mockedStep.Run, `echo "uploaded=$files" >> "$GITHUB_OUTPUT"`, "should contain echo to output")
		require.Contains(t, mockedStep.Run, `mkdir -p /gcs/${DEST_PATH}`, "should contain bucket folder creation")
		require.Contains(t, mockedStep.Run, "shopt -s globstar nullglob", "should enable globstar")
		require.Contains(t, mockedStep.Run, `for file in ${GLOB_PATTERN}`, "should iterate over glob matches")
		require.NotContains(t, mockedStep.Run, `cp -r "${SRC_PATH}"`, "should not use recursive cp with glob")
		require.Equal(t, map[string]string{
			"DEST_PATH":    "integration-artifacts/grafana-foo-plugin/folder",
			"SRC_PATH":     "/tmp/dist-artifacts",
			"GLOB_PATTERN": "**/*.zip",
			"FOLDER_NAME":  "dist-artifacts/",
		}, mockedStep.Env, "should have correct env vars including glob pattern")
	})

	t.Run("upload with glob and parent=false", func(t *testing.T) {
		step := Step{
			Name: "Upload GCS",
			Uses: "google-github-actions/upload-cloud-storage@6397bd7208e18d13ba2619ee21b9873edc94427a",
			With: map[string]any{
				"path":        "/tmp/dist-artifacts",
				"destination": "integration-artifacts/grafana-foo-plugin/folder",
				"glob":        "*.txt",
				"parent":      false,
			},
		}
		mockedStep, err := MockGCSUploadStep(step)
		require.NoError(t, err)
		require.Contains(t, mockedStep.Run, "shopt -s globstar nullglob", "should enable globstar")
		require.Equal(t, map[string]string{
			"DEST_PATH":    "integration-artifacts/grafana-foo-plugin/folder",
			"SRC_PATH":     "/tmp/dist-artifacts",
			"GLOB_PATTERN": "*.txt",
			"FOLDER_NAME":  "",
		}, mockedStep.Env, "should have empty FOLDER_NAME when parent=false")
	})
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
		With: map[string]any{
			"namespace":         "grafana-plugins-cd",
			"workflow_template": "grafana-plugins-deploy",
			"parameters":        "slug=test-plugin\nversion=1.0.0",
		},
	}
	mockServerURL := "http://host.docker.internal:12345"
	mockedStep, err := MockArgoWorkflowStep(step, mockServerURL)
	require.NoError(t, err)
	// Note: The mocked name is set by MockAllStepsUsingAction, not by MockArgoWorkflowStep directly

	// Verify the step builds JSON from env vars and POSTs to the mock server
	require.Contains(t, mockedStep.Run, `INPUTS_JSON=$(echo "${HTTPSPY_INPUT_KEYS}"`)
	require.Contains(t, mockedStep.Run, `curl -s -X POST`)
	// Verify the step parses JSON response and sets outputs using jq
	require.Contains(t, mockedStep.Run, `jq -r 'to_entries[]`)
	require.Contains(t, mockedStep.Run, `>> "$GITHUB_OUTPUT"`)

	// Verify each input is passed as a separate env var (values will be evaluated at runtime)
	// Keys are preserved as-is (no uppercase/underscore transformation)
	require.Equal(t, "grafana-plugins-cd", mockedStep.Env["HTTPSPY_INPUT_namespace"])
	require.Equal(t, "grafana-plugins-deploy", mockedStep.Env["HTTPSPY_INPUT_workflow_template"])
	require.Equal(t, "slug=test-plugin\nversion=1.0.0", mockedStep.Env["HTTPSPY_INPUT_parameters"])

	// Verify input keys are passed so bash can build JSON
	require.Contains(t, mockedStep.Env["HTTPSPY_INPUT_KEYS"], `"namespace"`)
	require.Contains(t, mockedStep.Env["HTTPSPY_INPUT_KEYS"], `"workflow_template"`)
	require.Contains(t, mockedStep.Env["HTTPSPY_INPUT_KEYS"], `"parameters"`)

	require.Equal(t, mockServerURL, mockedStep.Env["HTTPSPY_MOCK_SERVER_URL"])
}
