package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestIndexReturnsHTML(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	serveIndex(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusOK)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("unexpected Content-Type: %q", ct)
	}
	if !strings.Contains(rr.Body.String(), `<img src="/stream"`) {
		t.Error("response body missing <img src=\"/stream\"> tag")
	}
}

func TestStreamSetsMultipartContentType(t *testing.T) {
	h := NewHub()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	req := httptest.NewRequest("GET", "/stream", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	h.ServeStream(rr, req)

	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "multipart/x-mixed-replace") {
		t.Errorf("unexpected Content-Type: %q", ct)
	}
}
