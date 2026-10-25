package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/tbc/internal/forever/catalog"
	"github.com/wowsims/tbc/internal/forever/domain"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	return New(cat, nil)
}

func get(t *testing.T, handler http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestMeta(t *testing.T) {
	srv := newTestServer(t)
	rec := get(t, srv, "/api/meta")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		CatalogRevision string                       `json:"catalogRevision"`
		Counts          domain.ManifestCounts        `json:"counts"`
		Classes         []string                     `json:"classes"`
		Specs           map[string][]domain.SpecInfo `json:"specs"`
		Races           []string                     `json:"races"`
		ClassRaces      map[string][]string          `json:"classRaces"`
		DualWield       map[string]bool              `json:"dualWield"`
		Professions     []string                     `json:"professions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(body.CatalogRevision, "wowsims/forever@0cfb1a0:") {
		t.Fatalf("revision = %q", body.CatalogRevision)
	}
	if len(body.Classes) != 9 {
		t.Fatalf("classes = %d: %v", len(body.Classes), body.Classes)
	}
	if len(body.Specs["paladin"]) != 3 {
		t.Fatalf("paladin specs = %+v", body.Specs["paladin"])
	}
	if !slices.Contains(body.Races, "undead") {
		t.Fatalf("races = %v", body.Races)
	}
	if !slices.Contains(body.ClassRaces["paladin"], "undead") {
		t.Fatalf("paladin races = %v", body.ClassRaces["paladin"])
	}
	if body.DualWield["paladin_retribution"] {
		t.Fatal("paladin dual-wield")
	}
	if !body.DualWield["warrior_fury"] {
		t.Fatalf("warrior fury dual-wield = %v", body.DualWield)
	}
	if !slices.Contains(body.Professions, "blacksmithing") {
		t.Fatalf("professions = %v", body.Professions)
	}
	if body.Counts.ItemsTotal != 8599 || body.Counts.Talents != 465 {
		t.Fatalf("counts = %+v", body.Counts)
	}
}
