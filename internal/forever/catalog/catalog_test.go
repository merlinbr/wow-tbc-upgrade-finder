package catalog

import (
	"testing"
	"testing/fstest"
)

func TestLoadEmbeddedCatalog(t *testing.T) {
	cat, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cat.Manifest.CatalogRevision == "" {
		t.Fatal("empty revision")
	}
	if len(cat.Items) < 1000 {
		t.Fatalf("items = %d", len(cat.Items))
	}
	if cat.Talents["paladin"] == nil || len(cat.Talents["paladin"].Trees) != 3 {
		t.Fatal("paladin trees")
	}
}

func TestLoadRejectsCorruptCatalog(t *testing.T) {
	_, err := LoadFrom(fstest.MapFS{
		"catalog.json.gz": &fstest.MapFile{Data: []byte("not gzip")},
	})
	if err == nil {
		t.Fatal("corrupt catalog accepted")
	}
}
