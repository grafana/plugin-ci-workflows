// This file contains functions to create mocked jobs and steps for testing purposes.
package workflow

import (
	"encoding/json"
	"fmt"
)

// CopyMockFilesStep returns a Step that copies mock files from a source folder to a destination folder.
// The mock files are present in tests/act/mockdata in the repo, which is mounted into the act container at /mockdata.
// The sourceFolder is relative to /mockdata, e.g., "dist/simple-frontend".
// The destFolder is the destination path inside the GitHub Actions runner workspace.
// You can use GitHub Actions expressions in destFolder, e.g., "${{ github.workspace }}/plugins/my-plugin/dist".
func CopyMockFilesStep(sourceFolder string, destFolder string) Step {
	return Step{
		Name: "Copy mock dist files",
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
// The srcPath is the source path of the files to upload inside the GitHub Actions runner workspace.
// The destPath is the destination path inside the mocked GCS bucket (i.e., inside /gcs).
func MockGCSUploadStep(srcPath, destPath string) Step {
	return Step{
		Name: "Upload to GCS (mocked)",
		Run: Commands{
			"set -x",
			`mkdir -p /gcs/` + destPath,
			"cp -r " + srcPath + " /gcs/" + destPath,

			// For debugging
			"echo 'Mock GCS upload complete. Mock GCS bucket content:'",
			"ls -la /gcs/" + destPath,

			// Get list of all uploaded files, separate them by commas
			`files=$(find ` + srcPath + ` -type f | sed 's|^\./||' | tr '\n' ',' | sed 's/,$//')`,

			// Set output (simplified)
			`echo "uploaded=$files" >> "$GITHUB_OUTPUT"`,
		}.String(),
		Shell: "bash",
	}
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
