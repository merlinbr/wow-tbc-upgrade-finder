package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestItemsSearch(t *testing.T) {
	srv := newTestServer(t)
	rec := get(t, srv, "/api/items?class=paladin&slot=head&limit=1")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Items []struct {
			ID         int    `json:"id"`
			Name       string `json:"name"`
			Compatible bool   `json:"compatible"`
			Reason     string `json:"reason"`
		} `json:"items"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Items) != 1 {
		t.Fatalf("items = %d", len(body.Items))
	}
	if body.Total < 2 {
		t.Fatalf("total = %d", body.Total)
	}
	if !body.Items[0].Compatible {
		t.Fatalf("paladin head incompatible: %s", body.Items[0].Reason)
	}

	if rec := get(t, srv, "/api/items?limit=nope"); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad limit status = %d", rec.Code)
	}
}

func TestItemDetailUnknownReturns404(t *testing.T) {
	srv := newTestServer(t)
	rec := get(t, srv, "/api/items/999999999")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
	var e struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
		t.Fatal(err)
	}
	if e.Error.Code != "not_found" {
		t.Fatalf("code %q", e.Error.Code)
	}
}

func TestItemDetailCarriesCompatibility(t *testing.T) {
	srv := newTestServer(t)
	rec := get(t, srv, "/api/items/23019?class=mage&slot=head")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		ID         int    `json:"id"`
		Compatible bool   `json:"compatible"`
		Reason     string `json:"reason"`
		LevelReq   string `json:"levelRequirement"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.ID != 23019 || body.Compatible || body.Reason != "incompatible_armor" {
		t.Fatalf("detail = %+v", body)
	}
	if body.LevelReq != "unknown" {
		t.Fatalf("level requirement = %q", body.LevelReq)
	}
}

func TestEnchantsSearch(t *testing.T) {
	srv := newTestServer(t)
	rec := get(t, srv, "/api/enchants?slot=neck&class=paladin")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Enchants []struct {
			EffectID int    `json:"effectId"`
			Name     string `json:"name"`
		} `json:"enchants"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Enchants) == 0 {
		t.Fatal("no neck enchants")
	}
	for _, enchant := range body.Enchants {
		if enchant.EffectID == 0 {
			t.Fatal("enchant without effect id")
		}
	}
}

func TestTalentsRequiresKnownClass(t *testing.T) {
	srv := newTestServer(t)
	if rec := get(t, srv, "/api/talents?class=paladin"); rec.Code != http.StatusOK {
		t.Fatalf("paladin %d: %s", rec.Code, rec.Body.String())
	}
	rec := get(t, srv, "/api/talents?class=bard")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bard %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "unsupported_class") {
		t.Fatalf("bard body = %s", rec.Body.String())
	}
}
