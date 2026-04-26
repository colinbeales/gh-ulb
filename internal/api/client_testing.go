package api

import (
	"net/http"
	"strings"

	ghapi "github.com/cli/go-gh/v2/pkg/api"
)

// testTransport redirects all outgoing requests to a fixed test server base URL,
// preserving the original path and query string.
type testTransport struct {
	serverURL string // e.g. "http://127.0.0.1:12345"
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	target := t.serverURL + req.URL.Path
	if req.URL.RawQuery != "" {
		target += "?" + req.URL.RawQuery
	}
	newReq, err := http.NewRequestWithContext(req.Context(), req.Method, target, req.Body)
	if err != nil {
		return nil, err
	}
	for k, v := range req.Header {
		newReq.Header[k] = v
	}
	return http.DefaultTransport.RoundTrip(newReq)
}

// NewTestClient creates a Client that routes all requests to serverURL.
// Intended for use in tests only.
func NewTestClient(serverURL string) (*Client, error) {
	opts := ghapi.ClientOptions{
		Host:      "github.com",
		AuthToken: "test-token",
		Transport: &testTransport{serverURL: strings.TrimSuffix(serverURL, "/")},
	}
	rest, err := ghapi.NewRESTClient(opts)
	if err != nil {
		return nil, err
	}
	return &Client{rest: rest}, nil
}
