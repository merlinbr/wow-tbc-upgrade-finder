package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

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

	fmt.Printf(
		"items=%d itemsEligible=%d leftoverItems=%d enchants=%d leftoverEnchants=%d suffixCount=%d\n",
		counts.ItemsTotal, len(catalog.Items), counts.ItemsLeftover, counts.Enchants,
		counts.LeftoverEnchants, counts.Suffixes,
	)

	if verify {
		expected := RawCounts{
			ItemsTotal:       8599,
			ItemsLeftover:    1186,
			Enchants:         246,
			LeftoverEnchants: 20,
		}
		if counts.ItemsTotal != expected.ItemsTotal ||
			counts.ItemsLeftover != expected.ItemsLeftover ||
			counts.Enchants != expected.Enchants ||
			counts.LeftoverEnchants != expected.LeftoverEnchants {
			return fmt.Errorf("pinned counts changed: got %+v, want at least %+v", counts, expected)
		}
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	return writeGzipJSON(filepath.Join(outDir, "catalog.json.gz"), catalog)
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
