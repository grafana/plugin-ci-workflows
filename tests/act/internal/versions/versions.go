// Package versions provides functions to fetch the latest versions of Node.js and Go
// from their official release APIs.
package versions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

const (
	nodeReleasesURL = "https://nodejs.org/download/release/index.json"
	goReleasesURL   = "https://golang.org/dl/?mode=json&include=all"
	httpTimeout     = 30 * time.Second
)

type nodeRelease struct {
	Version string `json:"version"` // e.g., "v24.12.0"
}

type goRelease struct {
	Version string `json:"version"` // e.g., "go1.25.6"
	Stable  bool   `json:"stable"`
}

// LatestNodeVersion returns the latest Node.js version for a given major version.
// For example, LatestNodeVersion("24") might return "v24.12.0".
func LatestNodeVersion(major string) (string, error) {
	releases, err := fetchJSON[[]nodeRelease](nodeReleasesURL)
	if err != nil {
		return "", fmt.Errorf("fetch node releases: %w", err)
	}

	// Match versions like "v24.x.y" for major "24"
	prefix := "v" + major + "."
	for _, r := range releases {
		if strings.HasPrefix(r.Version, prefix) {
			// The releases are sorted newest first, so the first match is the latest
			return r.Version, nil
		}
	}

	return "", fmt.Errorf("no releases found for Node.js major version %q", major)
}

// LatestGoVersion returns the latest Go version for a given major.minor version.
// For example, LatestGoVersion("1.25") might return "1.25.6".
func LatestGoVersion(majorMinor string) (string, error) {
	releases, err := fetchJSON[[]goRelease](goReleasesURL)
	if err != nil {
		return "", fmt.Errorf("fetch go releases: %w", err)
	}

	prefix := "go" + majorMinor
	var candidates []string

	for _, r := range releases {
		if !r.Stable {
			continue
		}
		// Match "go1.25" or "go1.25.x"
		if r.Version == prefix || strings.HasPrefix(r.Version, prefix+".") {
			// Convert to semver format: "go1.25.6" -> "v1.25.6"
			v := "v" + strings.TrimPrefix(r.Version, "go")
			candidates = append(candidates, v)
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no stable releases found for Go version %q", majorMinor)
	}

	semver.Sort(candidates)
	latest := candidates[len(candidates)-1]

	// Convert back: "v1.25.6" -> "1.25.6"
	return strings.TrimPrefix(latest, "v"), nil
}

func fetchJSON[T any](url string) (T, error) {
	var result T

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return result, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("fetch %s: status %d", url, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("decode JSON from %s: %w", url, err)
	}

	return result, nil
}
