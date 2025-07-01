package checkpoint

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_RunBodyPassthrough(t *testing.T) {
	ctx := context.Background()
	urlPath := "/test"
	urlPattern := "/test"
	method := "GET"
	body := "request body content"

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		bytes, err := io.ReadAll(r.Body)
		assert.NoError(t, err, "Failed to read request body")
		defer func() {
			_ = r.Body.Close()
		}()

		_, _ = w.Write(bytes)
	})

	// Check the test
	result, err := Check(
		ctx,
		urlPath,
		urlPattern,
		nil,
		method,
		nil,
		body,
		handler)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if result.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, result.StatusCode)
	}

	if !strings.Contains(string(result.Body), `request body content`) {
		t.Errorf("Expected body to contain 'success', got %s", string(result.Body))
	}
}

func Test_RunWithHeadersPassthrough(t *testing.T) {
	ctx := context.Background()
	urlPath := "/test"
	urlPattern := "/test"
	method := "GET"
	body := ""

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Test-Header", r.Header.Get("X-Test-Header"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "success"}`))
	})

	// Check the test with headers
	result, err := Check(
		ctx,
		urlPath,
		urlPattern,
		WithHeaders(
			Header("X-Test-Header", "TestValue"),
		),
		method,
		nil,
		body,
		handler)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if result.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, result.StatusCode)
	}

	assert.Equal(t, "TestValue", result.Headers["X-Test-Header"], "Expected header X-Test-Header to be 'TestValue'")
}

func Test_RunWithMiddlewares(t *testing.T) {
	ctx := context.Background()
	urlPath := "/test"
	urlPattern := "/test"
	method := "GET"
	body := ""

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Middleware-Header", r.Header.Get("X-Middleware-Header"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "success"}`))

	})

	// Middleware to add a custom header
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Middleware-Header", "MiddlewareValue")
			next.ServeHTTP(w, r)
		})
	}

	// Check the test with middleware
	result, err := Check(
		ctx,
		urlPath,
		urlPattern,
		WithMiddlewares(middleware),
		method,
		nil,
		body,
		handler)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if result.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, result.StatusCode)
	}

	assert.Equal(t, "MiddlewareValue", result.Headers["X-Middleware-Header"], "Expected header X-Middleware-Header to be 'MiddlewareValue'")
}

func Test_RunWithMiddlewaresStacked(t *testing.T) {
	ctx := context.Background()
	urlPath := "/test"
	urlPattern := "/test"
	method := "GET"
	body := ""

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Middleware-Header-Combined", r.Header.Get("X-Middleware-Header-1")+":"+r.Header.Get("X-Middleware-Header-2"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "success"}`))
	})

	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Middleware-Header-1", "MiddlewareValue1")
			next.ServeHTTP(w, r)
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := r.Header["X-Middleware-Header-1"]; ok {
				r.Header.Set("X-Middleware-Header-2", "MiddlewareValue2")
			}
			next.ServeHTTP(w, r)
		})
	}

	// Check the test with multiple middlewares
	result, err := Check(
		ctx,
		urlPath,
		urlPattern,
		WithMiddlewares(
			middleware1,
			middleware2,
		),
		method,
		nil,
		body,
		handler)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if result.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, result.StatusCode)
	}

	assert.Equal(
		t,
		"MiddlewareValue1:MiddlewareValue2",
		result.Headers["X-Middleware-Header-Combined"],
		"Expected header X-Middleware-Header to be 'MiddlewareValue1:MiddlewareValue2'")
}

func Test_RunWithMiddlewaresError(t *testing.T) {
	ctx := context.Background()
	urlPath := "/test"
	urlPattern := "/test"
	method := "GET"
	body := ""

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Middleware error", http.StatusInternalServerError)
		})
	}

	// Check the test with middleware that returns an error
	result, err := Check(
		ctx,
		urlPath,
		urlPattern,
		WithMiddlewares(middleware),
		method,
		nil,
		body,
		handler)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
}
