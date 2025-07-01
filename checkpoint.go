package checkpoint

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

// Result represents the response from a test execution
type Result struct {
	Headers    map[string]string
	StatusCode int
	Body       []byte
}

type TestOption func(config *testConfig)

// testConfig holds the configuration for the Test function
type testConfig struct {
	headers     map[string]string
	middlewares []func(http.Handler) http.Handler
}

// Check executes a test request against the provided handler with the given parameters
func Check(
	ctx context.Context,
	urlPath string,
	urlPattern string,
	middlewares TestOption,
	method string,
	headers TestOption,
	body string,
	handler http.Handler,
) (*Result, error) {
	// Apply functional options
	config := &testConfig{
		headers:     make(map[string]string),
		middlewares: []func(http.Handler) http.Handler{},
	}
	if headers != nil {
		headers(config)
	}
	if middlewares != nil {
		middlewares(config)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, urlPath, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Add headers to request
	for key, value := range config.headers {
		req.Header.Set(key, value)
	}

	// Apply middlewares to handler
	finalHandler := handler
	for i := len(config.middlewares) - 1; i >= 0; i-- {
		finalHandler = config.middlewares[i](finalHandler)
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute request
	finalHandler.ServeHTTP(rr, req)

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

	// Return result
	result := &Result{
		Headers:    responseHeaders,
		StatusCode: rr.Code,
		Body:       bodyBytes,
	}

	return result, nil
}

func WithHeaders(headers ...TestOption) TestOption {
	return func(config *testConfig) {
		for _, addHeaderTo := range headers {
			addHeaderTo(config)
		}
	}
}

// Header sets the headers for the request
func Header(key, value string) TestOption {
	return func(config *testConfig) {
		config.headers[key] = value
	}
}

// WithMiddlewares sets the middlewares to be applied to the handler
func WithMiddlewares(middlewares ...func(http.Handler) http.Handler) TestOption {
	return func(config *testConfig) {
		for _, middleware := range middlewares {
			config.middlewares = append(config.middlewares, middleware)
		}
	}
}
