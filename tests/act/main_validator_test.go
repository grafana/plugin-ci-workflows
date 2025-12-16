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
	runner, err := act.NewRunner(t, act.WithLinuxAMD64ContainerArchitecture())
	require.NoError(t, err)

	wf, err := workflow.NewSimpleCI(
		workflow.WithPluginDirectoryInput(filepath.Join("tests", "simple-frontend")),
		workflow.WithDistArtifactPrefixInput("simple-frontend-"),

		// Disable some features to speed up the test
		workflow.WithPlaywrightInput(false),
		workflow.WithRunTruffleHogInput(false),

		// Enable the plugin validator (opt-in)
		workflow.WithRunPluginValidatorInput(true),

		// Mock dist so we don't spend time building the plugin
		workflow.WithMockedPackagedDistArtifacts(t, "simple-frontend", false),
	)
	require.NoError(t, err)

	r, err := runner.Run(wf, act.NewEmptyEventPayload())
	require.NoError(t, err)
	require.True(t, r.Success, "workflow should succeed")

	// Check summary entries
	expSummary := []act.SummaryEntry{
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
	require.Subset(t, r.Summary, expSummary)
	var validatorSummaryCount int
	for _, s := range r.Summary {
		if strings.HasPrefix(s.Title, "plugin-validator:") {
			validatorSummaryCount++
		}
	}
	require.Equal(t, validatorSummaryCount, len(expSummary), "found unexpected plugin-validator gha summary entries")
}
