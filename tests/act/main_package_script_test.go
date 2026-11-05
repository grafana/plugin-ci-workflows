package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const packageScriptPath = "actions/internal/plugins/package/package.sh"

// TestPackageScriptPlatformExecutables runs package.sh directly, without act,
// on an app with a nested datasource. The script builds the os/arch packages in
// parallel, so a package that fails must still fail the whole script.
func TestPackageScriptPlatformExecutables(t *testing.T) {
	t.Parallel()

	for _, tool := range []string{"bash", "jq", "zip", "sha1sum", "md5sum"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("package.sh needs %s", tool)
		}
	}
	script, err := filepath.Abs(packageScriptPath)
	require.NoError(t, err)

	for _, tc := range []struct {
		name         string
		datasourceOS []string
		expErr       bool
		expZips      []string
	}{
		{
			name:         "every executable present",
			datasourceOS: []string{"linux_amd64", "linux_arm64", "darwin_arm64"},
			expZips:      []string{"linux_amd64", "linux_arm64", "darwin_arm64"},
		},
		{
			name:         "nested datasource executable missing",
			datasourceOS: []string{"linux_amd64", "darwin_arm64"},
			expErr:       true,
			// The other packages still finish.
			expZips: []string{"linux_amd64", "darwin_arm64"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			dist := filepath.Join(dir, "dist")
			out := filepath.Join(dir, "out")
			writeFile(t, filepath.Join(dist, "plugin.json"), `{"id":"test-app","type":"app","executable":"gpx_app","info":{"version":"1.0.0"}}`)
			writeFile(t, filepath.Join(dist, "datasource", "plugin.json"), `{"id":"test-datasource","type":"datasource","executable":"gpx_ds"}`)
			for _, osArch := range []string{"linux_amd64", "linux_arm64", "darwin_arm64"} {
				writeFile(t, filepath.Join(dist, "gpx_app_"+osArch), "app")
			}
			for _, osArch := range tc.datasourceOS {
				writeFile(t, filepath.Join(dist, "datasource", "gpx_ds_"+osArch), "datasource")
			}

			cmd := exec.Command("bash", script, dist, out)
			// No token, so the script does not sign.
			cmd.Env = append(os.Environ(), "GRAFANA_ACCESS_POLICY_TOKEN=")
			output, err := cmd.CombinedOutput()
			if tc.expErr {
				require.Error(t, err, string(output))
				require.Contains(t, string(output), "Missing gpx_ds_linux_arm64 in 'datasource', aborting.")
			} else {
				require.NoError(t, err, string(output))
			}
			for _, osArch := range tc.expZips {
				require.FileExists(t, filepath.Join(out, "test-app-1.0.0."+osArch+".zip"))
			}
		})
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o755))
}
