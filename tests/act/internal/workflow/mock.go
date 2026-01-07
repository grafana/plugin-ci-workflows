// This file contains functions to create mocked jobs and steps for testing purposes.
package workflow

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

const (
	gcsLoginAction  = "google-github-actions/auth"
	gcsUploadAction = "google-github-actions/upload-cloud-storage"
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
func NoOpStep(id string) Step {
	return Step{
		Name:  id + " (no-opp'ed for testing)",
		ID:    id,
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
	if !strings.HasPrefix(originalStep.Uses, gcsUploadAction) {
		return Step{}, fmt.Errorf("cannot mock gcs for a step that uses %q action, must be %q", originalStep.Uses, gcsUploadAction)
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
			`mkdir -p /gcs/` + destPath,
			"cp -r " + srcPath + " /gcs/" + destPath,

			// For debugging
			"echo 'Mock GCS upload complete. Mock GCS bucket content:'",
			"find /gcs -type f",
			"cd " + srcPath,

			// Get a list of all uploaded files, separated by commas.
			// Find all files, prepend destPath, remove leading ./, get relative path (remove bucket name after `/gcs`), join with commas
			`files=$(find . -type f | sed 's|^\./|` + destPath + `/|' | cut -d'/' -f2- | tr '\n' ',' | sed 's/,$//')`,

			// Set output (simplified)
			`echo "uploaded=$files" >> "$GITHUB_OUTPUT"`,
		}.String(),
		Shell: "bash",
	}, nil
}

// MockWorkflowContextStep returns a Step that mocks the "workflow-context" step
// to return the given mocked Context.
func MockWorkflowContextStep(ctx Context) (Step, error) {
	ctxJSON, err := json.Marshal(ctx)
	if err != nil {
		return Step{}, fmt.Errorf("marshal workflow context to json: %w", err)
	}
	return Step{
		Name: "Determine workflow context (mocked)",
		Run: Commands{
			`echo "result=$RESULT" >> "$GITHUB_OUTPUT"`,
		}.String(),
		Env: map[string]string{
			"RESULT": string(ctxJSON),
		},
		Shell: "bash",
	}, nil
}

// localMockdataPath returns the full path to a file or folder inside tests/act/mockdata
// used for accessing mock data locally, outside of the act container.
func localMockdataPath(parts ...string) string {
	return filepath.Join("tests", "act", "mockdata", filepath.Join(parts...))
}
