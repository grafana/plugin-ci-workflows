package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-logfmt/logfmt"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow/ci"
	"github.com/stretchr/testify/require"
)

func TestValidator(t *testing.T) {
	// Get DEFAULT_PLUGIN_VALIDATOR_VERSION from ci.yml
	ciWf, err := workflow.NewBaseWorkflowFromFile(filepath.Join(".github", "workflows", "ci.yml"))
	require.NoError(t, err)
	defaultPluginValidatorVersion := ciWf.Env["DEFAULT_PLUGIN_VALIDATOR_VERSION"]
	require.NotEmpty(t, defaultPluginValidatorVersion, "could not find DEFAULT_PLUGIN_VALIDATOR_VERSION env in ci.yml workflow")
	defaultPluginValidatorVersion = "v" + defaultPluginValidatorVersion

	baseValidatorAnnotations := []act.Annotation{
		{
			Level:   act.AnnotationLevelWarning,
			Title:   "plugin-validator: Warning: unsigned plugin",
			Message: `This is a new (unpublished) plugin. This is expected during the initial review process. Please allow the review to continue, and a member of our team will inform you when your plugin can be signed.`,
		}, {
			Level:   act.AnnotationLevelWarning,
			Title:   "plugin-validator: Warning: plugin.json: should include screenshots for the Plugin catalog",
			Message: `Screenshots are displayed in the Plugin catalog. Please add at least one screenshot to your plugin.json.`,
		}, {
			Level:   act.AnnotationLevelNotice,
			Title:   "plugin-validator: Recommendation: You can include a sponsorship link if you want users to support your work",
			Message: `Consider to add a sponsorship link in your plugin.json file (Info.Links section: with Name: 'sponsor' or Name: 'sponsorship'), which will be shown on the plugin details page to allow users to support your work if they wish.`,
		}, {
			Level:   act.AnnotationLevelWarning,
			Title:   "plugin-validator: Warning: plugin.json: description is empty",
			Message: `Consider providing a plugin description for better discoverability.`,
		}, {
			Level:   act.AnnotationLevelWarning,
			Title:   "plugin-validator: Warning: License file contains generic text",
			Message: `Your current license file contains generic text from the license template. Please make sure to replace {name of copyright owner} and {yyyy} with the correct values in your LICENSE file.`,
		},
	}
	// Custom validator config used by most test cases, which disables some analyzers
	// that are too slow or flaky for testing.
	validatorConfig := `
global:
  enabled: true
analyzers:
  # Disabled because it takes too much resources and time.
  osv-scanner:
    enabled: false

  # Warns if backend SDK is too old. Disabled so we don't have to bump it to fix tests.
  sdkusage:
    enabled: false
`

	for _, tc := range []struct {
		name               string
		sourceFolder       string
		packagedDistFolder string

		validatorConfig *string

		expSuccess                          *bool
		expAnnotations                      []act.Annotation
		unexpectedAnnotationTitleSubstrings []string
		expPluginValidatorVersion           string
		expPluginValidatorConfigSource      string
	}{
		{
			name:                           "simple-backend succeeds with warnings",
			sourceFolder:                   "simple-backend",
			packagedDistFolder:             "dist-artifacts-unsigned/simple-backend",
			validatorConfig:                &validatorConfig,
			expSuccess:                     newPointer(true),
			expAnnotations:                 baseValidatorAnnotations,
			expPluginValidatorVersion:      defaultPluginValidatorVersion,
			expPluginValidatorConfigSource: "custom",
		},
		{
			name:                           "simple-frontend-yarn succeeds with warnings",
			sourceFolder:                   "simple-frontend-yarn",
			packagedDistFolder:             "dist-artifacts-unsigned/simple-frontend-yarn",
			validatorConfig:                &validatorConfig,
			expSuccess:                     newPointer(true),
			expAnnotations:                 baseValidatorAnnotations,
			expPluginValidatorVersion:      defaultPluginValidatorVersion,
			expPluginValidatorConfigSource: "custom",
		},
		// Special ZIP where the archive is malformed, used to test plugin-validator error handling
		{
			name:               "simple-frontend-validator-error fails",
			sourceFolder:       "simple-frontend",
			packagedDistFolder: "dist-artifacts-other/simple-frontend-validator-error",
			validatorConfig:    &validatorConfig,
			expSuccess:         newPointer(false),
			expAnnotations: []act.Annotation{
				{
					Level:   act.AnnotationLevelError,
					Title:   "plugin-validator: Error: Archive contains more than one directory",
					Message: `Archive should contain only one directory named after plugin id. Found 2 directories. Please see https://grafana.com/developers/plugin-tools/publish-a-plugin/package-a-plugin for more information on how to package a plugin.`,
				}, {
					Level:   act.AnnotationLevelError,
					Title:   "plugin-validator: Error: Plugin archive is improperly structured",
					Message: `It is possible your plugin archive structure is incorrect. Please see https://grafana.com/developers/plugin-tools/publish-a-plugin/package-a-plugin for more information on how to package a plugin.`,
				}, {
					Level:   act.AnnotationLevelWarning,
					Title:   "plugin-validator: Warning: Code diff skipped due to errors in archive",
					Message: `Fix the errors reported by archive before code diff can run.`,
				}, {
					Level:   act.AnnotationLevelWarning,
					Title:   "plugin-validator: Warning: LLM review skipped due to errors in archive",
					Message: `Fix the errors reported by archive before LLM review can run.`,
				},
			},
			expPluginValidatorVersion:      defaultPluginValidatorVersion,
			expPluginValidatorConfigSource: "custom",
		},
		// Test that the default configuration (from the workflow env DEFAULT_PLUGIN_VALIDATOR_CONFIG)
		// is used when no custom config is provided.
		// We don't assert on success/failure or exact annotations because the default config
		// leaves osv-scanner enabled, which may produce different results depending on the
		// current vulnerability database. Instead, we verify the config was applied by checking
		// that annotations from disabled analyzers are NOT present.
		{
			name:               "simple-frontend-yarn with default config",
			sourceFolder:       "simple-frontend-yarn",
			packagedDistFolder: "dist-artifacts-unsigned/simple-frontend-yarn",
			validatorConfig:    nil,
			// Annotations from disabled analyzers (discoverability, license, sponsorshiplink)
			// should not appear when the default config is used.
			unexpectedAnnotationTitleSubstrings: []string{
				"description is empty", // discoverability analyzer
				"License file",         // license analyzer
				"sponsorship link",     // sponsorshiplink analyzer
			},
			expPluginValidatorVersion:      defaultPluginValidatorVersion,
			expPluginValidatorConfigSource: "default",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runner, err := act.NewRunner(t, act.WithLinuxAMD64ContainerArchitecture())
			require.NoError(t, err)

			inputs := ci.WorkflowInputs{
				PluginDirectory:     workflow.Input("tests/" + tc.sourceFolder),
				DistArtifactsPrefix: workflow.Input(tc.sourceFolder + "-"),

				// Disable some features to speed up the test
				RunPlaywright: workflow.Input(false),
				RunTruffleHog: workflow.Input(false),

				// Enable the plugin validator (opt-in)
				RunPluginValidator: workflow.Input(true),
			}
			if tc.validatorConfig != nil {
				inputs.PluginValidatorConfig = tc.validatorConfig
			}

			wf, err := ci.NewWorkflow(
				ci.WithWorkflowInputs(inputs),
				// Mock dist so we don't spend time building the plugin
				ci.WithMockedPackagedDistArtifacts(t, "dist/"+tc.sourceFolder, tc.packagedDistFolder),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewPushEventPayload("main"))
			require.NoError(t, err)
			if tc.expSuccess != nil {
				if *tc.expSuccess {
					require.True(t, r.Success, "workflow should succeed")
				} else {
					require.False(t, r.Success, "workflow should fail")
				}
			}

			// Check debug annotation with the validator version, if necessary
			if tc.expPluginValidatorVersion != "" {
				logFmtMsg, err := logfmt.MarshalKeyvals("msg", "Running plugin-validator", "version", tc.expPluginValidatorVersion)
				require.NoError(t, err)
				require.Contains(t, r.Annotations, act.Annotation{
					Level:   act.AnnotationLevelDebug,
					Message: string(logFmtMsg),
				})
			}

			// Check debug annotation with the validator config source
			if tc.expPluginValidatorConfigSource != "" {
				logFmtMsg, err := logfmt.MarshalKeyvals("msg", "plugin-validator configuration", "source", tc.expPluginValidatorConfigSource)
				require.NoError(t, err)
				require.Contains(t, r.Annotations, act.Annotation{
					Level:   act.AnnotationLevelDebug,
					Message: string(logFmtMsg),
				})
			}

			// Check expected annotation entries and exact count
			if tc.expAnnotations != nil {
				require.Subset(t, r.Annotations, tc.expAnnotations)
				var validatorAnnotationCount int
				for _, s := range r.Annotations {
					if strings.HasPrefix(s.Title, "plugin-validator:") {
						validatorAnnotationCount++
					}
				}
				require.Equal(t, validatorAnnotationCount, len(tc.expAnnotations), "found unexpected plugin-validator gha annotation entries")
			}

			// Check that annotations from disabled analyzers are NOT present
			for _, unexpected := range tc.unexpectedAnnotationTitleSubstrings {
				for _, a := range r.Annotations {
					require.NotContains(t, a.Title, unexpected, "annotation from a disabled analyzer should not be present: %s", a.Title)
				}
			}
		})
	}
}
