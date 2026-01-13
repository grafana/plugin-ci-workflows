package act

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

// GCOM is a mock GCOM API server for testing CD workflows.
// It intercepts HTTP requests, records them for assertions, and returns
// configurable mock responses.
type GCOM struct {
	t      *testing.T
	server *httptest.Server
	mux    *http.ServeMux
}

// newGCOM creates a new GCOM mock server that listens on all interfaces.
// The server is automatically cleaned up when the test finishes.
//
// To make the server accessible from Docker containers (e.g., act containers),
// use DockerAccessibleURL() instead of URL().
func newGCOM(t *testing.T) *GCOM {
	mux := http.NewServeMux()

	// Create a listener on all interfaces (0.0.0.0) so Docker containers can reach it
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		t.Fatalf("failed to create listener for GCOM mock: %v", err)
	}

	server := &httptest.Server{
		Listener: listener,
		Config:   &http.Server{Handler: mux},
	}
	server.Start()

	mock := &GCOM{
		t:      t,
		server: server,
		mux:    mux,
	}

	// Register a catch-all handler to record requests that don't match any registered pattern
	mux.HandleFunc("/", mock.catchAllHandler)

	t.Cleanup(func() {
		server.Close()
	})

	return mock
}

// catchAllHandler records requests and returns 404 for unhandled paths.
func (m *GCOM) catchAllHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`{"code":"NotFound","message":"no mock handler registered for this path"}`)) //nolint:errcheck
}

// DockerAccessibleURL returns a URL that can be used from inside Docker containers.
// It uses host.docker.internal which works on:
//   - Docker Desktop (Mac/Windows): built-in support
//   - Linux: enabled via --add-host=host.docker.internal:host-gateway in act.go
//
// The returned URL includes the /api path prefix that the GCOM scripts expect.
func (m *GCOM) DockerAccessibleURL() string {
	// Extract just the port from the server URL
	addr := m.server.Listener.Addr().(*net.TCPAddr)
	return fmt.Sprintf("http://host.docker.internal:%d/api", addr.Port)
}

// HandleFunc registers a handler for the given pattern.
// The pattern follows the http.ServeMux pattern syntax (e.g., "GET /plugins", "POST /plugins/{pluginID}").
//
// Example:
//
//	mock.HandleFunc("POST /plugins", func(w http.ResponseWriter, r *http.Request) {
//		// Some assertions...
//		require.Equal(t, r.Header["Accept"], "application/json")
//
//		// Response for the client
//	    w.WriteHeader(http.StatusOK)
//	    json.NewEncoder(w).Encode(map[string]any{"plugin": map[string]any{"id": "test-plugin"}})
//	})
func (m *GCOM) HandleFunc(pattern string, handler http.HandlerFunc) {
	m.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		handler(w, r)
	})
}
