package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/stretchr/testify/require"
)

func TestCheckArtifacts(t *testing.T) {
	t.Parallel()

	for _, tool := range []string{"bash", "curl", "unzip", "jq"} {
		_, err := exec.LookPath(tool)
		require.NoError(t, err, "check-artifacts.sh requires %s", tool)
	}

	const pluginID = "myorg-test-app"
	zips := map[string]map[string]string{
		"signed.zip": {
			pluginID + "/plugin.json":  `{"id":"` + pluginID + `"}`,
			pluginID + "/MANIFEST.txt": "-----BEGIN PGP SIGNED MESSAGE-----",
		},
		"unsigned.zip": {
			pluginID + "/plugin.json": `{"id":"` + pluginID + `"}`,
		},
		"empty-manifest.zip": {
			pluginID + "/plugin.json":  `{"id":"` + pluginID + `"}`,
			pluginID + "/MANIFEST.txt": "",
		},
		"no-plugin-json.zip": {
			pluginID + "/MANIFEST.txt": "-----BEGIN PGP SIGNED MESSAGE-----",
		},
		"wrong-plugin-id.zip": {
			"other-plugin-app/plugin.json":  `{"id":"other-plugin-app"}`,
			"other-plugin-app/MANIFEST.txt": "-----BEGIN PGP SIGNED MESSAGE-----",
		},
	}

	mux := http.NewServeMux()
	for name, files := range zips {
		content := newZip(t, files)
		mux.HandleFunc("/"+name, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write(content)
		})
	}
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	jsonURLs := func(names ...string) string {
		quoted := make([]string, 0, len(names))
		for _, name := range names {
			quoted = append(quoted, `"`+srv.URL+"/"+name+`"`)
		}
		return "[" + strings.Join(quoted, ",") + "]"
	}

	for _, tc := range []struct {
		name          string
		zips          string
		expectSuccess bool
		expOutput     string
	}{
		{name: "single URL string", zips: srv.URL + "/signed.zip", expectSuccess: true},
		{name: "all valid", zips: jsonURLs("signed.zip", "signed.zip"), expectSuccess: true},
		{name: "missing MANIFEST.txt", zips: jsonURLs("unsigned.zip"), expOutput: "MANIFEST.txt missing or empty"},
		{name: "empty MANIFEST.txt", zips: jsonURLs("empty-manifest.zip"), expOutput: "MANIFEST.txt missing or empty"},
		{name: "missing plugin.json", zips: jsonURLs("no-plugin-json.zip"), expOutput: "plugin.json missing or empty"},
		{name: "wrong plugin ID folder", zips: jsonURLs("wrong-plugin-id.zip"), expOutput: "Folder '" + pluginID + "' not found"},
		{name: "one invalid among valid", zips: jsonURLs("signed.zip", "unsigned.zip", "signed.zip"), expOutput: "Completed with 1 failures"},
		{name: "download not a ZIP", zips: jsonURLs("does-not-exist.zip"), expOutput: "Failed to extract"},
		{name: "empty array", zips: "[]", expOutput: "No ZIP URLs to validate"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			script := filepath.Join("actions", "internal", "plugins", "publish", "check-artifacts", "check-artifacts.sh")
			out, err := exec.Command("bash", script, tc.zips, pluginID).CombinedOutput()
			if tc.expectSuccess {
				require.NoError(t, err, "output:\n%s", out)
				require.Contains(t, string(out), "All ZIP files validated successfully")
				return
			}
			var exitErr *exec.ExitError
			require.True(t, errors.As(err, &exitErr), "expected non-zero exit, got err=%v, output:\n%s", err, out)
			require.Contains(t, string(out), tc.expOutput)
			require.NotContains(t, string(out), "All ZIP files validated successfully")
		})
	}
}

func TestCheckArtifacts_GCSAccessToken(t *testing.T) {
	t.Parallel()

	const (
		pluginID = "myorg-test-app"
		token    = "dummy-gcs-access-token"
	)
	withTokenURL := "https://storage.googleapis.com/integration-artifacts-private/with-token.zip"
	otherHostURL := "https://example.com/other-host.zip"
	noTokenURL := "https://storage.googleapis.com/integration-artifacts/no-token.zip"

	runner, err := act.NewRunner(t)
	require.NoError(t, err)

	content := newZip(t, map[string]string{
		pluginID + "/plugin.json":  `{"id":"` + pluginID + `"}`,
		pluginID + "/MANIFEST.txt": "-----BEGIN PGP SIGNED MESSAGE-----",
	})
	var mu sync.Mutex
	var requests []gcsMockRequest
	mock, err := act.NewCurlHTTPSMock(t, runner, []string{"storage.googleapis.com", "example.com"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, gcsMockRequest{URL: "https://" + r.Host + r.URL.Path, Authorization: r.Header.Get("Authorization")})
		mu.Unlock()
		_, _ = w.Write(content)
	}))
	require.NoError(t, err)

	checkArtifactsStep := func(name string, zips []string, gcsAccessToken string) workflow.Step {
		zipsJSON, err := json.Marshal(zips)
		require.NoError(t, err)
		return workflow.Step{
			Name: name,
			Uses: "grafana/plugin-ci-workflows/actions/internal/plugins/publish/check-artifacts@main",
			With: map[string]any{
				"zips":             string(zipsJSON),
				"plugin-id":        pluginID,
				"gcs-access-token": gcsAccessToken,
			},
			Env: mock.CurlEnv(),
		}
	}
	baseWf := workflow.BaseWorkflow{
		Name: "Check artifacts GCS access token test",
		On:   workflow.On{Push: workflow.OnPush{Branches: []string{"main"}}},
		Jobs: map[string]*workflow.Job{
			"check": {
				RunsOn: "ubuntu-arm64-small",
				Steps: workflow.Steps{
					checkArtifactsStep("Check with token", []string{withTokenURL, otherHostURL}, token),
					checkArtifactsStep("Check without token", []string{noTokenURL}, ""),
				},
			},
		},
	}
	wf := workflow.NewTestingWorkflow("check-artifacts-gcs-access-token", baseWf)
	wf.AddUUIDToAllJobsRecursive()

	r, err := runner.Run(wf, act.NewPushEventPayload("main"))
	require.NoError(t, err)
	require.True(t, r.Success, "workflow should succeed")

	mu.Lock()
	defer mu.Unlock()
	require.ElementsMatch(t, []gcsMockRequest{
		{URL: withTokenURL, Authorization: "Bearer " + token},
		{URL: otherHostURL},
		{URL: noTokenURL},
	}, requests)
}

type gcsMockRequest struct {
	URL           string
	Authorization string
}

func newZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		f, err := zw.Create(name)
		require.NoError(t, err)
		_, err = f.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, zw.Close())
	return buf.Bytes()
}
