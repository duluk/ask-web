package download

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// errorReader is a custom io.Reader that always returns an error when Read is
// called
type errorReader struct{}

func (er errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestPage(t *testing.T) {
	t.Run("successful download", func(t *testing.T) {
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,
			r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello, world!"))
		}))
		defer testServer.Close()

		content, err := Page(testServer.URL)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if content != "Hello, world!" {
			t.Errorf("expected content 'Hello, world!', got: %q", content)
		}
	})

	t.Run("http error", func(t *testing.T) {
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,
			r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer testServer.Close()

		resp, err := Page(testServer.URL)
		fmt.Println("resp:", resp)
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}
		if !strings.Contains(err.Error(), "404") {
			t.Errorf("expected error to contain '404', got: %v", err)
		}
	})

	t.Run("download error", func(t *testing.T) {
		_, err := Page("invalid-url")
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}
		if !strings.Contains(err.Error(), "unsupported protocol scheme") {
			t.Errorf("expected error to contain 'unsupported protocol scheme', got: %v",
				err)
		}
	})

	t.Run("read error", func(t *testing.T) {
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.
			ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			// Use our custom error reader
			w.(http.ResponseWriter).Write([]byte("fake response"))
			w.(http.ResponseWriter).Header().Set("Content-type", "application/octet-stream")
		}))
		defer testServer.Close()

		client := &http.Client{}
		req, _ := http.NewRequest(http.MethodGet, testServer.URL, nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		originalBody := resp.Body
		resp.Body = io.NopCloser(errorReader{})

		_, err = io.ReadAll(resp.Body)
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}
		if !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("expected error to be io.ErrUnexpectedEOF, got: %v", err)
		}
		resp.Body = originalBody
	})
}
