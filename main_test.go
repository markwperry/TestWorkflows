package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHelloWorldEndpoint(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/helloworld", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	if body["message"] != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got '%s'", body["message"])
	}
}

func TestHealthEndpoint(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("expected 'ok', got '%v'", body["status"])
	}

	// Verify new fields are present
	for _, field := range []string{"version", "goVersion", "memAlloc", "goroutines"} {
		if _, exists := body[field]; !exists {
			t.Errorf("expected field '%s' in health response", field)
		}
	}
}
