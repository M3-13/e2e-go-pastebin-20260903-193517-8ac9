package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthEndpointReturns200(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	newRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /health returned %d, want 200", rec.Code)
	}
}

func TestAccessLogLineContainsMethodPathStatusDuration(t *testing.T) {
	var buf bytes.Buffer
	handler := loggingMiddleware(&buf, newRouter())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /health returned %d, want 200", rec.Code)
	}

	line := buf.String()
	if line == "" {
		t.Fatal("expected an access log line, got none")
	}
	fields := strings.Fields(line)
	if len(fields) != 4 {
		t.Fatalf("access log line should have 4 fields (method path status duration), got %d in %q", len(fields), line)
	}
	if fields[0] != http.MethodGet {
		t.Errorf("log method = %q, want %q", fields[0], http.MethodGet)
	}
	if fields[1] != "/health" {
		t.Errorf("log path = %q, want %q", fields[1], "/health")
	}
	if fields[2] != "200" {
		t.Errorf("log status = %q, want %q", fields[2], "200")
	}
	if fields[3] == "" {
		t.Error("log duration should be non-empty")
	}
}
