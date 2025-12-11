package act

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type GCS struct {
	basePath string
}

func newGCS(r *Runner) (GCS, error) {
	path := filepath.Join("/tmp", "gcs", r.uuid.String())
	if err := os.MkdirAll(path, 0755); err != nil {
		return GCS{}, fmt.Errorf("mkdir mock gcs %q: %w", path, err)
	}
	return GCS{basePath: path}, nil
}

func (g *GCS) Get(fn string) (io.ReadCloser, error) {
	fn = sanitizeGCSFileName(fn)
	f, err := os.Open(filepath.Join(g.basePath, fn))
	if err != nil {
		return nil, fmt.Errorf("open mock gcs path %q: %w", g.basePath, err)
	}
	return f, nil
}

func (g *GCS) List(fn string) ([]string, error) {
	fn = sanitizeGCSFileName(fn)
	dirPath := filepath.Join(g.basePath, fn)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("read dir mock gcs path %q: %w", dirPath, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}

func sanitizeGCSFileName(fn string) string {
	return filepath.FromSlash(fn)
}
