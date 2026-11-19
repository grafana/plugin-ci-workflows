package act

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// httpsMocksContainerPath is where the CurlHTTPSMock files are mounted in act containers.
const httpsMocksContainerPath = "/https-mocks"

// CurlHTTPSMock is an HTTPS server, running in the Go test, that curl in act containers
// reaches through real host names such as storage.googleapis.com, with TLS verification on.
//
// Only curl, in steps that set the environment variables returned by CurlEnv, is affected.
// curl in those steps connects the mocked host names to this server and trusts only its certificate,
// so HTTPS requests to other servers fail TLS verification. It is not network isolation.
type CurlHTTPSMock struct {
	containerDir string
}

// NewCurlHTTPSMock starts an HTTPS server that serves handler for the given host names (port 443).
// Create it before calling r.Run, because the runner mounts its files in the act containers.
// The server is automatically cleaned up when the test finishes.
func NewCurlHTTPSMock(t *testing.T, r *Runner, hosts []string, handler http.Handler) (*CurlHTTPSMock, error) {
	if len(hosts) == 0 {
		return nil, errors.New("at least one host name is required")
	}
	cert, certPEM, err := newSelfSignedCertificate(hosts)
	if err != nil {
		return nil, fmt.Errorf("create certificate: %w", err)
	}

	// Listen on all interfaces so Docker containers can reach the server.
	listener, err := listenFreePort()
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}
	server := &httptest.Server{
		Listener: listener,
		Config:   &http.Server{Handler: handler},
		TLS:      &tls.Config{Certificates: []tls.Certificate{cert}},
	}
	server.StartTLS()
	t.Cleanup(server.Close)

	// Each mock gets its own subdirectory in the runner's directory, which containerOptions mounts.
	// The first mock creates the runner's directory and removes it when the runner's test ends.
	// This contains the self-signed certificate and a .curlrc file that makes curl connect to the mocked https server.
	if r.httpsMocksPath == "" {
		r.httpsMocksPath = filepath.Join("/tmp", "act-https-mocks", r.uuid.String())
		mocksPath := r.httpsMocksPath
		r.t.Cleanup(func() { _ = os.RemoveAll(mocksPath) })
	}
	id := uuid.New().String()
	dir := filepath.Join(r.httpsMocksPath, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir %q: %w", dir, err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	var curlrcBuilder strings.Builder
	for _, host := range hosts {
		fmt.Fprintf(&curlrcBuilder, `connect-to = "%s:443:host.docker.internal:%d"`, host, port)
		curlrcBuilder.WriteRune('\n')
	}
	curlrcBuilder.WriteString(`noproxy = "*"`)
	if err := os.WriteFile(filepath.Join(dir, ".curlrc"), []byte(curlrcBuilder.String()), 0o644); err != nil {
		return nil, fmt.Errorf("write .curlrc: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cert.pem"), certPEM, 0o644); err != nil {
		return nil, fmt.Errorf("write certificate: %w", err)
	}

	return &CurlHTTPSMock{containerDir: httpsMocksContainerPath + "/" + id}, nil
}

// CurlEnv returns the environment variables that make curl in a step use this mock.
//
// Example:
//
//	mock, err := act.NewCurlHTTPSMock(t, runner, []string{"storage.googleapis.com"}, handler)
//	require.NoError(t, err)
//	step := workflow.Step{
//		Uses: "grafana/plugin-ci-workflows/actions/internal/plugins/publish/check-artifacts@main",
//		With: map[string]any{"zips": `["https://storage.googleapis.com/bucket/plugin.zip"]`},
//		Env:  mock.CurlEnv(),
//	}
func (m *CurlHTTPSMock) CurlEnv() map[string]string {
	return map[string]string{
		"CURL_HOME":      m.containerDir,
		"CURL_CA_BUNDLE": m.containerDir + "/cert.pem",
	}
}

// newSelfSignedCertificate returns a TLS certificate valid for the given host names, and its PEM encoding.
func newSelfSignedCertificate(hosts []string) (tls.Certificate, []byte, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("generate key: %w", err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: hosts[0]},
		DNSNames:              hosts,
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("create certificate: %w", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, certPEM, nil
}
