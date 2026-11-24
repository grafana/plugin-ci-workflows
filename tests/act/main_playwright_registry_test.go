package main

import (
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/stretchr/testify/require"
)

const enterpriseNightlyRegistry = "us-docker.pkg.dev/grafanalabs-dev/docker-grafana-enterprise-dev"

var playwrightRegistryWorkflows = []string{"playwright.yml", "playwright-docker.yml"}

func playwrightStep(t *testing.T, file, name string) *workflow.Step {
	t.Helper()

	wf, err := workflow.NewBaseWorkflowFromFile(filepath.Join(".github", "workflows", file))
	require.NoError(t, err)
	for _, step := range wf.Jobs["playwright-tests"].Steps {
		if step.Name == name {
			return &step
		}
	}
	require.FailNowf(t, "step not found", "%s has no step %q", file, name)
	return nil
}

func runBash(t *testing.T, script, dir string, env map[string]string) (string, error) {
	t.Helper()

	cmd := exec.Command("bash", "-e", "-c", script)
	cmd.Dir = dir
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME")}
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestPlaywrightEnterpriseNightlyResolveStep(t *testing.T) {
	t.Parallel()

	nightly := map[string]string{"IMAGE_NAME": "grafana-enterprise", "IMAGE_VERSION": "nightly"}
	private := map[string]string{"REPOSITORY_IS_PRIVATE": "true", "REPOSITORY_VISIBILITY": "private"}

	merge := func(sets ...map[string]string) map[string]string {
		out := map[string]string{}
		for _, m := range sets {
			maps.Copy(out, m)
		}
		return out
	}
	registry := map[string]string{"ENTERPRISE_NIGHTLY_REGISTRY": enterpriseNightlyRegistry}

	for _, tc := range []struct {
		name         string
		env          map[string]string
		expectErr    string
		expectOutput string
		expectEnv    string
	}{
		{
			name: "registry not set does nothing",
			env:  merge(map[string]string{"REPOSITORY_IS_PRIVATE": "false", "REPOSITORY_VISIBILITY": "public"}, nightly),
		},
		{
			name:      "public repository is rejected",
			env:       merge(registry, nightly, map[string]string{"REPOSITORY_IS_PRIVATE": "false", "REPOSITORY_VISIBILITY": "public"}),
			expectErr: "private or internal repositories",
		},
		{
			name:      "missing repository metadata is rejected",
			env:       merge(registry, nightly),
			expectErr: "private or internal repositories",
		},
		{
			name:         "internal repository reported through visibility is accepted",
			env:          merge(registry, nightly, map[string]string{"REPOSITORY_IS_PRIVATE": "", "REPOSITORY_VISIBILITY": "internal"}),
			expectOutput: "gar-host=us-docker.pkg.dev",
			expectEnv:    "GRAFANA_REPOSITORY=" + enterpriseNightlyRegistry + "/grafana-enterprise",
		},
		{
			name:      "invalid registry is rejected",
			env:       merge(private, nightly, map[string]string{"ENTERPRISE_NIGHTLY_REGISTRY": "bad value;x"}),
			expectErr: "must be a registry path",
		},
		{
			name: "stable enterprise entry keeps Docker Hub",
			env:  merge(registry, private, map[string]string{"IMAGE_NAME": "grafana-enterprise", "IMAGE_VERSION": "12.3.0"}),
		},
		{
			name: "oss nightly entry keeps Docker Hub",
			env:  merge(registry, private, map[string]string{"IMAGE_NAME": "grafana", "IMAGE_VERSION": "nightly"}),
		},
		{
			name:         "enterprise nightly entry uses the registry",
			env:          merge(registry, private, nightly),
			expectOutput: "gar-host=us-docker.pkg.dev",
			expectEnv:    "GRAFANA_REPOSITORY=" + enterpriseNightlyRegistry + "/grafana-enterprise",
		},
	} {
		for _, file := range playwrightRegistryWorkflows {
			t.Run(file+"/"+tc.name, func(t *testing.T) {
				t.Parallel()

				step := playwrightStep(t, file, "Resolve Grafana repository")
				dir := t.TempDir()
				outputFile := filepath.Join(dir, "output")
				envFile := filepath.Join(dir, "env")
				require.NoError(t, os.WriteFile(outputFile, nil, 0o600))
				require.NoError(t, os.WriteFile(envFile, nil, 0o600))

				out, err := runBash(t, step.Run, dir, merge(tc.env, map[string]string{
					"GITHUB_OUTPUT": outputFile,
					"GITHUB_ENV":    envFile,
				}))

				if tc.expectErr != "" {
					require.Error(t, err)
					require.Contains(t, out, tc.expectErr)
				} else {
					require.NoError(t, err, out)
				}

				gotOutput, err := os.ReadFile(outputFile)
				require.NoError(t, err)
				gotEnv, err := os.ReadFile(envFile)
				require.NoError(t, err)
				require.Equal(t, tc.expectOutput, strings.TrimSpace(string(gotOutput)))
				require.Equal(t, tc.expectEnv, strings.TrimSpace(string(gotEnv)))
			})
		}
	}
}

func TestPlaywrightEnterpriseNightlyComposeProbe(t *testing.T) {
	t.Parallel()

	if err := exec.Command("docker", "compose", "version").Run(); err != nil {
		t.Skip("docker compose is not available")
	}

	const composeTemplate = `services:
  grafana:
    build:
      context: .
      args:
        grafana_image: ${GRAFANA_IMAGE:-grafana}
%s        grafana_version: ${GRAFANA_VERSION:-latest}
`

	for _, tc := range []struct {
		name      string
		args      string
		expectErr bool
	}{
		{name: "compose file that reads GRAFANA_REPOSITORY", args: "        grafana_repository: ${GRAFANA_REPOSITORY:-}\n"},
		{name: "compose file that ignores GRAFANA_REPOSITORY", args: "", expectErr: true},
	} {
		for _, file := range playwrightRegistryWorkflows {
			t.Run(file+"/"+tc.name, func(t *testing.T) {
				t.Parallel()

				step := playwrightStep(t, file, "Check docker-compose setup supports GRAFANA_REPOSITORY")
				dir := t.TempDir()
				compose := filepath.Join(dir, "docker-compose.yaml")
				content := strings.Replace(composeTemplate, "%s", tc.args, 1)
				require.NoError(t, os.WriteFile(compose, []byte(content), 0o600))

				out, err := runBash(t, step.Run, dir, map[string]string{"DOCKER_COMPOSE_FILE": compose})
				if tc.expectErr {
					require.Error(t, err)
					require.Contains(t, out, "does not use GRAFANA_REPOSITORY")
					return
				}
				require.NoError(t, err, out)
			})
		}
	}
}
