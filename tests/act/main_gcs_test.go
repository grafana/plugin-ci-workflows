package main

import (
	"path/filepath"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/stretchr/testify/require"
)

func TestGCS(t *testing.T) {
	commitHash, err := getGitCommitSHA()
	require.NoError(t, err)

	for _, tc := range []struct {
		folder     string
		id         string
		version    string
		hasBackend bool
	}{
		{
			folder:     "simple-frontend",
			id:         "grafana-simplefrontend-panel",
			version:    "1.0.0",
			hasBackend: false,
		},
		{
			folder:     "simple-backend",
			id:         "grafana-simplebackend-datasource",
			version:    "1.0.0",
			hasBackend: true,
		},
	} {
		t.Run(tc.folder, func(t *testing.T) {
			// t.Parallel()

			for _, event := range []act.EventPayload{
				act.NewPushEventPayload("main"),
				act.NewPullRequestEventPayload("feature-branch"),
			} {
				t.Run(event.Name(), func(t *testing.T) {
					runner, err := act.NewRunner(t)
					require.NoError(t, err)

					wf, err := workflow.NewSimpleCI(
						workflow.WithPluginDirectoryInput(filepath.Join("tests", tc.folder)),
						workflow.WithDistArtifactPrefixInput(tc.folder+"-"),

						// Disable some features to speed up the test
						workflow.WithPlaywrightInput(false),
						workflow.WithRunTruffleHogInput(false),
						workflow.WithRunPluginValidatorInput(false),

						// Mock dist so we don't spend time building the plugin
						workflow.WithMockedDist(t, tc.folder),
						// Mock a trusted context to enable GCS upload
						workflow.WithMockedWorkflowContext(t, workflow.Context{
							IsTrusted: true,
						}),
						// Mock all GCS access
						workflow.WithMockedGCS(t),

						// No-op steps that are normally executed in a trusted context
						// but are not relevant for this test and would error out otherwise.
						workflow.WithNoOpStep(t, "get-secrets"),
						workflow.WithNoOpStep(t, "generate-github-token"),
						workflow.WithAllowUnsignedInput(true),
					)
					require.NoError(t, err)

					r, err := runner.Run(wf, event)
					require.NoError(t, err)
					require.True(t, r.Success, "workflow should succeed")

					// Assert files uploaded to GCS (commit hash)
					anyZipFn := anyZipFileName(tc.id, tc.version)
					commitBasePath := filepath.Join("integration-artifacts", tc.id, tc.version, "main", commitHash, tc.folder+"-dist-artifacts")
					latestBasePath := filepath.Join("integration-artifacts", tc.id, tc.version, "main", "latest", tc.folder+"-dist-artifacts")

					// Expect commit hash any zip
					expFiles := []string{filepath.Join(commitBasePath, anyZipFn)}
					if event.IsPush() {
						// Also expect zips in the "latest" folder if the event is a push to main, rather than a PR
						expFiles = append(expFiles, filepath.Join(latestBasePath, anyZipFn))
					}
					if tc.hasBackend {
						// Expect backend zips
						for _, osArch := range osArchCombos {
							// Expect commit hash os/arch zip
							backendZipFn := osArchZipFileName(tc.id, tc.version, osArch)
							expFiles = append(expFiles, filepath.Join(commitBasePath, backendZipFn))

							if event.IsPush() {
								// Also latest os/arch zip
								expFiles = append(expFiles, filepath.Join(latestBasePath, backendZipFn))
							}
						}
					}

					// For each zip file, expect the corresponding .md5 and .sha1 files
					for _, fn := range expFiles {
						expFiles = append(expFiles, fn+".md5", fn+".sha1")
					}

					// Check files in mocked GCS
					err = checkFilesExist(runner.GCS.Fs, expFiles, checkFilesExistOptions{strict: true})
					require.NoErrorf(t, err, "wrong files uploaded to GCS (commit hash)")

					// TODO: assert job outputs (GCS URLs used later for publishing)
				})
			}
		})
	}
}
