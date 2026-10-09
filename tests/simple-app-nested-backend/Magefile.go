//go:build mage
// +build mage

package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var osArchCombos = [][2]string{
	{"darwin", "amd64"},
	{"darwin", "arm64"},
	{"linux", "amd64"},
	{"linux", "arm"},
	{"linux", "arm64"},
	{"windows", "amd64"},
}

// Both the app itself and its nested datasource have a backend. This is the
// whole point of this fixture: per-platform packaging must handle multiple
// backend executables.
var backends = []struct {
	outDir string
	exe    string
	pkg    string
}{
	{"dist", "gpx_simpleappnested_app", "./pkg"},
	{"dist/datasource", "gpx_simpleappnested_datasource", "./pkg/datasource"},
}

// Default configures the default target.
var Default = BuildAll

// BuildAll builds the app and nested datasource backend binaries for all
// supported os/arch combos into dist/, mirroring the layout the SDK build
// target produces for regular plugins (including go_plugin_build_manifest).
func BuildAll() error {
	manifest := ""
	for _, backend := range backends {
		if err := os.MkdirAll(backend.outDir, 0o755); err != nil {
			return err
		}
		for _, combo := range osArchCombos {
			goos, goarch := combo[0], combo[1]
			out := filepath.Join(backend.outDir, fmt.Sprintf("%s_%s_%s", backend.exe, goos, goarch))
			if goos == "windows" {
				out += ".exe"
			}
			fmt.Println("building", out)
			cmd := exec.Command("go", "build", "-o", out, backend.pkg)
			cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch, "CGO_ENABLED=0")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}

			content, err := os.ReadFile(out)
			if err != nil {
				return err
			}
			rel, err := filepath.Rel("dist", out)
			if err != nil {
				return err
			}
			manifest += fmt.Sprintf("%x:%s\n", sha256.Sum256(content), rel)
		}
	}
	return os.WriteFile(filepath.Join("dist", "go_plugin_build_manifest"), []byte(manifest), 0o755)
}
