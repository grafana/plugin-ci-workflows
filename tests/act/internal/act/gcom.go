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

// GCOMRequest represents an intercepted GCOM API request.
type GCOMRequest struct {
	// Method is the HTTP method (GET, POST, etc.)
	Method string

	// Path is the URL path (e.g., "/plugins" or "/plugins/my-plugin")
	Path string

	// Headers contains the request headers
	Headers http.Header

	// Body contains the raw request body
	Body []byte
}

// GCOMResponse represents a mock response to return from the GCOM mock server.
type GCOMResponse struct {
	// StatusCode is the HTTP status code to return (e.g., 200, 404)
	StatusCode int

	// Body is the response body. If it's a struct/map, it will be JSON-encoded.
	// If it's a string or []byte, it will be sent as-is.
	Body any

	// Headers are additional headers to include in the response
	Headers map[string]string
}

// GCOM is a mock GCOM API server for testing CD workflows.
// It intercepts HTTP requests, records them for assertions, and returns
// configurable mock responses.
type GCOM struct {
	t      *testing.T
	server *httptest.Server
	mux    *http.ServeMux

	mu       sync.Mutex
	requests []GCOMRequest
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
		t:        t,
		server:   server,
		mux:      mux,
		requests: make([]GCOMRequest, 0),
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
	// Record the request
	body, _ := io.ReadAll(r.Body)
	m.recordRequest(GCOMRequest{
		Method:  r.Method,
		Path:    r.URL.Path,
		Headers: r.Header.Clone(),
		Body:    body,
	})

	// Return 404 for unhandled requests
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`{"code":"NotFound","message":"no mock handler registered for this path"}`)) //nolint:errcheck
}

// recordRequest safely records a request.
func (m *GCOM) recordRequest(req GCOMRequest) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = append(m.requests, req)
}

// URL returns the mock server URL (e.g., "http://127.0.0.1:12345").
// Use this for local testing.
func (m *GCOM) URL() string {
	return m.server.URL
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
// The handler receives the request body as a convenience.
//
// Example:
//
//	mock.HandleFunc("POST /plugins", func(w http.ResponseWriter, r *http.Request, body []byte) {
//	    w.WriteHeader(http.StatusOK)
//	    json.NewEncoder(w).Encode(map[string]any{"plugin": map[string]any{"id": "test-plugin"}})
//	})
func (m *GCOM) HandleFunc(pattern string, handler func(w http.ResponseWriter, r *http.Request, body []byte)) {
	m.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		// Read and record the request
		body, _ := io.ReadAll(r.Body)
		m.recordRequest(GCOMRequest{
			Method:  r.Method,
			Path:    r.URL.Path,
			Headers: r.Header.Clone(),
			Body:    body,
		})

		// Set default content type
		w.Header().Set("Content-Type", "application/json")

		// Call the user's handler
		handler(w, r, body)
	})
}

// OnRequest registers a simple handler that returns the provided response.
// This is a convenience method for common use cases.
//
// Example:
//
//	mock.OnRequest("GET /plugins/{pluginID}", GCOMResponse{
//	    StatusCode: 200,
//	    Body: map[string]any{"id": "test-plugin", "signatureType": "grafana"},
//	})
func (m *GCOM) OnRequest(pattern string, response GCOMResponse) {
	m.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request, body []byte) {
		// Set custom headers
		for k, v := range response.Headers {
			w.Header().Set(k, v)
		}

		// Write status code
		statusCode := response.StatusCode
		if statusCode == 0 {
			statusCode = http.StatusOK
		}
		w.WriteHeader(statusCode)

		// Write body
		if response.Body != nil {
			switch b := response.Body.(type) {
			case string:
				w.Write([]byte(b)) //nolint:errcheck
			case []byte:
				w.Write(b) //nolint:errcheck
			default:
				json.NewEncoder(w).Encode(b) //nolint:errcheck
			}
		}
	})
}

// Requests returns a copy of all recorded requests.
// Use this to assert on the requests made to the mock server.
func (m *GCOM) Requests() []GCOMRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]GCOMRequest, len(m.requests))
	copy(result, m.requests)
	return result
}

// ClearRequests clears all recorded requests.
func (m *GCOM) ClearRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = m.requests[:0]
}
