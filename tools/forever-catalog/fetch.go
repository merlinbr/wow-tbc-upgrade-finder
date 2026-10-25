package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	pinnedCommit       = "0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4"
	pinnedDBHash       = "e78be48cbc65e3a82ad3baa15ccb5f8fad524ccc35d18887b9be1ce241194de7"
	pinnedLeftoverHash = "6de2991a0f23e3a565aba0f06af5f740fc41925f552c074ba196b6801e30b761"
	pinnedBaseURL      = "https://raw.githubusercontent.com/wowsims/forever/" + pinnedCommit
)

// RawPayload is the downloaded upstream data, kept as bytes so the pinned
// hashes are computed over exactly what the server served.
type RawPayload struct {
	DB           []byte
	LeftoverDB   []byte
	DBHash       string
	LeftoverHash string
}

// Revision is the catalog revision derived from the raw payload hashes.
func (r RawPayload) Revision() string {
	if r.DBHash == "" || r.LeftoverHash == "" {
		return ""
	}
	return fmt.Sprintf("wowsims/forever@%s:%s:%s", pinnedCommit[:7], r.DBHash, r.LeftoverHash)
}

// FetchPinned downloads the pinned catalogs and verifies both hashes.
func FetchPinned(ctx context.Context, baseURL string) (RawPayload, error) {
	base := strings.TrimRight(baseURL, "/")
	if base == "" {
		base = pinnedBaseURL
	}
	db, err := fetchBytes(ctx, base+"/assets/database/db.json")
	if err != nil {
		return RawPayload{}, fmt.Errorf("fetch db.json: %w", err)
	}
	leftover, err := fetchBytes(ctx, base+"/assets/database/leftover_db.json")
	if err != nil {
		return RawPayload{}, fmt.Errorf("fetch leftover_db.json: %w", err)
	}
	dbHash := sha256Hex(db)
	leftoverHash := sha256Hex(leftover)
	if dbHash != pinnedDBHash {
		return RawPayload{}, fmt.Errorf("db.json sha256 = %s, want %s", dbHash, pinnedDBHash)
	}
	if leftoverHash != pinnedLeftoverHash {
		return RawPayload{}, fmt.Errorf("leftover_db.json sha256 = %s, want %s", leftoverHash, pinnedLeftoverHash)
	}
	return RawPayload{
		DB:           db,
		LeftoverDB:   leftover,
		DBHash:       dbHash,
		LeftoverHash: leftoverHash,
	}, nil
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

var httpClient = &http.Client{Timeout: 120 * time.Second}

func fetchBytes(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
