package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/stretchr/testify/require"
)

func TestValidator(t *testing.T) {
	baseValidatorSummary := []act.SummaryEntry{
		{
			Level:   act.SummaryLevelWarning,
			Title:   "plugin-validator: Warning: unsigned plugin",
			Message: `This is a new (unpublished) plugin. This is expected during the initial review process. Please allow the review to continue, and a member of our team will inform you when your plugin can be signed.`,
		}, {
			Level:   act.SummaryLevelWarning,
			Title:   "plugin-validator: Warning: plugin.json: should include screenshots for the Plugin catalog",
			Message: `Screenshots are displayed in the Plugin catalog. Please add at least one screenshot to your plugin.json.`,
		}, {
			Level:   act.SummaryLevelNotice,
			Title:   "plugin-validator: Recommendation: You can include a sponsorship link if you want users to support your work",
			Message: `Consider to add a sponsorship link in your plugin.json file (Info.Links section: with Name: 'sponsor' or Name: 'sponsorship'), which will be shown on the plugin details page to allow users to support your work if they wish.`,
		}, {
			Level:   act.SummaryLevelWarning,
			Title:   "plugin-validator: Warning: plugin.json: description is empty",
			Message: `Consider providing a plugin description for better discoverability.`,
		}, {
			Level:   act.SummaryLevelWarning,
			Title:   "plugin-validator: Warning: License file contains generic text",
			Message: `Your current license file contains generic text from the license template. Please make sure to replace {name of copyright owner} and {yyyy} with the correct values in your LICENSE file.`,
		},
	}
	for _, tc := range []struct {
		name               string
		distFolder         string
		packagedDistFolder string

		expSuccess bool
		expSummary []act.SummaryEntry
	}{
		{
			name:               "simple-backend succeeds with warnings",
			distFolder:         "dist/simple-backend",
			packagedDistFolder: "dist-artifacts-unsigned/simple-backend",
			expSuccess:         true,
			expSummary:         baseValidatorSummary,
		},
		{
			name:               "simple-frontend-yarn succeeds with warnings",
			distFolder:         "dist/simple-frontend-yarn",
			packagedDistFolder: "dist-artifacts-unsigned/simple-frontend-yarn",
			expSuccess:         true,
			expSummary:         baseValidatorSummary,
		},
		// Special ZIP where the archive is malformed, used to test plugin-validator error handling
		{
			name:               "simple-frontend-validator-error fails",
			distFolder:         "dist/simple-frontend",
			packagedDistFolder: "dist-artifacts-other/simple-frontend-validator-error",
			expSuccess:         false,
			expSummary: []act.SummaryEntry{
				{
					Level:   act.SummaryLevelError,
					Title:   "plugin-validator: Error: Archive contains more than one directory",
					Message: `Archive should contain only one directory named after plugin id. Found 2 directories. Please see https://grafana.com/developers/plugin-tools/publish-a-plugin/package-a-plugin for more information on how to package a plugin.`,
				}, {
					Level:   act.SummaryLevelError,
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

			wf, err := workflow.NewSimpleCI(
				workflow.WithPluginDirectoryInput(filepath.Join("tests", tc.distFolder)),
				workflow.WithDistArtifactPrefixInput(tc.distFolder+"-"),

				// Disable some features to speed up the test
				workflow.WithPlaywrightInput(false),
				workflow.WithRunTruffleHogInput(false),

				// Enable the plugin validator (opt-in)
				workflow.WithRunPluginValidatorInput(true),

				// Mock dist so we don't spend time building the plugin
				workflow.WithMockedPackagedDistArtifacts(t, tc.distFolder, tc.packagedDistFolder),
			)
			require.NoError(t, err)

			r, err := runner.Run(wf, act.NewEmptyEventPayload())
			require.NoError(t, err)
			if tc.expSuccess {
				require.True(t, r.Success, "workflow should succeed")
			} else {
				require.False(t, r.Success, "workflow should fail")
			}

			// Check summary entries
			require.Subset(t, r.Summary, tc.expSummary)
			var validatorSummaryCount int
			for _, s := range r.Summary {
				if strings.HasPrefix(s.Title, "plugin-validator:") {
					validatorSummaryCount++
				}
			}
			require.Equal(t, validatorSummaryCount, len(tc.expSummary), "found unexpected plugin-validator gha summary entries")
		})
	}
}
