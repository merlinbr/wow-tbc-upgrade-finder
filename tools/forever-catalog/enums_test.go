package main

import (
	"os"
	"testing"
)

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func TestParseEnums(t *testing.T) {
	enums, err := ParseEnums(string(mustRead(t, "testdata/proto.txt")))
	if err != nil {
		t.Fatal(err)
	}
	if got := enums.StatName(27); got != "armor" {
		t.Fatalf("stat 27 = %q", got)
	}
	if got := enums.SlotName(0); got != "head" {
		t.Fatalf("slot 0 = %q", got)
	}
	if got := enums.ItemTypeName(7); got != "hands" {
		t.Fatalf("type 7 = %q", got)
	}
	if got := enums.WeaponName(9); got != "sword" {
		t.Fatalf("weapon 9 = %q", got)
	}
}
