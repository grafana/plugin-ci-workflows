// This file contains functions to create mocked jobs and steps for testing purposes.
package workflow

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

const (
	GCSLoginAction  = "google-github-actions/auth"
	GCSUploadAction = "google-github-actions/upload-cloud-storage"

	VaultSecretsAction = "grafana/shared-workflows/actions/get-vault-secrets"
	ArgoWorkflowAction = "grafana/shared-workflows/actions/trigger-argo-workflow"

	GitHubAppTokenAction = "actions/create-github-app-token"
)

// CopyMockFilesStep returns a Step that copies mock files from a source folder to a destination folder.
// The mock files are present in tests/act/mockdata in the repo, which is mounted into the act container at /mockdata.
// The sourceFolder is relative to /mockdata, e.g., "dist/simple-frontend".
// The destFolder is the destination path inside the GitHub Actions runner workspace.
// You can use GitHub Actions expressions in destFolder, e.g., "${{ github.workspace }}/plugins/my-plugin/dist".
func CopyMockFilesStep(sourceFolder string, destFolder string) Step {
	return Step{
		Name: "Copy mock files",
		Run: Commands{
			"set -x",
			"mkdir -p " + destFolder,
			"cp -r /mockdata/" + sourceFolder + "/. " + destFolder,
			"cd " + destFolder,
			"ls -la",
		}.String(),
		Shell: "bash",
	}
}

// NoOpStep returns a Step that does nothing (no-op) for testing purposes.
// The step simply echoes a message indicating it is a no-op step.
func NoOpStep(originalStep Step) Step {
	nameOrID := originalStep.Name
	if nameOrID == "" {
		nameOrID = originalStep.ID
	}
	return Step{
		Name:  nameOrID + " (no-opp'ed for testing)",
		ID:    originalStep.ID,
		Run:   "echo 'noop-ed step for testing'",
		Shell: "bash",
	}
}

// MockGCSUploadStep returns a Step that mocks uploading files to Google Cloud Storage (GCS).
// Instead of actually uploading to GCS, it copies files to a local folder mounted into the act container at /gcs.
// The originalStep parameter is the original GCS upload step to be mocked.
// The original step must use the `google-github-actions/upload-cloud-storage` action
// and have valid "path" and "destination" inputs.
// If those conditions are not met, an error is returned.
func MockGCSUploadStep(originalStep Step) (Step, error) {
	// Make sure the original step is indeed a GCS upload step
	if !strings.HasPrefix(originalStep.Uses, GCSUploadAction) {
		return Step{}, fmt.Errorf("cannot mock gcs for a step that uses %q action, must be %q", originalStep.Uses, GCSUploadAction)
	}

	// Extract the existing inputs and use them in the mocked bash step.
	// Make sure they are strings and not empty
	srcPath, ok1 := originalStep.With["path"].(string)
	destPath, ok2 := originalStep.With["destination"].(string)
	if srcPath == "" || destPath == "" || !ok1 || !ok2 {
		return Step{}, fmt.Errorf("could not mock gcs step %q (id: %q) because inputs are not valid", originalStep.Name, originalStep.ID)
	}

	return Step{
		Name: originalStep.Name + " (mocked)",
		Run: Commands{
			"set -x",
			`mkdir -p /gcs/${DEST_PATH}`,

			// Handle both file and directory srcPath
			`if [ -f "${SRC_PATH}" ]; then`,
			// srcPath is a file: copy directly
			`  cp "${SRC_PATH}" /gcs/${DEST_PATH}/`,
			`  filename=$(basename "${SRC_PATH}")`,
			`  files="${DEST_PATH}/${filename}"`,
			`  files=$(echo "$files" | cut -d'/' -f2-)`,
			`else`,
			// srcPath is a directory: copy recursively
			`  cp -r "${SRC_PATH}" /gcs/${DEST_PATH}`,
			`  cd "${SRC_PATH}"`,
			// Get a list of all uploaded files, separated by commas.
			// Find all files, prepend destPath, remove leading ./, get relative path (remove bucket name after `/gcs`), join with commas
			`  files=$(find . -type f | sed 's|^\./|${DEST_PATH}/|' | cut -d'/' -f2- | tr '\n' ',' | sed 's/,$//')`,
			`fi`,

			// For debugging
			"echo 'Mock GCS upload complete. Mock GCS bucket content:'",
			"find /gcs -type f",

			// Set output (simplified)
			`echo "uploaded=$files" >> "$GITHUB_OUTPUT"`,
		}.String(),
		Shell: "bash",
		Env: map[string]string{
			"SRC_PATH":  srcPath,
			"DEST_PATH": destPath,
		},
	}, nil
}

// LocalMockdataPath returns the full path to a file or folder inside tests/act/mockdata
// used for accessing mock data locally, outside of the act container.
func LocalMockdataPath(parts ...string) string {
	return filepath.Join("tests", "act", "mockdata", filepath.Join(parts...))
}

// VaultSecrets allows defining the secret values that the mocked get-vault-secrets step should return.
// The keys in the maps must match the secret reference used in the workflow step inputs
// (the value on the right side of the equals sign).
//
// Example:
//
//	If the workflow has:
//	  common_secrets: |
//	    MY_SECRET=secret/path:key
//
//	Then the VaultSecrets should have:
//	  CommonSecrets: map[string]string{
//	    "secret/path:key": "mock-value",
//	  }
type VaultSecrets struct {
	// CommonSecrets contains secrets that are referenced in the 'common_secrets' input.
	CommonSecrets map[string]string
	// RepoSecrets contains secrets that are referenced in the 'repo_secrets' input.
	RepoSecrets map[string]string

	// DefaultValue is the default value to use for secrets that are not defined in CommonSecrets or RepoSecrets.
	// If nil, the step will fail to be constructed if a secret is not defined.
	// If not nil, this value will be used for secrets that are not defined.
	DefaultValue *string
}

// MockVaultSecretsStep returns a Step that mocks the grafana/shared-workflows/actions/get-vault-secrets action.
// Instead of actually fetching secrets from Vault, it outputs the provided secrets in the expected format.
// The originalStep parameter is the original Vault secrets step to be mocked.
// The original step must use the `grafana/shared-workflows/actions/get-vault-secrets` action.
// If those conditions are not met, an error is returned.
//
// The mocked step mimics the behavior of the original step, but instead of fetching secrets from Vault,
// it outputs the provided secrets in the expected format.
// If export_env is true, the secrets are exported as environment variables.
// If export_env is false, the secrets are exported as a JSON object.
//
// If a secret is not found in the provided VaultSecrets:
//   - If DefaultValue is nil, an error is returned.
//   - If DefaultValue is not nil, the DefaultValue is used.
func MockVaultSecretsStep(originalStep Step, secrets VaultSecrets) (Step, error) {
	// Make sure the original step is indeed a Vault secrets step
	if !strings.HasPrefix(originalStep.Uses, VaultSecretsAction) {
		return Step{}, fmt.Errorf("cannot mock vault secrets for a step that uses %q action, must be %q", originalStep.Uses, VaultSecretsAction)
	}

	// Extract original inputs with safe type assertions and defaults
	var commonSecretsInput, repoSecretsInput string
	if v, ok := originalStep.With["common_secrets"].(string); ok {
		commonSecretsInput = v
	}
	if v, ok := originalStep.With["repo_secrets"].(string); ok {
		repoSecretsInput = v
	}
	exportEnvInput := true
	if v, ok := originalStep.With["export_env"].(bool); ok {
		exportEnvInput = v
	}
	output := map[string]string{}
	for _, s := range []struct {
		input   string
		secrets map[string]string
	}{
		{commonSecretsInput, secrets.CommonSecrets},
		{repoSecretsInput, secrets.RepoSecrets},
	} {
		for i, line := range strings.Split(s.input, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.Split(line, "=")
			if len(parts) != 2 {
				return Step{}, fmt.Errorf("invalid input, not enough parts on line %d: %s", i, line)
			}
			outputName := strings.TrimSpace(parts[0])
			secretReference := strings.TrimSpace(parts[1])
			secretValue, ok := s.secrets[secretReference]
			if !ok && secrets.DefaultValue == nil {
				return Step{}, fmt.Errorf("secret reference %q not found in provided fake secrets %+v", secretReference, secrets)
			}
			if !ok {
				secretValue = *secrets.DefaultValue
			}
			output[outputName] = secretValue
		}
	}

	step := Step{
		ID:    originalStep.ID,
		Name:  originalStep.Name + " (mocked)",
		Env:   map[string]string{},
		Shell: "bash",
	}
	var stepCommands Commands
	if exportEnvInput {
		// Workflow-level env vars output
		for k, v := range output {
			stepCommands = append(stepCommands, fmt.Sprintf(`echo "%s=%s" >> "$GITHUB_ENV"`, k, v))
		}
	} else {
		// JSON output
		secretsJSON, err := json.Marshal(output)
		if err != nil {
			return Step{}, fmt.Errorf("marshal vault secrets to json: %w", err)
		}
		stepCommands = append(stepCommands, `echo "secrets=${SECRETS_JSON}" >> "$GITHUB_OUTPUT"`)
		step.Env = map[string]string{"SECRETS_JSON": string(secretsJSON)}
	}
	step.Run = stepCommands.String()
	return step, nil
}

// MockArgoWorkflowStep returns a Step that mocks the grafana/shared-workflows/actions/trigger-argo-workflow action.
// Instead of actually triggering an Argo Workflow, it outputs a mock URI for the workflow.
// The originalStep parameter is the original Argo Workflow trigger step to be mocked.
// The original step must use the `grafana/shared-workflows/actions/trigger-argo-workflow` action.
// If those conditions are not met, an error is returned.
//
// The mocked step outputs the `uri` output expected by subsequent steps.
func MockArgoWorkflowStep(originalStep Step) (Step, error) {
	// Make sure the original step is indeed an Argo Workflow trigger step
	if !strings.HasPrefix(originalStep.Uses, ArgoWorkflowAction) {
		return Step{}, fmt.Errorf("cannot mock argo workflow for a step that uses %q action, must be %q", originalStep.Uses, ArgoWorkflowAction)
	}

	return Step{
		Name: originalStep.Name + " (mocked)",
		Run: Commands{
			`echo "Mocking Argo Workflow trigger step"`,
			`echo "uri=https://mock-argo-workflows.example.com/workflows/grafana-plugins-cd/mock-workflow-id" >> "$GITHUB_OUTPUT"`,
		}.String(),
		Shell: "bash",
	}, nil
}

func MockGitHubAppTokenStep(originalStep Step, token string) (Step, error) {
	if !strings.HasPrefix(originalStep.Uses, GitHubAppTokenAction) {
		return Step{}, fmt.Errorf("cannot mock github app token for a step that uses %q action, must be %q", originalStep.Uses, GitHubAppTokenAction)
	}
	return Step{
		Name: originalStep.Name + " (mocked)",
		Run: Commands{
			`echo "Mocking GitHub app token step"`,
			`echo "token=${MOCK_TOKEN}" >> "$GITHUB_OUTPUT"`,
		}.String(),
		Shell: "bash",
		Env: map[string]string{
			"MOCK_TOKEN": token,
		},
	}, nil
}
