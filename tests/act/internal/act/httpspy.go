package act

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// SpyCallInputs represents the inputs recorded from a single call to the HTTPSpy.
type SpyCallInputs struct {
	Inputs map[string]any
}

// HTTPSpy is a generic HTTP server that records incoming requests for testing.
// It can be used to mock any external service that workflow steps call via HTTP,
// recording the inputs for later assertions and returning configurable outputs.
//
// This is useful for mocking actions that run inside act's Docker containers,
// which cannot directly communicate with the Go test process.
type HTTPSpy struct {
	t       *testing.T
	server  *httptest.Server
	outputs map[string]string
	calls   []SpyCallInputs
	mux     sync.Mutex
}

// NewHTTPSpy creates a new HTTPSpy server that listens on all interfaces.
// The server is automatically cleaned up when the test finishes.
//
// The outputs parameter is a map of output names to values that will be returned
// as JSON for all POST requests. These correspond to action outputs that the
// calling step can parse and set as step outputs.
//
// The server records all incoming POST requests, parsing the JSON body as inputs.
//
// To make the server accessible from Docker containers (e.g., act containers),
// use DockerAccessibleURL() instead of the server's direct URL.
func NewHTTPSpy(t *testing.T, outputs map[string]string) *HTTPSpy {
	spy := &HTTPSpy{
		t:       t,
		outputs: outputs,
		calls:   make([]SpyCallInputs, 0),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /", spy.handleRequest)

	// Create a listener on all interfaces (0.0.0.0) so Docker containers can reach it
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		t.Fatalf("failed to create listener for HTTPSpy: %v", err)
	}

	server := &httptest.Server{
		Listener: listener,
		Config:   &http.Server{Handler: mux},
	}
	server.Start()

	spy.server = server

	t.Cleanup(func() {
		server.Close()
	})

	return spy
}

// handleRequest handles incoming POST requests, recording the JSON body as inputs
// and returning the configured outputs as JSON.
func (s *HTTPSpy) handleRequest(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.t.Logf("HTTPSpy: failed to read request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var inputs map[string]any
	if len(body) > 0 {
		if err := json.Unmarshal(body, &inputs); err != nil {
			s.t.Logf("HTTPSpy: failed to parse JSON body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	s.recordCall(SpyCallInputs{Inputs: inputs})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(s.outputs) //nolint:errcheck
}

// recordCall records a call with the given inputs.
func (s *HTTPSpy) recordCall(inputs SpyCallInputs) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.calls = append(s.calls, inputs)
}

// GetCalls returns all recorded calls.
// This method is thread-safe.
func (s *HTTPSpy) GetCalls() []SpyCallInputs {
	s.mux.Lock()
	defer s.mux.Unlock()
	// Return a copy to avoid race conditions
	result := make([]SpyCallInputs, len(s.calls))
	copy(result, s.calls)
	return result
}

// Reset clears all recorded calls.
// This method is thread-safe.
func (s *HTTPSpy) Reset() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.calls = make([]SpyCallInputs, 0)
}

// DockerAccessibleURL returns a URL that can be used from inside Docker containers.
// It uses host.docker.internal which works on:
//   - Docker Desktop (Mac/Windows): built-in support
//   - Linux: enabled via --add-host=host.docker.internal:host-gateway in act.go
func (s *HTTPSpy) DockerAccessibleURL() string {
	addr := s.server.Listener.Addr().(*net.TCPAddr)
	return fmt.Sprintf("http://host.docker.internal:%d", addr.Port)
}
