package checkpoint

import (
	"context"
	"errors"
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
	urlPath     string
	urlPattern  string
	method      string
	body        string
}

type HeaderFunc func() (string, string)

func WithHeaders(headers ...HeaderFunc) TestOption {
	return func(config *testConfig) {
		if config.headers == nil {
			config.headers = make(map[string]string)
		}
		for _, h := range headers {
			k, v := h()
			config.headers[k] = v
		}
	}
}

func Header(key string, value string) HeaderFunc {
	return func() (string, string) {
		return key, value
	}
}

func WithMiddlewares(middlewares ...func(http.Handler) http.Handler) TestOption {
	return func(config *testConfig) {
		config.middlewares = append(config.middlewares, middlewares...)
	}
}

func WithURLPath(urlPath string) TestOption {
	return func(config *testConfig) {
		config.urlPath = urlPath
	}
}

func WithURLPattern(urlPattern string) TestOption {
	return func(config *testConfig) {
		config.urlPattern = urlPattern
	}
}

func WithMethod(method string) TestOption {
	return func(config *testConfig) {
		config.method = method
	}
}

func WithBody(body string) TestOption {
	return func(config *testConfig) {
		config.body = body
	}
}

type CheckFunc func(
	ctx context.Context,
	handler http.Handler,
	options ...TestOption,
) (*Result, error)

func NewChecker[T Router](router T) CheckFunc {
	return func(
		ctx context.Context,
		handler http.Handler,
		options ...TestOption,
	) (*Result, error) {
		config := &testConfig{
			headers:     make(map[string]string),
			middlewares: []func(http.Handler) http.Handler{},
			method:      "GET", // default method
		}

		// Apply all options
		for _, option := range options {
			option(config)
		}
		if config.urlPath == "" {
			return nil, errors.New("url path cannot be empty")
		}
		// Create request
		req, err := http.NewRequestWithContext(ctx, config.method, config.urlPath, strings.NewReader(config.body))
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

		router.Handle(config.urlPattern, finalHandler)
		router.ServeHTTP(rr, req)
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
}
