package versions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLatestNodeVersion(t *testing.T) {
	version, err := LatestNodeVersion("24")
	require.NoError(t, err)
	require.Regexp(t, `^v24\.\d+\.\d+$`, version, "should match v24.x.y pattern")
	t.Logf("Latest Node.js 24.x: %s", version)
}

func TestLatestGoVersion(t *testing.T) {
	version, err := LatestGoVersion("1.25")
	require.NoError(t, err)
	require.Regexp(t, `^1\.25(\.\d+)?$`, version, "should match 1.25 or 1.25.x pattern")
	t.Logf("Latest Go 1.25.x: %s", version)
}
