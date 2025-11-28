package act

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/afero/zipfs"
)

type ArtifactFolder struct {
	Fs      afero.Fs
	rawFile *os.File
}

func (a *ArtifactFolder) Close() error {
	return a.rawFile.Close()
}

func (a *ArtifactFolder) open(fn string) (io.ReaderAt, error) {
	c, err := afero.ReadFile(a.Fs, fn)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(c), nil
}

func (a *ArtifactFolder) OpenZIP(fn string) (afero.Fs, error) {
	f, err := a.open(fn)
	if err != nil {
		return nil, fmt.Errorf("open nested zip %q: %w", fn, err)
	}
	finfo, err := a.Fs.Stat(fn)
	if err != nil {
		return nil, fmt.Errorf("stat %q: %w", fn, err)
	}
	zr, err := zip.NewReader(f, finfo.Size())
	if err != nil {
		return nil, fmt.Errorf("new zip reader %q: %w", fn, err)
	}
	return zipfs.New(zr), nil
}

func (a *ArtifactFolder) ReadFile(fn string) ([]byte, error) {
	return afero.ReadFile(a.Fs, fn)
}

type ArtifactsStorage struct {
	basePath string
}

func newDefaultArtifactsStorage() ArtifactsStorage {
	return ArtifactsStorage{
		basePath: "/tmp/artifacts",
	}
}

func (a ArtifactsStorage) runFolder(runID string) string {
	return filepath.Join(a.basePath, runID)
}

/* func (a ArtifactsStorage) List(runID string) (map[string]struct{}, error) {
	bd := a.runFolder(runID)
	artifacts := map[string]struct{}{}
	if err := filepath.WalkDir(bd, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return filepath.SkipDir
		}
		artifacts[d.Name()] = struct{}{}
		return nil
	}); err != nil {
		return nil, err
	}
	return artifacts, nil
} */

func (a ArtifactsStorage) GetFolder(runID string, artifactName string) (*ArtifactFolder, error) {
	bd := a.runFolder(runID)
	artifactFn := filepath.Join(bd, artifactName, artifactName+".zip")
	finfo, err := os.Stat(artifactFn)
	if err != nil {
		return nil, err
	}
	rawF, err := os.Open(artifactFn)
	if err != nil {
		return nil, err
	}
	zf, err := zip.NewReader(rawF, finfo.Size())
	if err != nil {
		return nil, err
	}
	return &ArtifactFolder{Fs: zipfs.New(zf), rawFile: rawF}, nil
}
