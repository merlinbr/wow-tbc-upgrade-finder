package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/wowsims/tbc/internal/forever/domain"
)

// classFiles are the nine pinned talent documents, in the upstream file names.
var classFiles = []domain.Class{
	"druid", "hunter", "mage", "paladin", "priest", "rogue", "shaman", "warlock", "warrior",
}

func main() {
	verify := flag.Bool("verify", false, "assert the pinned source counts before writing")
	baseURL := flag.String("base-url", pinnedBaseURL, "upstream base URL")
	outDir := flag.String("out", filepath.Join("assets", "forever"), "output directory")
	flag.Parse()

	if err := run(context.Background(), *baseURL, *outDir, *verify); err != nil {
		fmt.Fprintln(os.Stderr, "forever-catalog:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, baseURL, outDir string, verify bool) error {
	protoText, err := fetchBytes(ctx, baseURL+"/proto/common.proto")
	if err != nil {
		return fmt.Errorf("fetch proto/common.proto: %w", err)
	}
	enums, err := ParseEnums(string(protoText))
	if err != nil {
		return err
	}
	raw, err := FetchPinned(ctx, baseURL)
	if err != nil {
		return err
	}
	counts, err := CountRaw(raw)
	if err != nil {
		return err
	}
	catalog, err := ProjectItems(raw, enums)
	if err != nil {
		return err
	}

	spellIcons, err := ParseSpellIcons(raw.DB)
	if err != nil {
		return err
	}
	rawTalents := RawTalents{
		Classes:    make(map[domain.Class][]byte, len(classFiles)),
		SpellIcons: spellIcons,
	}
	for _, class := range classFiles {
		data, err := fetchBytes(
			ctx, baseURL+"/ui/sim/talents/trees/"+string(class)+".json",
		)
		if err != nil {
			return fmt.Errorf("fetch %s talents: %w", class, err)
		}
		rawTalents.Classes[class] = data
	}
	trees, err := ProjectTalents(rawTalents, enums)
	if err != nil {
		return err
	}
	specs := SpecsFromTrees(trees)
	capsText, err := fetchBytes(
		ctx, baseURL+"/ui/sim/player/classes/capabilities_auto_gen.ts",
	)
	if err != nil {
		return fmt.Errorf("fetch capabilities: %w", err)
	}
	capabilities, err := ParseCapabilities(string(capsText), enums, specs)
	if err != nil {
		return err
	}

	talentCount := CountTalents(trees)
	classCount := len(trees)
	manifest := BuildManifest(raw, catalog, counts, talentCount, classCount)

	fmt.Printf(
		"items=%d itemsEligible=%d leftoverItems=%d enchants=%d leftoverEnchants=%d "+
			"talents=%d classes=%d suffixCount=%d\n",
		counts.ItemsTotal, len(catalog.Items), counts.ItemsLeftover, counts.Enchants,
		counts.LeftoverEnchants, talentCount, classCount, counts.Suffixes,
	)

	if verify {
		if counts.ItemsTotal != 8599 ||
			counts.ItemsLeftover != 1186 ||
			counts.Enchants != 246 ||
			counts.LeftoverEnchants != 20 ||
			talentCount != 465 ||
			classCount != 9 {
			return fmt.Errorf("pinned counts changed: got %+v talents=%d classes=%d", counts, talentCount, classCount)
		}
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	if err := writeGzipJSON(filepath.Join(outDir, "catalog.json.gz"), catalog); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(outDir, "capabilities.json"), capabilities); err != nil {
		return err
	}
	talentsDir := filepath.Join(outDir, "talents")
	if err := os.RemoveAll(talentsDir); err != nil {
		return err
	}
	if err := os.MkdirAll(talentsDir, 0o755); err != nil {
		return err
	}
	classes := make([]string, 0, len(trees))
	for class := range trees {
		classes = append(classes, class)
	}
	sort.Strings(classes)
	for _, class := range classes {
		if err := writeJSON(filepath.Join(talentsDir, class+".json"), trees[class]); err != nil {
			return err
		}
	}
	return writeJSON(filepath.Join(outDir, "manifest.json"), manifest)
}

func writeGzipJSON(path string, value any) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := gzip.NewWriter(file)
	encoder := json.NewEncoder(writer)
	if err := encoder.Encode(value); err != nil {
		return err
	}
	return writer.Close()
}
