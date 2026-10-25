package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wowsims/tbc/internal/forever/domain"
)

func post(t *testing.T, handler http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	return request(t, handler, http.MethodPost, path, body)
}

func patch(t *testing.T, handler http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	return request(t, handler, http.MethodPatch, path, body)
}

func del(t *testing.T, handler http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	return request(t, handler, http.MethodDelete, path, "")
}

func request(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(rec, req)
	return rec
}

func createCharacterRequest(t *testing.T, handler http.Handler, name, class, spec, race string, level int) domain.Character {
	t.Helper()
	body := fmt.Sprintf(
		`{"name":%q,"class":%q,"spec":%q,"race":%q,"level":%d}`,
		name, class, spec, race, level,
	)
	rec := post(t, handler, "/api/characters", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create character status %d: %s", rec.Code, rec.Body.String())
	}
	var ch domain.Character
	if err := json.Unmarshal(rec.Body.Bytes(), &ch); err != nil {
		t.Fatal(err)
	}
	if ch.ID == "" {
		t.Fatal("no id")
	}
	return ch
}

func listBuildsFor(t *testing.T, handler http.Handler, characterID string) []domain.Build {
	t.Helper()
	rec := get(t, handler, "/api/characters/"+characterID+"/builds")
	if rec.Code != http.StatusOK {
		t.Fatalf("list builds status %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Builds []domain.Build `json:"builds"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return body.Builds
}

func findBuildKind(builds []domain.Build, kind string) domain.Build {
	for _, build := range builds {
		if build.Kind == kind {
			return build
		}
	}
	return domain.Build{}
}

func TestCreateCharacterAndPlan(t *testing.T) {
	srv := newTestServer(t)
	rec := post(t, srv, "/api/characters",
		`{"name":"Elara","class":"paladin","spec":"paladin_retribution","race":"human","level":60}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var ch domain.Character
	if err := json.Unmarshal(rec.Body.Bytes(), &ch); err != nil {
		t.Fatal(err)
	}
	if ch.ID == "" {
		t.Fatal("no id")
	}

	if builds := listBuildsFor(t, srv, ch.ID); findBuildKind(builds, "current").ID == "" {
		t.Fatal("no current")
	}

	plan := post(t, srv, "/api/characters/"+ch.ID+"/builds", `{"name":"Pre-raid"}`)
	if plan.Code != http.StatusCreated {
		t.Fatalf("plan %d: %s", plan.Code, plan.Body.String())
	}
}

func TestUpdateBuildConflictsAndAtomicity(t *testing.T) {
	srv := newTestServer(t)
	ch := createCharacterRequest(t, srv, "Mira", "mage", "mage_frost", "gnome", 60)
	current := findBuildKind(listBuildsFor(t, srv, ch.ID), "current")
	if current.ID == "" {
		t.Fatal("no current")
	}

	// A valid save bumps the version.
	rec := patch(t, srv, "/api/builds/"+current.ID,
		`{"version":1,"equipment":{"head":{"itemId":208424}}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid save %d: %s", rec.Code, rec.Body.String())
	}
	var saved domain.Build
	if err := json.Unmarshal(rec.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Version != 2 {
		t.Fatalf("version = %d", saved.Version)
	}

	// A stale version conflicts and changes nothing.
	rec = patch(t, srv, "/api/builds/"+current.ID,
		`{"version":1,"equipment":{"head":{"itemId":208424}}}`)
	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), "version_conflict") {
		t.Fatalf("stale version %d: %s", rec.Code, rec.Body.String())
	}

	// Plate on a mage is rejected atomically.
	rec = patch(t, srv, "/api/builds/"+current.ID,
		`{"version":2,"equipment":{"head":{"itemId":23019}}}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid save %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "incompatible_armor") {
		t.Fatalf("violations: %s", rec.Body.String())
	}

	// The validate endpoint returns findings as data and writes nothing.
	rec = post(t, srv, "/api/characters/"+ch.ID+"/validate",
		`{"equipment":{"head":{"itemId":23019}},"talents":{}}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "incompatible_armor") {
		t.Fatalf("validate %d: %s", rec.Code, rec.Body.String())
	}

	storedRec := get(t, srv, "/api/builds/"+current.ID)
	var stored domain.Build
	if err := json.Unmarshal(storedRec.Body.Bytes(), &stored); err != nil {
		t.Fatal(err)
	}
	if stored.Equipment[domain.SlotHead].ItemID != 208424 {
		t.Fatalf("failed save mutated build: %+v", stored.Equipment)
	}

	// Current cannot be deleted.
	rec = del(t, srv, "/api/builds/"+current.ID)
	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), "cannot_delete_current") {
		t.Fatalf("delete current %d: %s", rec.Code, rec.Body.String())
	}

	// A plan can be created and deleted.
	planRec := post(t, srv, "/api/characters/"+ch.ID+"/builds", `{"name":"Pre-raid"}`)
	if planRec.Code != http.StatusCreated {
		t.Fatalf("plan %d: %s", planRec.Code, planRec.Body.String())
	}
	var plan domain.Build
	if err := json.Unmarshal(planRec.Body.Bytes(), &plan); err != nil {
		t.Fatal(err)
	}
	if rec := del(t, srv, "/api/builds/"+plan.ID); rec.Code != http.StatusNoContent {
		t.Fatalf("delete plan %d: %s", rec.Code, rec.Body.String())
	}
	if builds := listBuildsFor(t, srv, ch.ID); len(builds) != 1 {
		t.Fatalf("plans after delete = %d", len(builds))
	}
}
