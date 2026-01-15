package main

import (
	"strings"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow/ci"
	"github.com/stretchr/testify/require"
)

func TestValidator(t *testing.T) {
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
	for _, tc := range []struct {
		name               string
		sourceFolder       string
		packagedDistFolder string

		expSuccess     bool
		expAnnotations []act.Annotation
	}{
		{
			name:               "simple-backend succeeds with warnings",
			sourceFolder:       "simple-backend",
			packagedDistFolder: "dist-artifacts-unsigned/simple-backend",
			expSuccess:         true,
			expAnnotations:     baseValidatorAnnotations,
		},
		{
			name:               "simple-frontend-yarn succeeds with warnings",
			sourceFolder:       "simple-frontend-yarn",
			packagedDistFolder: "dist-artifacts-unsigned/simple-frontend-yarn",
			expSuccess:         true,
			expAnnotations:     baseValidatorAnnotations,
		},
		// Special ZIP where the archive is malformed, used to test plugin-validator error handling
		{
			name:               "simple-frontend-validator-error fails",
			sourceFolder:       "simple-frontend",
			packagedDistFolder: "dist-artifacts-other/simple-frontend-validator-error",
			expSuccess:         false,
			expAnnotations: []act.Annotation{
				{
					Level:   act.AnnotationLevelError,
					Title:   "plugin-validator: Error: Archive contains more than one directory",
					Message: `Archive should contain only one directory named after plugin id. Found 2 directories. Please see https://grafana.com/developers/plugin-tools/publish-a-plugin/package-a-plugin for more information on how to package a plugin.`,
				}, {
					Level:   act.AnnotationLevelError,
					Title:   "plugin-validator: Error: Plugin archive is improperly structured",
					Message: `It is possible your plugin archive structure is incorrect. Please see https://grafana.com/developers/plugin-tools/publish-a-plugin/package-a-plugin for more information on how to package a plugin.`,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runner, err := act.NewRunner(t, act.WithLinuxAMD64ContainerArchitecture())
			require.NoError(t, err)

			validatorConfig := `
global:
  enabled: true
analyzers:
  osv-scanner:
    enabled: false
`
			wf, err := ci.NewWorkflow(
				ci.WithWorkflowInputs(ci.WorkflowInputs{
					PluginDirectory:     workflow.Input("tests/" + tc.sourceFolder),
					DistArtifactsPrefix: workflow.Input(tc.sourceFolder + "-"),

					// Disable some features to speed up the test
					RunPlaywright: workflow.Input(false),
					RunTruffleHog: workflow.Input(false),

					// Enable the plugin validator (opt-in)
					RunPluginValidator:    workflow.Input(true),
					PluginValidatorConfig: workflow.Input(validatorConfig),
				}),
				// Mock dist so we don't spend time building the plugin
				ci.WithMockedPackagedDistArtifacts(t, "dist/"+tc.sourceFolder, tc.packagedDistFolder),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewPushEventPayload("main"))
			require.NoError(t, err)
			if tc.expSuccess {
				require.True(t, r.Success, "workflow should succeed")
			} else {
				require.False(t, r.Success, "workflow should fail")
			}

			// Check annotation entries
			require.Subset(t, r.Annotations, tc.expAnnotations)
			var validatorAnnotationCount int
			for _, s := range r.Annotations {
				if strings.HasPrefix(s.Title, "plugin-validator:") {
					validatorAnnotationCount++
				}
			}
			require.Equal(t, validatorAnnotationCount, len(tc.expAnnotations), "found unexpected plugin-validator gha annotation entries")
		})
	}
}
