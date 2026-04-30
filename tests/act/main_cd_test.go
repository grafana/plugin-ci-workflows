package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow/cd"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow/ci"
	"github.com/stretchr/testify/require"
)

func TestCD(t *testing.T) {
	gitSha, err := getGitCommitSHA()
	require.NoError(t, err)

	const (
		pluginVersion = "1.0.0"

		fakeGcomTokenDev  = "dummy-gcom-api-key-dev"
		fakeGcomTokenOps  = "dummy-gcom-api-key-ops"
		fakeGcomTokenProd = "dummy-gcom-api-key-prod"

		fakeIapToken = "dummy-iap-token"
	)

	mockVault := workflow.VaultSecrets{
		DefaultValue: newPointer(""),
		CommonSecrets: map[string]string{
			"plugins/gcom-publish-token:dev":  fakeGcomTokenDev,
			"plugins/gcom-publish-token:ops":  fakeGcomTokenOps,
			"plugins/gcom-publish-token:prod": fakeGcomTokenProd,
		},
	}

	for _, tc := range []struct {
		name         string
		pluginFolder string
		pluginSlug   string
		hasBackend   bool
	}{
		{
			name:         "simple-frontend",
			pluginFolder: "simple-frontend",
			pluginSlug:   "grafana-simplefrontend-panel",
			hasBackend:   false,
		},
		{
			name:         "simple-backend",
			pluginFolder: "simple-backend",
			pluginSlug:   "grafana-simplebackend-datasource",
			hasBackend:   true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var publishCalls int
			runner, err := act.NewRunner(t)
			require.NoError(t, err)

			assertHeadersAndAuth := func(t *testing.T, r *http.Request) {
				expHeaders := http.Header{
					"Accept":        []string{"application/json"},
					"User-Agent":    []string{"github-actions-shared-workflows:/plugins/publish"},
					"Authorization": []string{"Bearer " + fakeGcomTokenDev},
				}
				if r.Header.Get("X-Api-Key") != "" {
					expHeaders.Set("Authorization", "Bearer "+fakeIapToken)
					expHeaders.Set("X-Api-Key", fakeGcomTokenDev)
				} else {
					expHeaders.Set("Authorization", "Bearer "+fakeGcomTokenDev)
				}
				require.Subset(t, r.Header, expHeaders)
			}

			runner.GCOM.HandleFunc("GET /api/plugins/{pluginID}", func(w http.ResponseWriter, r *http.Request) {
				assertHeadersAndAuth(t, r)

				w.WriteHeader(http.StatusOK)
				require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
					"id":     1,
					"status": "active",
					"slug":   tc.pluginSlug,
				}))
			})

			runner.GCOM.HandleFunc("POST /api/plugins", func(w http.ResponseWriter, r *http.Request) {
				publishCalls++
				assertHeadersAndAuth(t, r)

				var body map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body), "should be able to decode json body")
				require.Equal(t, []any{"universal"}, body["scopes"], "should have correct scopes")
				require.Equal(t, false, body["pending"], "should have correct pending status")
				require.Equal(t, "https://github.com/grafana/plugin-ci-workflows", body["url"], "should have correct url")
				require.Equal(t, gitSha, body["commit"], "should have correct commit sha")

				// Different URLs depending on backend (os/arch zips) or not (just "any" zip)
				expDownloadURLs := map[string]any{}
				if tc.hasBackend {
					for _, osArch := range osArchCombos {
						expDownloadURLs[osArch.os+"-"+osArch.arch] = map[string]any{"url": gcsPublishURLBackend(tc.pluginSlug, "1.0.0", osArch.os, osArch.arch)}
					}
				}
				expDownloadURLs["any"] = map[string]any{"url": gcsPublishURL(tc.pluginSlug, "1.0.0", "any")}
				require.Equal(t, expDownloadURLs, body["download"], "should have correct download URLs")

				w.WriteHeader(http.StatusOK)
				require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
					"plugin": map[string]any{"id": 1337},
				}))
			})

			// HACK: act doesn't support dynamic matrices (e.g.: `${{ fromJson(needs.setup.outputs.environments) }}`),
			// so hardcoding it for now.
			// TODO: add sanity check: make workflow fail if there's a dynamic matrix, otherwise it fails silently.
			environmentMatrixValue := []string{"dev"}
			var platformMatrixValue []string
			if tc.hasBackend {
				platformMatrixValue = []string{"linux", "darwin", "windows", "any"}
			} else {
				platformMatrixValue = []string{"any"}
			}

			wf, err := cd.NewWorkflow(
				cd.WithWorkflowInputs(cd.WorkflowInputs{
					CI: ci.WorkflowInputs{
						PluginDirectory:     workflow.Input(filepath.Join("tests", tc.pluginFolder)),
						DistArtifactsPrefix: workflow.Input(tc.pluginFolder + "-"),

						// Disable testing mode so the deployment is triggered
						Testing: workflow.Input(false),
						// No logic for mocking signatures yet, so allow unsigned for now
						AllowUnsigned: workflow.Input(true),

						// Disable some options for speeding up CI execution
						RunTruffleHog:      workflow.Input(false),
						RunPluginValidator: workflow.Input(false),
						RunPlaywright:      workflow.Input(false),
					},
					DisableDocsPublishing: workflow.Input(true),
					DisableGitHubRelease:  workflow.Input(true),

					// This doesn't work due to a bug in act, so hardcode it for now
					Environment: workflow.Input("dev"),

					Scopes: workflow.Input("universal"),
				}),
				cd.WithCIOptions(
					// Do not build the plugin, put pre-built zips in place to speed up CI execution
					ci.WithMockedPackagedDistArtifacts(
						t,
						"dist/"+tc.pluginFolder,
						"dist-artifacts-unsigned/"+tc.pluginFolder,
					),
					ci.WithMockedWorkflowContext(t, ci.Context{IsTrusted: true}),
				),

				// Mocks
				cd.WithMockedGCOM(t, runner.GCOM),
				cd.MutateAllWorkflows().With(
					workflow.WithMockedVault(t, mockVault),
					workflow.WithMockedGitHubAppToken(t),
					workflow.WithMockedGCS(t),
				),

				// CD mutations to make it work in tests
				cd.MutateCDWorkflow().With(
					// Mock GCS artifacts exist safety check
					workflow.WithNoOpStep(t, "upload-to-gcs-release", "gcloud-sdk"),
					workflow.WithNoOpStep(t, "publish-to-catalog", "check-artifact-zips"),

					// Mock IAP token for GCOM API calls
					workflow.WithReplacedStep(
						t, "publish-to-catalog", "gcloud",
						workflow.MockOutputsStep(map[string]string{
							"id_token": fakeIapToken,
						}),
					),

					// Hack for act dynamic matrices
					workflow.WithMatrix("publish-to-catalog", map[string][]string{
						"environment": environmentMatrixValue,
					}),
					workflow.WithMatrix("upload-to-gcs-release", map[string][]string{
						"platform": platformMatrixValue,
					}),
				),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewPushEventPayload("main"))
			require.NoError(t, err)

			require.True(t, r.Success, "workflow should succeed")

			// Check setup outputs which define the deployment target(s)
			// TODO: separate test case that tests for the setup outputs because the logic is quite complex.
			platformsValue, err := json.Marshal(platformMatrixValue)
			require.NoError(t, err)
			for k, v := range map[string]string{
				"platforms":             string(platformsValue),
				"plugin-version-suffix": "",
				"environments":          `["dev"]`,
				"publish-docs":          "false",
			} {
				o, ok := r.Outputs.Get("setup", "vars", k)
				require.True(t, ok)
				require.Equalf(t, v, o, "output %q should be %q", k, v)
			}

			// Ensure GCOM API calls were made and assertions were run
			require.Equal(t, 1, publishCalls, "GCOM API POST /plugins should be called exactly once")

			// Assert summary content
			require.Len(t, r.Summary, 1, "should have exactly one summary")
			require.Contains(t, r.Summary[0], "## 📦 Published to Catalog (dev)")
			require.Contains(t, r.Summary[0], "- **Plugin ID**: `"+tc.pluginSlug+"`")
			require.Contains(t, r.Summary[0], "- **Version**: `"+pluginVersion+"`")

			// Check GCS release upload
			expGCSFiles := []string{
				// CI artifacts
				filepath.Join("integration-artifacts", tc.pluginSlug, pluginVersion, "main", "latest", anyZipFileName(tc.pluginSlug, pluginVersion)),
				filepath.Join("integration-artifacts", tc.pluginSlug, pluginVersion, "main", gitSha, anyZipFileName(tc.pluginSlug, pluginVersion)),

				// Release artifacts
				filepath.Join("integration-artifacts", tc.pluginSlug, "release", pluginVersion, "any", anyZipFileName(tc.pluginSlug, pluginVersion)),
				filepath.Join("integration-artifacts", tc.pluginSlug, "release", "latest", "any", anyZipFileName(tc.pluginSlug, "latest")),
			}
			if tc.hasBackend {
				for _, osArch := range osArchCombos {
					expGCSFiles = append(
						expGCSFiles,
						// Os/arch CI artifacts
						filepath.Join("integration-artifacts", tc.pluginSlug, pluginVersion, "main", "latest", osArchZipFileName(tc.pluginSlug, pluginVersion, osArch)),
						filepath.Join("integration-artifacts", tc.pluginSlug, pluginVersion, "main", gitSha, osArchZipFileName(tc.pluginSlug, pluginVersion, osArch)),

						// Os/arch release artifacts
						filepath.Join("integration-artifacts", tc.pluginSlug, "release", pluginVersion, osArch.os, osArchZipFileName(tc.pluginSlug, pluginVersion, osArch)),
						filepath.Join("integration-artifacts", tc.pluginSlug, "release", "latest", osArch.os, osArchZipFileName(tc.pluginSlug, "latest", osArch)),
					)
				}
			}
			// Also expect the checksums
			for _, fn := range expGCSFiles {
				expGCSFiles = append(expGCSFiles, fn+".md5", fn+".sha1")
			}
			// This artifact for some reason doesn't have the corresponding checksum files,
			// so we add it manually after adding the checksums for all other files.
			expGCSFiles = append(expGCSFiles, filepath.Join("integration-artifacts", tc.pluginSlug, "release", "latest", anyZipFileName(tc.pluginSlug, "latest")))
			// Assert files exist in mocked GCS
			require.NoError(t, checkFilesExist(runner.GCS.Fs, expGCSFiles, checkFilesExistOptions{strict: true}), "GCS files should be present")
		})
	}
}

func TestCD_PublishIncludesProvenanceAttestation(t *testing.T) {
	t.Parallel()

	gitSha, err := getGitCommitSHA()
	require.NoError(t, err)

	const (
		pluginVersion    = "1.0.0"
		pluginFolder     = "simple-frontend"
		pluginSlug       = "grafana-simplefrontend-panel"
		fakeGcomTokenDev = "dummy-gcom-api-key-dev"
		fakeIapToken     = "dummy-iap-token"
	)

	zipBytes, err := os.ReadFile(workflow.LocalMockdataPath(
		"dist-artifacts-unsigned",
		pluginFolder,
		anyZipFileName(pluginSlug, pluginVersion),
	))
	require.NoError(t, err)
	zipSHA256 := sha256.Sum256(zipBytes)
	expProvenanceAttestation := fmt.Sprintf("github#grafana#sha256:%x", zipSHA256)

	mockVault := workflow.VaultSecrets{
		DefaultValue: newPointer(""),
		CommonSecrets: map[string]string{
			"plugins/gcom-publish-token:dev": fakeGcomTokenDev,
		},
	}

	var publishCalls int
	runner, err := act.NewRunner(t)
	require.NoError(t, err)

	runner.GCOM.HandleFunc("POST /api/plugins", func(w http.ResponseWriter, r *http.Request) {
		publishCalls++

		require.Subset(t, r.Header, http.Header{
			"Accept":        []string{"application/json"},
			"User-Agent":    []string{"github-actions-shared-workflows:/plugins/publish"},
			"Authorization": []string{"Bearer " + fakeGcomTokenDev},
		})

		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body), "should be able to decode json body")
		require.Equal(t, expProvenanceAttestation, body["provenanceAttestation"])

		w.WriteHeader(http.StatusOK)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"plugin": map[string]any{"id": 1337},
		}))
	})

	gcsZipURLs, err := json.Marshal([]string{gcsPublishURL(pluginSlug, pluginVersion, "any")})
	require.NoError(t, err)
	pluginOutput, err := json.Marshal(map[string]string{
		"id":      pluginSlug,
		"version": pluginVersion,
	})
	require.NoError(t, err)

	wf, err := cd.NewWorkflow(
		cd.WithWorkflowInputs(cd.WorkflowInputs{
			CI: ci.WorkflowInputs{
				PluginDirectory:     workflow.Input(filepath.Join("tests", pluginFolder)),
				DistArtifactsPrefix: workflow.Input(pluginFolder + "-"),
				Testing:             workflow.Input(false),
			},
			Attestation:           workflow.Input(true),
			DisableDocsPublishing: workflow.Input(true),
			DisableGitHubRelease:  workflow.Input(true),
			Environment:           workflow.Input("dev"),
			Scopes:                workflow.Input("universal"),
		}),
		cd.WithMockedGCOM(t, runner.GCOM),
		cd.MutateAllWorkflows().With(
			workflow.WithMockedVault(t, mockVault),
		),
		cd.MutateCDWorkflow().With(
			workflow.WithOnlyOneJob(t, "publish-to-catalog", false),
			workflow.WithNoOpJobWithOutputs(t, "setup", map[string]string{
				"commit-sha":            gitSha,
				"environments":          `["dev"]`,
				"publish-docs":          "false",
				"plugin-version-suffix": "",
				"platforms":             `["any"]`,
				"is-release-reference":  "true",
			}),
			workflow.WithNoOpJobWithOutputs(t, "ci", map[string]string{
				"plugin": string(pluginOutput),
			}),
			workflow.WithNoOpJobWithOutputs(t, "upload-to-gcs-release", map[string]string{
				"gcs-zip-urls": string(gcsZipURLs),
			}),
			workflow.WithReplacedStep(
				t,
				"build-attestation",
				"download-dist-artifacts",
				workflow.CopyMockFilesStep("dist-artifacts-unsigned/"+pluginFolder, "/tmp/dist-artifacts"),
			),
			workflow.WithReplacedStep(
				t,
				"build-attestation",
				"attestation",
				workflow.MockOutputsStep(map[string]string{
					"attestation-id":  "attestation-id",
					"attestation-url": "https://github.com/grafana/plugin-ci-workflows/attestations/attestation-id",
					"bundle-path":     "/tmp/attestation.jsonl",
				}),
			),
			workflow.WithEnvironment(t, "build-attestation", "provenance-attestation", map[string]string{
				"GITHUB_REPOSITORY_OWNER": "grafana",
			}),
			workflow.WithNoOpStep(t, "publish-to-catalog", "check-and-create-stub"),
			workflow.WithNoOpStep(t, "publish-to-catalog", "check-artifact-zips"),
			workflow.WithReplacedStep(
				t,
				"publish-to-catalog",
				"gcloud",
				workflow.MockOutputsStep(map[string]string{
					"id_token": fakeIapToken,
				}),
			),
			workflow.WithMatrix("publish-to-catalog", map[string][]string{
				"environment": {"dev"},
			}),
		),
	)
	require.NoError(t, err)

	r, err := runner.Run(wf, act.NewPushEventPayload("main"))
	require.NoError(t, err)
	require.True(t, r.Success, "workflow should succeed")
	require.Equal(t, 1, publishCalls, "GCOM API POST /plugins should be called exactly once")
}

func TestCD_GCSLatest(t *testing.T) {
	// Tests that "latest" GCS release artifacts are only uploaded for release references,
	// and can be forced via the upload-gcs-latest input.

	const (
		pluginVersion = "1.0.0"
		pluginFolder  = "simple-frontend"
		pluginSlug    = "grafana-simplefrontend-panel"
		fakeIapToken  = "dummy-iap-token"
	)

	mockVault := workflow.VaultSecrets{
		DefaultValue: newPointer(""),
		CommonSecrets: map[string]string{
			"plugins/gcom-publish-token:dev": "dummy-gcom-api-key-dev",
		},
	}

	for _, tc := range []struct {
		name              string
		branch            string
		uploadGCSLatest   bool
		expLatestUploaded bool
	}{
		{
			name:              "does not upload latest for non-release branch",
			branch:            "feature-branch",
			uploadGCSLatest:   false,
			expLatestUploaded: false,
		},
		{
			name:              "uploads latest for non-release branch when upload-gcs-latest is true",
			branch:            "feature-branch",
			uploadGCSLatest:   true,
			expLatestUploaded: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			runner, err := act.NewRunner(t)
			require.NoError(t, err)

			runner.GCOM.HandleFunc("GET /api/plugins/{pluginID}", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
					"id":     1,
					"status": "active",
					"slug":   pluginSlug,
				}))
			})
			runner.GCOM.HandleFunc("POST /api/plugins", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
					"plugin": map[string]any{"id": 1337},
				}))
			})

			wf, err := cd.NewWorkflow(
				cd.WithWorkflowInputs(cd.WorkflowInputs{
					CI: ci.WorkflowInputs{
						PluginDirectory:     workflow.Input(filepath.Join("tests", pluginFolder)),
						DistArtifactsPrefix: workflow.Input(pluginFolder + "-"),
						Testing:             workflow.Input(false),
						AllowUnsigned:       workflow.Input(true),
						RunTruffleHog:       workflow.Input(false),
						RunPluginValidator:  workflow.Input(false),
						RunPlaywright:       workflow.Input(false),
					},
					DisableDocsPublishing: workflow.Input(true),
					DisableGitHubRelease:  workflow.Input(true),
					Environment:           workflow.Input("dev"),
					Scopes:                workflow.Input("universal"),
					UploadGCSLatest:       workflow.Input(tc.uploadGCSLatest),
				}),
				cd.WithCIOptions(
					ci.WithMockedPackagedDistArtifacts(
						t,
						"dist/"+pluginFolder,
						"dist-artifacts-unsigned/"+pluginFolder,
					),
					ci.WithMockedWorkflowContext(t, ci.Context{IsTrusted: true}),
				),
				cd.WithMockedGCOM(t, runner.GCOM),
				cd.MutateAllWorkflows().With(
					workflow.WithMockedVault(t, mockVault),
					workflow.WithMockedGitHubAppToken(t),
					workflow.WithMockedGCS(t),
				),
				cd.MutateCDWorkflow().With(
					workflow.WithNoOpStep(t, "upload-to-gcs-release", "gcloud-sdk"),
					workflow.WithNoOpStep(t, "publish-to-catalog", "check-artifact-zips"),
					workflow.WithReplacedStep(
						t, "publish-to-catalog", "gcloud",
						workflow.MockOutputsStep(map[string]string{
							"id_token": fakeIapToken,
						}),
					),
					workflow.WithMatrix("publish-to-catalog", map[string][]string{
						"environment": {"dev"},
					}),
					workflow.WithMatrix("upload-to-gcs-release", map[string][]string{
						"platform": {"any"},
					}),
				),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewPushEventPayload(tc.branch))
			require.NoError(t, err)
			require.True(t, r.Success, "workflow should succeed")

			// The "latest" release files are the ones under release/latest/ in GCS.
			// The "any" platform zip has two entries: one in a platform subdirectory and one at the root.
			latestFiles := []string{
				filepath.Join("integration-artifacts", pluginSlug, "release", "latest", "any", anyZipFileName(pluginSlug, "latest")),
				filepath.Join("integration-artifacts", pluginSlug, "release", "latest", anyZipFileName(pluginSlug, "latest")),
			}

			if tc.expLatestUploaded {
				err = checkFilesExist(runner.GCS.Fs, latestFiles, checkFilesExistOptions{strict: false})
				require.NoError(t, err, "release/latest/ GCS files should be present")
			} else {
				err = checkFilesDontExist(runner.GCS.Fs, latestFiles)
				require.NoError(t, err, "release/latest/ GCS files should NOT be uploaded for non-release branches")
			}
		})
	}
}

func gcsPublishURL(pluginSlug string, version string, platform string) string {
	return "https://storage.googleapis.com/integration-artifacts/" + pluginSlug + "/release/" + version + "/" + platform + "/" + pluginSlug + "-" + version + ".zip"
}

func gcsPublishURLBackend(pluginSlug string, version string, os string, arch string) string {
	return "https://storage.googleapis.com/integration-artifacts/" + pluginSlug + "/release/" + version + "/" + os + "/" + pluginSlug + "-" + version + "." + os + "_" + arch + ".zip"
}
