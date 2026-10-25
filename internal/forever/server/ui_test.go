package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServesBuiltIndex(t *testing.T) {
	ui := os.DirFS(filepath.Join("..", "..", "..", "cmd", "forever-armory", "webui"))
	srv := newTestServerWithUI(t, ui)
	rec := get(t, srv, "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `<div id="app">`) {
		t.Fatal("index")
	}
}
