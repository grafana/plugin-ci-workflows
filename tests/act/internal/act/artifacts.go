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

// ArtifactFolder represents a folder containing artifacts uploaded during a workflow run
// via the actions/upload-artifact action.
// On the local filesystem, artifacts are stored as ZIP files. This struct provides
// an abstraction over that, allowing access to the contents of the artifact as a folder.
// It uses an afero.Fs to provide access to the files within the artifact.
// The caller should call Close() when done using the ArtifactFolder to release resources.
type ArtifactFolder struct {
	Fs      afero.Fs
	rawFile *os.File
}

// Close closes the underlying ZIP file that is used by the ArtifactFolder.
func (a *ArtifactFolder) Close() error {
	return a.rawFile.Close()
}

// open opens a file within the artifact folder and returns a ReaderAt for it.
func (a *ArtifactFolder) open(fn string) (io.ReaderAt, error) {
	c, err := afero.ReadFile(a.Fs, fn)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(c), nil
}

// OpenZIP opens a nested ZIP file within the artifact folder and returns an afero.Fs for it.
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

// ReadFile reads a file within the artifact folder and returns its content.
func (a *ArtifactFolder) ReadFile(fn string) ([]byte, error) {
	return afero.ReadFile(a.Fs, fn)
}

// ArtifactsStorage manages the storage of artifacts uploaded during workflow runs.
// An ArtifactStorage can hold multiple artifacts from different workflow runs.
// Each artifact is identified by a run ID and artifact name
// and is stored as a ZIP file on the local filesystem.
type ArtifactsStorage struct {
	// basePath is the base path where artifacts are stored.
	// Each artifact is stored as a ZIP file, containing the uploaded files.
	basePath string
}

// newArtifactsStorage creates a new ArtifactsStorage instance for the given Runner.
func newArtifactsStorage(r *Runner) ArtifactsStorage {
	return ArtifactsStorage{
		basePath: "/tmp/act-artifacts/" + r.uuid.String() + "/",
	}
}

// runFolder returns the folder path for the given run ID.
func (a ArtifactsStorage) runFolder(runID string) string {
	return filepath.Join(a.basePath, runID)
}

// GetFolder retrieves the artifact folder for the given run ID and artifact name.
// The caller should call Close() on the returned ArtifactFolder when done using it.
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

// TODO: ArtifactsStorage.Close() ?
