package cmd

import (
	"testing"
)

func TestRankUpgradesCommandIsRegistered(t *testing.T) {
	command, _, err := newRootCommand("test").Find([]string{"rank-upgrades"})
	if err != nil {
		t.Fatal(err)
	}
	if command.Name() != "rank-upgrades" {
		t.Fatalf("command = %q, want rank-upgrades", command.Name())
	}
}
