package checkpoint

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

type Body []byte
type Result struct {
	Headers    map[string]string
	StatusCode int
	Body       Body
}

// TestConfig holds the configuration for the Test function
type TestConfig struct {
	router      Router                                   // Required
	RouteFunc   func(http.ResponseWriter, *http.Request) // Required
	Path        string                                   // Required
	Headers     map[string]string                        // Nullable
	Middlewares []func(http.Handler) http.Handler        // Nullable
	URLPattern  *string                                  // Nullable
	Method      *string                                  // Nullable
	Body        *string                                  // Nullable
}

type HeaderFunc func() (string, string)

// WithHeaders adds headers to the TestConfig
func (tc *TestConfig) WithHeaders(headers ...HeaderFunc) *TestConfig {
	if tc.Headers == nil {
		tc.Headers = make(map[string]string)
	}
	for _, h := range headers {
		k, v := h()
		tc.Headers[k] = v
	}
	return tc
}

// Header creates a HeaderFunc
func Header(key string, value string) HeaderFunc {
	return func() (string, string) {
		return key, value
	}
}

// WithMiddlewares adds middlewares to the TestConfig
func (tc *TestConfig) WithMiddlewares(middlewares ...func(http.Handler) http.Handler) *TestConfig {
	tc.Middlewares = append(tc.Middlewares, middlewares...)
	return tc
}

// Run executes the test with the current configuration
func (tc *TestConfig) Run(ctx context.Context) (*Result, error) {
	// Validate required fields
	if tc.RouteFunc == nil {
		return nil, errors.New("handler cannot be nil")
	}
	if tc.Path == "" {
		return nil, errors.New("path cannot be empty")
	}

	// Set defaults for nullable fields
	method := "GET"
	if tc.Method != nil {
		method = *tc.Method
	}

	body := ""
	if tc.Body != nil {
		body = *tc.Body
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, tc.Path, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Add headers to request
	if len(tc.Headers) > 0 {
		for key, value := range tc.Headers {
			req.Header.Set(key, value)
		}
	}

	// Apply middlewares to handler
	finalHandler := http.Handler(http.HandlerFunc(tc.RouteFunc))
	if len(tc.Middlewares) > 0 {
		for i := len(tc.Middlewares) - 1; i >= 0; i-- {
			finalHandler = tc.Middlewares[i](finalHandler)
		}
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	urlPattern := tc.Path
	if tc.URLPattern != nil {
		urlPattern = *tc.URLPattern
	}
	tc.router.Handle(urlPattern, finalHandler)
	tc.router.ServeHTTP(rr, req)

	// Extract response headers
	responseHeaders := make(map[string]string)
	for key, values := range rr.Header() {
		if len(values) > 0 {
			responseHeaders[key] = strings.Join(values, ", ")
		}
	}

	// Read response body
	bodyBytes, err := io.ReadAll(rr.Body)
	if err != nil {
		return nil, err
	}

	return &Result{
		Headers:    responseHeaders,
		StatusCode: rr.Code,
		Body:       bodyBytes,
	}, nil
}

// Init creates a new TestConfig with a given router
func Init(r Router) *TestConfig {
	return &TestConfig{
		router: r,
	}
}

func (b Body) String() string {
	if len(b) == 0 {
		return ""
	}
	return string(b)
}
