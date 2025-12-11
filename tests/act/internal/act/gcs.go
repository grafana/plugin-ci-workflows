package act

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// GCS represents a mock Google Cloud Storage service for testing purposes.
// It provides a read-only filesystem interface based on afero.Fs, which can be used
// to verify the presence of uploaded files during tests.
type GCS struct {
	basePath string

	// Fs is a read-only afero filesystem chroot-ed into the mock GCS base path.
	// It can be used to verify the presence of uploaded files.
	Fs afero.Fs
}

// newGCS creates a new mock GCS instance for the given Runner.
func newGCS(r *Runner) (GCS, error) {
	path := filepath.Join("/tmp", "act-gcs", r.uuid.String())
	if err := os.MkdirAll(path, 0755); err != nil {
		return GCS{}, fmt.Errorf("mkdir mock gcs %q: %w", path, err)
	}
	return GCS{
		basePath: path,
		// Read only afero fs chroot-ed into the mock gcs path
		Fs: afero.NewReadOnlyFs(afero.NewBasePathFs(afero.NewOsFs(), path)),
	}, nil
}

// TODO: GCS.Close() ?
