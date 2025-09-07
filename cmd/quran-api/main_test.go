package main

import (
  "net/http"
  "net/http/httptest"
  "strings"
  "testing"
)

func TestAPI_InvalidSurahNumber(t *testing.T) {
  h := newRouter(nil)
  // below 1
  req := httptest.NewRequest(http.MethodGet, "/surah/0", nil)
  w := httptest.NewRecorder()
  h.ServeHTTP(w, req)
  if w.Code != http.StatusBadRequest {
    t.Fatalf("expected 400, got %d", w.Code)
  }
  // above 114
  req = httptest.NewRequest(http.MethodGet, "/surah/115", nil)
  w = httptest.NewRecorder()
  h.ServeHTTP(w, req)
  if w.Code != http.StatusBadRequest {
    t.Fatalf("expected 400, got %d", w.Code)
  }
}

func TestAPI_SearchTooLong(t *testing.T) {
  h := newRouter(nil)
  longQ := strings.Repeat("a", 101)
  req := httptest.NewRequest(http.MethodGet, "/search?q="+longQ, nil)
  w := httptest.NewRecorder()
  h.ServeHTTP(w, req)
  if w.Code != http.StatusBadRequest {
    t.Fatalf("expected 400, got %d", w.Code)
  }
}

func TestAPI_Healthz(t *testing.T) {
  h := newRouter(nil)
  req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
  w := httptest.NewRecorder()
  h.ServeHTTP(w, req)
  if w.Code != http.StatusOK {
    t.Fatalf("expected 200, got %d", w.Code)
  }
}

