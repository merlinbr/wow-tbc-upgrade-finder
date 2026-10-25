// Package server exposes the Forever armory JSON API and the embedded UI.
package server

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http"
	"sort"
	"time"

	"github.com/wowsims/tbc/internal/forever/catalog"
	"github.com/wowsims/tbc/internal/forever/domain"
	"github.com/wowsims/tbc/internal/forever/service"
)

type api struct {
	catalog *catalog.Catalog
	service *service.Service
	ui      fs.FS
}

// New returns the HTTP handler. The service may be nil while the API is being
// scaffolded. The optional UI filesystem is the built Svelte bundle.
func New(cat *catalog.Catalog, svc *service.Service, ui ...fs.FS) http.Handler {
	handler := &api{catalog: cat, service: svc}
	if len(ui) > 0 {
		handler.ui = ui[0]
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/meta", handler.handleMeta)
	mux.HandleFunc("GET /api/items", handler.handleItems)
	mux.HandleFunc("GET /api/items/{id}", handler.handleItemDetail)
	mux.HandleFunc("GET /api/enchants", handler.handleEnchants)
	mux.HandleFunc("GET /api/talents", handler.handleTalents)
	mux.HandleFunc("GET /assets/{path...}", handler.handleAsset)
	mux.HandleFunc("GET /", handler.handleIndex)
	return mux
}

type metaResponse struct {
	CatalogRevision string                       `json:"catalogRevision"`
	SourceCommit    string                       `json:"sourceCommit"`
	Counts          domain.ManifestCounts        `json:"counts"`
	Classes         []string                     `json:"classes"`
	Specs           map[string][]domain.SpecInfo `json:"specs"`
	Races           []string                     `json:"races"`
	ClassRaces      map[string][]string          `json:"classRaces"`
	DualWield       map[string]bool              `json:"dualWield"`
	Professions     []string                     `json:"professions"`
}

func (a *api) handleMeta(w http.ResponseWriter, _ *http.Request) {
	caps := a.catalog.Capabilities

	classes := make([]string, 0, len(caps.Specs))
	for class := range caps.Specs {
		classes = append(classes, class)
	}
	sort.Strings(classes)

	races := make([]string, 0)
	seenRaces := make(map[string]bool)
	classRaces := make(map[string][]string, len(caps.Races))
	for class, classList := range caps.Races {
		classRaces[class] = classList
		for _, race := range classList {
			if !seenRaces[race] {
				seenRaces[race] = true
				races = append(races, race)
			}
		}
	}
	sort.Strings(races)

	dualWield := make(map[string]bool, len(caps.DualWield))
	for spec, allowed := range caps.DualWield {
		dualWield[spec] = allowed
	}

	writeJSON(w, http.StatusOK, metaResponse{
		CatalogRevision: a.catalog.Manifest.CatalogRevision,
		SourceCommit:    a.catalog.Manifest.SourceCommit,
		Counts:          a.catalog.Manifest.Counts,
		Classes:         classes,
		Specs:           caps.Specs,
		Races:           races,
		ClassRaces:      classRaces,
		DualWield:       dualWield,
		Professions:     caps.Professions,
	})
}

func (a *api) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}
	if a.ui == nil {
		http.Error(w, "Forever armory web UI is not built", http.StatusServiceUnavailable)
		return
	}
	index, err := fs.ReadFile(a.ui, "index.html")
	if err != nil {
		http.Error(w, "Forever armory web UI is not built", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(index)
}

func (a *api) handleAsset(w http.ResponseWriter, r *http.Request) {
	if a.ui == nil {
		writeError(w, http.StatusNotFound, "not_found", "asset not found")
		return
	}
	path := r.PathValue("path")
	if path == "" || !fs.ValidPath(path) {
		writeError(w, http.StatusNotFound, "not_found", "asset not found")
		return
	}
	data, err := fs.ReadFile(a.ui, "assets/"+path)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "asset not found")
		return
	}
	http.ServeContent(w, r, path, time.Time{}, bytes.NewReader(data))
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorEnvelope{Error: errorBody{Code: code, Message: message}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
