package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const sampleItemXML = `<?xml version="1.0" encoding="UTF-8"?><wowhead><item id="32461"><name><![CDATA[Furious Gizmatic Goggles]]></name><level>127</level><quality id="4">Epic</quality><class id="4"><![CDATA[Armor]]></class><subclass id="4"><![CDATA[Plate Armor]]></subclass><icon displayId="45779">inv_gizmo_newgoggles</icon><inventorySlot id="1">Head</inventorySlot><htmlTooltip><![CDATA[<table>ignored</table>]]></htmlTooltip></item></wowhead>`

func TestParseWowheadItemXML(t *testing.T) {
	displayID, slot, name, err := parseWowheadItemXML([]byte(sampleItemXML))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if displayID != 45779 || slot != 1 || name != "Furious Gizmatic Goggles" {
		t.Fatalf("got %d %d %q", displayID, slot, name)
	}
}

func TestParseWowheadItemXMLMissing(t *testing.T) {
	if _, _, _, err := parseWowheadItemXML([]byte(`<wowhead><item id="1"></item></wowhead>`)); err == nil {
		t.Fatal("expected error for missing displayId/inventorySlot")
	}
}

func newResolverWithStubs(t *testing.T, mux *http.ServeMux) *visualResolver {
	t.Helper()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	r := newVisualResolver()
	r.xmlURL = server.URL + "/xml/item=%d"
	r.assetBase = server.URL + "/zam/"
	return r
}

func TestVisualResolverChestProbePrefersExistingMeta(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/xml/item=30129", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<wowhead><item id="30129"><name><![CDATA[Crystalforge Breastplate]]></name><icon displayId="42306">inv</icon><inventorySlot id="5">Chest</inventorySlot></item></wowhead>`)
	})
	mux.HandleFunc("/zam/meta/armor/5/42306.json", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/zam/meta/armor/20/42306.json", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) })

	r := newResolverWithStubs(t, mux)
	it, ok, err := r.resolve(context.Background(), 30129)
	if err != nil || !ok {
		t.Fatalf("resolve: ok=%v err=%v", ok, err)
	}
	if it.ZamSlot != 5 || it.DisplayID != 42306 {
		t.Fatalf("got slot %d display %d", it.ZamSlot, it.DisplayID)
	}
}

func TestVisualResolverChestProbeRobe(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/xml/item=29077", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<wowhead><item id="29077"><name><![CDATA[Vestments of the Aldor]]></name><icon displayId="40468">inv</icon><inventorySlot id="5">Chest</inventorySlot></item></wowhead>`)
	})
	mux.HandleFunc("/zam/meta/armor/5/40468.json", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) })
	mux.HandleFunc("/zam/meta/armor/20/40468.json", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	r := newResolverWithStubs(t, mux)
	it, ok, err := r.resolve(context.Background(), 29077)
	if err != nil || !ok {
		t.Fatalf("resolve: ok=%v err=%v", ok, err)
	}
	if it.ZamSlot != 20 {
		t.Fatalf("robe chest got slot %d, want 20", it.ZamSlot)
	}
}

func TestVisualResolverSkipsNoDisplayItems(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/xml/item=27484", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<wowhead><item id="27484"><name><![CDATA[Libram of Avengement]]></name><icon displayId="0">inv</icon><inventorySlot id="28">Relic</inventorySlot></item></wowhead>`)
	})
	r := newResolverWithStubs(t, mux)
	if _, ok, err := r.resolve(context.Background(), 27484); err != nil || ok {
		t.Fatalf("relic: ok=%v err=%v, want skipped", ok, err)
	}
}

func TestVisualResolverCache(t *testing.T) {
	calls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/xml/item=32461", func(w http.ResponseWriter, r *http.Request) {
		calls++
		fmt.Fprint(w, sampleItemXML)
	})
	mux.HandleFunc("/zam/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	r := newResolverWithStubs(t, mux)
	if _, _, err := r.resolve(context.Background(), 32461); err != nil {
		t.Fatal(err)
	}
	if _, _, err := r.resolve(context.Background(), 32461); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("wowhead XML fetched %d times, want 1 (cached)", calls)
	}
}

func TestVisualResolverWeaponSlotPassesThrough(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/xml/item=28430", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<wowhead><item id="28430"><name><![CDATA[Lionheart Executioner]]></name><icon displayId="39571">inv</icon><inventorySlot id="17">Two-Hand</inventorySlot></item></wowhead>`)
	})
	r := newResolverWithStubs(t, mux)
	// Even with a 404 ZAM tree, non-chest items must not be probed.
	it, ok, err := r.resolve(context.Background(), 28430)
	if err != nil || !ok {
		t.Fatalf("resolve: ok=%v err=%v", ok, err)
	}
	if it.ZamSlot != 17 {
		t.Fatalf("got slot %d, want 17", it.ZamSlot)
	}
}

func TestVisualsRoutesGatedByFlag(t *testing.T) {
	s := newUpgradeServer("test")
	server := httptest.NewServer(s.routes())
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/api/visuals/resolve")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("disabled visuals route: status %d, want 404", resp.StatusCode)
	}
}

func TestVisualsResolveRoute(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/xml/item=32461", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, sampleItemXML) })
	mux.HandleFunc("/zam/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	visuals := newResolverWithStubs(t, mux)

	s := newUpgradeServer("test")
	s.visualsEnabled = true
	s.visuals = visuals
	server := httptest.NewServer(s.routes())
	t.Cleanup(server.Close)

	resp, err := http.Post(server.URL+"/api/visuals/resolve", "application/json", strings.NewReader(`{"items":[32461,27484]}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	var payload struct {
		Items   map[string]visualItem `json:"items"`
		Missing []int32               `json:"missing"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if got := payload.Items["32461"]; got.DisplayID != 45779 || got.ZamSlot != 1 {
		t.Fatalf("bad item: %+v", got)
	}
	if len(payload.Missing) != 1 || payload.Missing[0] != 27484 {
		t.Fatalf("missing: %v", payload.Missing)
	}
}

func TestZamAssetProxyPathValidation(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/zam/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})
	visuals := newResolverWithStubs(t, mux)

	s := newUpgradeServer("test")
	s.visualsEnabled = true
	s.visuals = visuals
	server := httptest.NewServer(s.routes())
	t.Cleanup(server.Close)

	valid := "meta/character/2.json"
	resp, err := http.Get(server.URL + "/visuals/zam/modelviewer/tbc/" + valid)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("valid asset: status %d", resp.StatusCode)
	}

	for _, bad := range []string{"../secret", "meta/..%2Fetc", "a/b//c", "//double", `a\b`} {
		resp, err := http.Get(server.URL + "/visuals/zam/modelviewer/tbc/" + bad)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		// Traversal may be canonicalized by the URL layer (404) or rejected by
		// the handler (400); both must reject, never proxy.
		if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
			t.Fatalf("invalid path %q: status %d, want 400 or 404", bad, resp.StatusCode)
		}
	}
}
