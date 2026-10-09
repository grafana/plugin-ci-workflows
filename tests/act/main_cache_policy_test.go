package main

import (
	"os"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
)

// These are static configuration tests: they assert on the workflow/action YAML only.
// They do not evaluate expressions or exercise the GitHub cache service.

const (
	setupActionPath      = "actions/internal/plugins/setup/action.yml"
	trufflehogActionPath = "actions/internal/plugins/trufflehog/action.yml"

	nodeCacheExpr = "${{ (inputs.node-setup-caching == 'true' && inputs.act-cache-warmup != 'true' && steps.package-manager.outputs.name != 'pnpm') && steps.package-manager.outputs.name || '' }}"
	nodeDepsExpr  = "${{ inputs.node-setup-caching == 'true' && inputs.act-cache-warmup != 'true' && steps.package-manager.outputs.lockFilePath || '' }}"
)

type compositeAction struct {
	Inputs map[string]workflow.WorkflowCallInput `yaml:"inputs"`
	Runs   workflow.Job                          `yaml:"runs"`
}

func loadAction(t *testing.T, path string) compositeAction {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	var a compositeAction
	require.NoError(t, yaml.Unmarshal(b, &a))
	return a
}

func loadWorkflow(t *testing.T, name string) workflow.BaseWorkflow {
	t.Helper()
	wf, err := workflow.NewBaseWorkflowFromFile(".github/workflows/" + name)
	require.NoError(t, err)
	return wf
}

func TestCachingFlagDeclarations(t *testing.T) {
	t.Parallel()

	ci := loadWorkflow(t, "ci.yml")
	cd := loadWorkflow(t, "cd.yml")
	pw := loadWorkflow(t, "playwright.yml")
	setup := loadAction(t, setupActionPath)
	trufflehog := loadAction(t, trufflehogActionPath)

	flags := []string{"go-setup-caching", "node-setup-caching", "go-tooling-caching", "trufflehog-caching", "playwright-caching"}
	for _, flag := range flags {
		t.Run(flag, func(t *testing.T) {
			t.Parallel()
			in := ci.On.WorkflowCall.Inputs[flag]
			require.Equal(t, workflow.WorkflowCallInputTypeBoolean, in.Type)
			require.Equal(t, true, in.Default)
			require.NotContains(t, cd.On.WorkflowCall.Inputs, flag, "CD must not expose cache flags")
			require.Equal(t, false, cd.Jobs["ci"].With[flag], "CD must force %s off", flag)
		})
	}

	t.Run("playwright defaults", func(t *testing.T) {
		t.Parallel()
		for _, flag := range []string{"node-setup-caching", "playwright-caching"} {
			require.Equal(t, true, pw.On.WorkflowCall.Inputs[flag].Default, flag)
		}
	})
	t.Run("composite action defaults", func(t *testing.T) {
		t.Parallel()
		for _, flag := range []string{"go-setup-caching", "node-setup-caching", "go-tooling-caching"} {
			require.Equal(t, "true", setup.Inputs[flag].Default, flag)
		}
		require.Equal(t, "true", trufflehog.Inputs["cache"].Default)
	})
}

func TestCachingFlagForwarding(t *testing.T) {
	t.Parallel()

	ci := loadWorkflow(t, "ci.yml")
	pw := loadWorkflow(t, "playwright.yml")
	ciSetup := ci.Jobs["test-and-build"].GetStep("setup")
	require.NotNil(t, ciSetup)
	pwSetup := pw.Jobs["playwright-tests"].GetStep("setup")
	require.NotNil(t, pwSetup)
	trufflehog := ci.Jobs["test-and-build"].GetStep("trufflehog")
	require.NotNil(t, trufflehog)

	for _, tc := range []struct {
		name string
		with map[string]any
		key  string
		want any
	}{
		{"ci setup go", ciSetup.With, "go-setup-caching", "${{ inputs.go-setup-caching }}"},
		{"ci setup node", ciSetup.With, "node-setup-caching", "${{ inputs.node-setup-caching }}"},
		{"ci setup go tooling", ciSetup.With, "go-tooling-caching", "${{ inputs.go-tooling-caching }}"},
		{"ci trufflehog", trufflehog.With, "cache", "${{ inputs.trufflehog-caching }}"},
		{"ci playwright node", ci.Jobs["playwright"].With, "node-setup-caching", "${{ inputs.node-setup-caching }}"},
		{"ci playwright browsers", ci.Jobs["playwright"].With, "playwright-caching", "${{ inputs.playwright-caching }}"},
		{"playwright setup node", pwSetup.With, "node-setup-caching", "${{ inputs.node-setup-caching }}"},
		// Frontend-only, so Playwright never enables Go caches.
		{"playwright setup frontend-only", pwSetup.With, "frontend-only", "true"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, tc.with[tc.key])
		})
	}
}

func TestSetupImplicitCachingDisabled(t *testing.T) {
	t.Parallel()

	setup := loadAction(t, setupActionPath).Runs
	node, goStep, pnpm := setup.GetStep("node"), setup.GetStep("go"), setup.GetStep("pnpm")
	require.NotNil(t, node)
	require.NotNil(t, goStep)
	require.NotNil(t, pnpm)

	require.Equal(t, false, node.With["package-manager-cache"])
	require.Equal(t, nodeCacheExpr, node.With["cache"])
	require.Equal(t, nodeDepsExpr, node.With["cache-dependency-path"])
	require.Equal(t, "${{ inputs.go-setup-caching == 'true' }}", goStep.With["cache"])
	require.Equal(t, false, pnpm.With["cache"])
}

func TestCacheStepGating(t *testing.T) {
	t.Parallel()

	setup := loadAction(t, setupActionPath).Runs
	trufflehog := loadAction(t, trufflehogActionPath).Runs
	pw := loadWorkflow(t, "playwright.yml").Jobs["playwright-tests"]
	const installIf = "${{ inputs.frontend-only != 'true' && steps.cache.outputs.cache-hit != 'true' }}"

	for _, tc := range []struct {
		name   string
		job    *workflow.Job
		id     string
		wantIf string
	}{
		{"pnpm store", &setup, "cache-pnpm", "${{ inputs.node-setup-caching == 'true' && steps.package-manager.outputs.name == 'pnpm' }}"},
		{"go tooling", &setup, "cache", "${{ inputs.frontend-only != 'true' && inputs.go-tooling-caching == 'true' }}"},
		{"trufflehog", &trufflehog, "cache", "${{ inputs.cache == 'true' }}"},
		{"playwright", pw, "cache", "${{ inputs.playwright-caching }}"},
		// Disabling the archive cache must not suppress fresh installs.
		{"mage install", &setup, "mage", installIf},
		{"golangci-lint install", &setup, "golangci-lint", installIf},
		{"trufflehog install", &trufflehog, "install", "${{ steps.cache.outputs.cache-hit != 'true' }}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			step := tc.job.GetStep(tc.id)
			require.NotNil(t, step)
			require.Equal(t, tc.wantIf, step.If)
		})
	}

	t.Run("no unaccounted actions/cache steps", func(t *testing.T) {
		t.Parallel()
		var got []string
		for group, job := range map[string]*workflow.Job{"setup": &setup, "trufflehog": &trufflehog, "playwright": pw} {
			for _, s := range job.Steps {
				if strings.HasPrefix(s.Uses, "actions/cache@") || strings.HasPrefix(s.Uses, "actions/cache/") {
					got = append(got, group+"/"+s.ID)
				}
			}
		}
		require.ElementsMatch(t, []string{"setup/cache-pnpm", "setup/cache", "trufflehog/cache", "playwright/cache"}, got)
	})
}
