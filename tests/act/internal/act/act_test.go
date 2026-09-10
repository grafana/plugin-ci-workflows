package act

import (
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestPortAllocationDoesNotOverlap checks that a port reserved by getFreePort for an act
// artifact server is never handed to a listener created by listenFreePort.
//
// The mock servers used to bind their ports with net.Listen on port 0, which does not
// consult takenPorts. The kernel could hand them a port that getFreePort had already
// promised to an act artifact server that had not bound it yet, and act then died with
// "bind: address already in use".
func TestPortAllocationDoesNotOverlap(t *testing.T) {
	t.Parallel()

	// Reserve a batch of ports the way the act runner does: registered, but not bound.
	const reserved = 16
	reservedPorts := make(map[int]struct{}, reserved)
	for range reserved {
		port, err := getFreePort()
		require.NoError(t, err)
		require.NotContains(t, reservedPorts, port, "getFreePort handed out the same port twice")
		reservedPorts[port] = struct{}{}
		t.Cleanup(func() { markPortAsFree(port) })
	}

	// Now take listeners the way the mock servers do. None may land on a reserved port.
	for range 64 {
		listener, err := listenFreePort()
		require.NoError(t, err)
		t.Cleanup(func() { _ = listener.Close() })

		port := listener.Addr().(*net.TCPAddr).Port
		require.NotContains(
			t, reservedPorts, port,
			"listenFreePort took port %d, which is reserved for an act artifact server", port,
		)
	}

	// Every reserved port must still be bindable, which is what act does with it.
	for port := range reservedPorts {
		listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
		require.NoError(t, err, "reserved port %d was not free when act tried to bind it", port)
		require.NoError(t, listener.Close())
	}
}

// TestListenFreePortReleasesOnClose checks that closing a listener returned by
// listenFreePort releases its takenPorts registration, so the port can be reused.
func TestListenFreePortReleasesOnClose(t *testing.T) {
	t.Parallel()

	listener, err := listenFreePort()
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port

	_, taken := takenPorts.Load(port)
	require.True(t, taken, "port should be registered while the listener is open")

	require.NoError(t, listener.Close())

	_, taken = takenPorts.Load(port)
	require.False(t, taken, "port should be released once the listener is closed")
}

// TestPortAllocationIsConcurrencySafe hammers both allocators from many goroutines and
// checks that no port is ever handed out twice.
func TestPortAllocationIsConcurrencySafe(t *testing.T) {
	t.Parallel()

	const goroutines = 32

	var (
		mu    sync.Mutex
		seen  = map[int]string{}
		wg    sync.WaitGroup
		claim = func(port int, via string) {
			mu.Lock()
			defer mu.Unlock()
			previous, dup := seen[port]
			require.False(t, dup, "port %d handed out by both %s and %s", port, previous, via)
			seen[port] = via
		}
	)

	wg.Add(goroutines * 2)
	for range goroutines {
		go func() {
			defer wg.Done()
			port, err := getFreePort()
			if !assertNoError(t, err) {
				return
			}
			claim(port, "getFreePort")
		}()
		go func() {
			defer wg.Done()
			listener, err := listenFreePort()
			if !assertNoError(t, err) {
				return
			}
			claim(listener.Addr().(*net.TCPAddr).Port, "listenFreePort")
		}()
	}
	wg.Wait()
}

// assertNoError reports err as a test failure from a goroutine and returns whether it was nil.
func assertNoError(t *testing.T, err error) bool {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return false
	}
	return true
}
