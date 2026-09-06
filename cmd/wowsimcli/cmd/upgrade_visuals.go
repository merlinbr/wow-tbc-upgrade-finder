package cmd

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// wowheadItemXMLURL is the TBC item source of display IDs. It is a
	// server-side fetch: browsers cannot consume it cross-origin, and the
	// per-item XML payloads are not part of the viewer's asset contract.
	wowheadItemXMLURL = "https://www.wowhead.com/tbc/item=%d&xml"
	// zamAssetBase is the ZAM TBC model asset tree. Assets are fetched
	// through this server because the browser cannot load them directly
	// from wow.zamimg.com (no CORS) — the same transport shape the reference
	// site uses, gated behind the explicit --enable-3d authorization flag.
	zamAssetBase             = "https://wow.zamimg.com/modelviewer/tbc/"
	maxItemXMLBytes          = 1 << 17 // 128 KiB
	maxAssetBytes            = 64 << 20
	maxVisualResolverEntries = 4096
	maxVisualResolveItems    = 32
)

// visualItem is the resolved preview input for one equipped item.
type visualItem struct {
	DisplayID int32  `json:"displayId"`
	ZamSlot   int32  `json:"zamSlot"`
	Name      string `json:"name"`
}

type visualResolver struct {
	client    *http.Client
	assetBase string
	xmlURL    string

	mu    sync.Mutex
	cache map[int32]*visualItem
}

func newVisualResolver() *visualResolver {
	return &visualResolver{
		client:    &http.Client{Timeout: 15 * time.Second},
		assetBase: zamAssetBase,
		xmlURL:    wowheadItemXMLURL,
		cache:     make(map[int32]*visualItem),
	}
}

// resolve returns the item's display ID and ZAM slot, or ok=false when the
// item has no visual representation (relics, no display ID, unresolved meta).
func (r *visualResolver) resolve(ctx context.Context, id int32) (*visualItem, bool, error) {
	r.mu.Lock()
	it := r.cache[id]
	r.mu.Unlock()
	if it != nil {
		return it, true, nil
	}

	xmlBody, err := r.fetchItemXML(ctx, id)
	if err != nil {
		return nil, false, err
	}
	displayID, inventorySlot, name, err := parseWowheadItemXML(xmlBody)
	if err != nil {
		return nil, false, err
	}
	if displayID == 0 || inventorySlot == 0 || inventorySlot == 28 {
		// No display (relic/projectile) or no usable inventory slot: not a
		// visible model part; the caller reports it as not applicable.
		return nil, false, nil
	}
	zamSlot, ok, err := r.normalizeSlot(ctx, inventorySlot, displayID)
	if err != nil || !ok {
		return nil, false, err
	}

	it = &visualItem{DisplayID: displayID, ZamSlot: zamSlot, Name: name}
	r.mu.Lock()
	if len(r.cache) >= maxVisualResolverEntries {
		r.cache = make(map[int32]*visualItem) // ponytail: bounded cache, clear-all; add LRU if imports grow
	}
	r.cache[id] = it
	r.mu.Unlock()
	return it, true, nil
}

func (r *visualResolver) fetchItemXML(ctx context.Context, id int32) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(r.xmlURL, id), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/xml, text/xml")
	req.Header.Set("User-Agent", "wow-upgrade-agent/1.0")
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wowhead item XML for %d: HTTP %d", id, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxItemXMLBytes))
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("wowhead item XML for %d: empty body", id)
	}
	return body, nil
}

// parseWowheadItemXML extracts displayId, inventorySlot and name from a
// wowhead /tbc/item XML response.
func parseWowheadItemXML(data []byte) (displayID, inventorySlot int32, name string, err error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var seenDisplay, seenSlot, seenName bool
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return 0, 0, "", err
		}
		switch el := tok.(type) {
		case xml.StartElement:
			switch el.Name.Local {
			case "icon":
				if !seenDisplay {
					if v, ok := attrInt(el, "displayId"); ok {
						displayID, seenDisplay = v, true
					}
				}
			case "inventorySlot":
				if !seenSlot {
					if v, ok := attrInt(el, "id"); ok {
						inventorySlot, seenSlot = v, true
					}
				}
			case "name":
				if !seenName {
					if text, ok := readElementText(dec); ok {
						name, seenName = strings.TrimSpace(text), true
					}
				}
			}
		}
		if seenDisplay && seenSlot && seenName {
			break
		}
	}
	if !seenDisplay || !seenSlot {
		return 0, 0, "", errors.New("wowhead item XML: displayId or inventorySlot missing")
	}
	return displayID, inventorySlot, name, nil
}

func attrInt(el xml.StartElement, name string) (int32, bool) {
	for _, a := range el.Attr {
		if a.Name.Local == name {
			v, err := strconv.ParseInt(a.Value, 10, 32)
			if err != nil {
				return 0, false
			}
			return int32(v), true
		}
	}
	return 0, false
}

func readElementText(dec *xml.Decoder) (string, bool) {
	var b strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil || errors.Is(err, io.EOF) {
			return "", false
		}
		if end, ok := tok.(xml.EndElement); ok && end.Name.Local == "name" {
			return b.String(), true
		}
		if ch, ok := tok.(xml.CharData); ok {
			b.Write(ch)
		}
	}
}

// normalizeSlot decides the ZAM viewer slot. Chest items are ambiguous in the
// Wowhead XML (chest 5 vs robe-cut 20): the ZAM meta tree is authoritative,
// so the resolver probes meta/armor/5 then meta/armor/20 before deciding.
// Weapon slots keep their XML value; the caller overrides them by equipped
// position (main hand 21, off hand 22, ranged 15).
func (r *visualResolver) normalizeSlot(ctx context.Context, inventorySlot, displayID int32) (int32, bool, error) {
	if inventorySlot == 5 {
		for _, candidate := range []int32{5, 20} {
			ok, err := r.metaExists(ctx, fmt.Sprintf("meta/armor/%d/%d.json", candidate, displayID))
			if err != nil {
				return 0, false, err
			}
			if ok {
				return candidate, true, nil
			}
		}
		return 0, false, nil
	}
	return inventorySlot, true, nil
}

func (r *visualResolver) metaExists(ctx context.Context, rel string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.assetBase+rel, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", "wow-upgrade-agent/1.0")
	resp, err := r.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("ZAM meta %s: HTTP %d", rel, resp.StatusCode)
	}
}

// handleResolveVisualItems resolves a batch of imported item IDs into viewer
// inputs. Unresolvable items are reported in "missing" and treated as no
// attachment by the caller; ranking/import never depends on this endpoint.
func (s *upgradeServer) handleResolveVisualItems(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Items []int32 `json:"items"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		return
	}
	if len(body.Items) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"items": map[string]visualItem{}, "missing": []int32{}})
		return
	}
	if len(body.Items) > maxVisualResolveItems {
		writeError(w, http.StatusBadRequest, "too_many_items", "at most 32 items per request")
		return
	}

	resolved := make(map[string]visualItem, len(body.Items))
	missing := make([]int32, 0, len(body.Items))
	for _, id := range body.Items {
		it, ok, err := s.visuals.resolve(r.Context(), id)
		if err != nil || !ok {
			missing = append(missing, id)
			continue
		}
		resolved[strconv.FormatInt(int64(id), 10)] = *it
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": resolved, "missing": missing})
}

// handleZamAsset proxies one ZAM TBC asset through this loopback origin so
// the browser can load meshes, skins and textures (no CORS upstream).
// Only the fixed modelviewer/tbc asset tree is reachable; no arbitrary URLs.
func (s *upgradeServer) handleZamAsset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		methodNotAllowed(w, http.MethodGet+", "+http.MethodHead)
		return
	}
	assetPath := r.PathValue("path")
	if assetPath == "" || len(assetPath) > 160 ||
		strings.Contains(assetPath, "..") || strings.Contains(assetPath, "\\") ||
		strings.HasPrefix(assetPath, "/") || path.Clean(assetPath) != assetPath {
		writeError(w, http.StatusBadRequest, "invalid_asset_path", "invalid asset path")
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, zamAssetBase+assetPath, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to build upstream request")
		return
	}
	req.Header.Set("User-Agent", "wow-upgrade-agent/1.0")
	resp, err := s.visuals.client.Do(req)
	if err != nil {
		writeError(w, http.StatusBadGateway, "upstream_unavailable", "model asset upstream unavailable")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		writeError(w, http.StatusNotFound, "asset_not_found", "model asset not found")
		return
	}
	if resp.ContentLength > maxAssetBytes {
		writeError(w, http.StatusBadGateway, "asset_too_large", "model asset exceeds size limit")
		return
	}

	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	if _, err := io.Copy(w, io.LimitReader(resp.Body, maxAssetBytes)); err != nil {
		_ = w // client disconnect; nothing further to report
	}
}
