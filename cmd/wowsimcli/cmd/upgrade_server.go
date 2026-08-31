package cmd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/browser"
	"github.com/wowsims/tbc/assets/database"
	"github.com/wowsims/tbc/cmd/wowsimcli/cmd/upgrades"
)

// browserOpen opens the resolved local URL in the user's browser.
func browserOpen(url string) error {
	return browser.OpenURL(url)
}

const maxBodyBytes = 1 << 20 // 1 MiB

const (
	jobStatusQueued    = "queued"
	jobStatusRunning   = "running"
	jobStatusCompleted = "completed"
	jobStatusFailed    = "failed"
	jobStatusCanceled  = "canceled"
)

// Ranker is the seam the HTTP layer consumes; the real implementation is
// upgrades.Service.
type Ranker interface {
	RankUpgrades(ctx context.Context, request upgrades.RankRequest, progress func(upgrades.Progress)) (*upgrades.UpgradeReport, error)
}

type rankJob struct {
	id        string
	status    string
	progress  upgrades.Progress
	character upgrades.CharacterSummary
	report    *upgrades.UpgradeReport
	err       *apiError
	cancel    context.CancelFunc
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type upgradeServer struct {
	version string
	ranker  Ranker
	catalog *upgrades.Catalog

	mu   sync.Mutex
	jobs map[string]*rankJob
}

func newUpgradeServer(version string) *upgradeServer {
	return &upgradeServer{
		version: version,
		catalog: upgrades.NewCatalog(database.Load()),
		jobs:    make(map[string]*rankJob),
	}
}

func (s *upgradeServer) Serve(ctx context.Context, addr string, openBrowser bool) error {
	if s.ranker == nil {
		s.ranker = upgrades.NewRankService(upgrades.NewRealSimulator(), s.catalog)
	}

	// Force loopback regardless of the requested host.
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		host, port = "127.0.0.1", "0"
	}
	if host == "" || host == "127.0.0.1" {
		host = "127.0.0.1"
	}

	listener, err := net.Listen("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return fmt.Errorf("failed to bind listener: %w", err)
	}

	resolvedURL := fmt.Sprintf("http://%s/", listener.Addr().String())
	fmt.Printf("Upgrade finder running at %s\n", resolvedURL)

	server := &http.Server{
		Handler:           s.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	if openBrowser {
		go func() { _ = browserOpen(resolvedURL) }()
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Serve(listener)
	}()

	select {
	case <-ctx.Done():
		s.cancelAllJobs()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

func (s *upgradeServer) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /assets/{path...}", s.handleAsset)
	mux.HandleFunc("/api/import", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			methodNotAllowed(w, http.MethodPost)
			return
		}
		s.handleImport(w, r)
	})
	mux.HandleFunc("/api/jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			methodNotAllowed(w, http.MethodPost)
			return
		}
		s.handleCreateJob(w, r)
	})
	mux.HandleFunc("/api/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handleGetJob(w, r)
		case http.MethodDelete:
			s.handleCancelJob(w, r)
		default:
			methodNotAllowed(w, http.MethodGet+", "+http.MethodDelete)
		}
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not_found", "unknown path")
	})
	return mux
}

func methodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]apiError{"error": {Code: code, Message: message}})
}
// decodeJSON rejects unknown fields, oversized bodies, and trailing content.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	body := http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "body_too_large", "request body exceeds 1 MiB")
		} else if w.Header().Get("Content-Type") == "" {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body: "+err.Error())
		}
		return err
	}
	if dec.More() {
		writeError(w, http.StatusBadRequest, "invalid_request", "multiple JSON values in body")
		return errors.New("multiple JSON values in body")
	}
	return nil
}

type importRequestBody struct {
	Link string `json:"link"`
}

type defaultsResponse struct {
	MaxPhase               int32    `json:"maxPhase"`
	SourceKinds            []string `json:"sourceKinds"`
	IncludeUnknown         bool     `json:"includeUnknown"`
	ScreeningIterations    int32    `json:"screeningIterations"`
	ConfirmationIterations int32    `json:"confirmationIterations"`
}

type importResponse struct {
	Character         upgrades.CharacterSummary `json:"character"`
	SettingsDigest    string                    `json:"settingsDigest"`
	SimulatorRevision string                    `json:"simulatorRevision"`
	DatabaseRevision  string                    `json:"databaseRevision"`
	Defaults          defaultsResponse          `json:"defaults"`
	Gear              []upgrades.GearSlotData    `json:"gear"`
	Stats             map[string]float64         `json:"stats"`
	DerivedStats      map[string]float64         `json:"derivedStats"`
}

func (s *upgradeServer) handleImport(w http.ResponseWriter, r *http.Request) {
	var body importRequestBody
	if err := decodeJSON(w, r, &body); err != nil {
		return
	}

	imported, err := upgrades.Import(body.Link)
	if err != nil {
		writeValidationError(w, err)
		return
	}
	armory, err := upgrades.EnrichArmory(imported, s.catalog)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	maxPhase := imported.Character.Phase
	if maxPhase < 1 {
		maxPhase = 5
	}

	writeJSON(w, http.StatusOK, importResponse{
		Character:         imported.Character,
		SettingsDigest:    imported.SettingsDigest,
		SimulatorRevision: upgrades.SimulatorRevision,
		DatabaseRevision:  upgrades.DatabaseRevision,
		Defaults: defaultsResponse{
			MaxPhase:               maxPhase,
			SourceKinds:            []string{},
			IncludeUnknown:         false,
			ScreeningIterations:    300,
			ConfirmationIterations: 1000,
		},
		Gear:         armory.Gear,
		Stats:        armory.Stats,
		DerivedStats: armory.DerivedStats,
	})
}

func writeValidationError(w http.ResponseWriter, err error) {
	var valErr *upgrades.ValidationError
	if errors.As(err, &valErr) {
		writeError(w, http.StatusBadRequest, valErr.Code, valErr.Message)
		return
	}
	writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
}

type RankJobInput struct {
	Link    string                     `json:"link"`
	Filters upgrades.ContentFilters    `json:"filters"`
	Policy  upgrades.ItemPolicy        `json:"policy"`
	Options upgrades.SimulationOptions `json:"options"`
}

type jobResponse struct {
	ID        string                     `json:"id"`
	Status    string                     `json:"status"`
	Progress  upgrades.Progress          `json:"progress"`
	Character *upgrades.CharacterSummary `json:"character,omitempty"`
	Error     *apiError                  `json:"error,omitempty"`
	Report    *upgrades.UpgradeReport    `json:"report,omitempty"`
}

func (s *upgradeServer) handleCreateJob(w http.ResponseWriter, r *http.Request) {
	var input RankJobInput
	if err := decodeJSON(w, r, &input); err != nil {
		return
	}

	// Validate the link synchronously; never create a job for an invalid link.
	imported, err := upgrades.Import(input.Link)
	if err != nil {
		writeValidationError(w, err)
		return
	}

	options := input.Options
	if options.ScreeningIterations == 0 && options.ConfirmationIterations == 0 {
		options = defaultSimulationOptions()
	}
	if options.ScreeningIterations <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_options", "screeningIterations must be greater than 0")
		return
	}
	if options.ConfirmationIterations < options.ScreeningIterations {
		writeError(w, http.StatusBadRequest, "invalid_options", "confirmationIterations must be greater than or equal to screeningIterations")
		return
	}

	jobID, err := newJobID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to generate job id")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	job := &rankJob{
		id:        jobID,
		status:    jobStatusQueued,
		character: imported.Character,
		cancel:    cancel,
	}

	s.mu.Lock()
	s.jobs[jobID] = job
	s.mu.Unlock()

	go s.runJob(ctx, job, upgrades.RankRequest{
		Imported: imported,
		Filters:  input.Filters,
		Policy:   input.Policy,
		Options:  options,
	})

	writeJSON(w, http.StatusAccepted, jobResponse{
		ID:        job.id,
		Status:    job.status,
		Character: &job.character,
	})
}

func defaultSimulationOptions() upgrades.SimulationOptions {
	return upgrades.SimulationOptions{
		ScreeningIterations:    300,
		ConfirmationIterations: 1000,
	}
}

func (s *upgradeServer) runJob(ctx context.Context, job *rankJob, request upgrades.RankRequest) {
	s.mu.Lock()
	if job.status != jobStatusQueued { // canceled before start
		s.mu.Unlock()
		return
	}
	job.status = jobStatusRunning
	s.mu.Unlock()

	report, err := s.ranker.RankUpgrades(ctx, request, func(p upgrades.Progress) {
		s.mu.Lock()
		if job.status == jobStatusRunning {
			job.progress = p
		}
		s.mu.Unlock()
	})

	s.mu.Lock()
	defer s.mu.Unlock()
	if job.status == jobStatusCanceled {
		return // never present a report for a canceled job
	}
	if err != nil {
		if errors.Is(err, context.Canceled) {
			job.status = jobStatusCanceled
			return
		}
		job.status = jobStatusFailed
		job.err = &apiError{Code: "simulation_failed", Message: err.Error()}
		return
	}
	job.status = jobStatusCompleted
	job.report = report
}

func (s *upgradeServer) handleGetJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.mu.Lock()
	job, ok := s.jobs[id]
	if !ok {
		s.mu.Unlock()
		writeError(w, http.StatusNotFound, "not_found", "unknown job id")
		return
	}
	response := jobResponse{
		ID:        job.id,
		Status:    job.status,
		Progress:  job.progress,
		Character: &job.character,
		Error:     job.err,
		Report:    job.report, // nil unless completed
	}
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, response)
}

func (s *upgradeServer) handleCancelJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.mu.Lock()
	job, ok := s.jobs[id]
	if !ok {
		s.mu.Unlock()
		writeError(w, http.StatusNotFound, "not_found", "unknown job id")
		return
	}
	if job.status == jobStatusQueued || job.status == jobStatusRunning {
		job.status = jobStatusCanceled
		job.report = nil
		if job.cancel != nil {
			job.cancel()
		}
	}
	response := jobResponse{
		ID:        job.id,
		Status:    job.status,
		Progress:  job.progress,
		Character: &job.character,
		Error:     job.err,
		Report:    job.report,
	}
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, response)
}

func (s *upgradeServer) cancelAllJobs() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, job := range s.jobs {
		if job.status == jobStatusQueued || job.status == jobStatusRunning {
			job.status = jobStatusCanceled
			job.report = nil
			if job.cancel != nil {
				job.cancel()
			}
		}
	}
}

func newJobID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (s *upgradeServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	s.serveAsset(w, "index.html")
}

// handleAsset serves one validated UI asset; no directory listing.
func (s *upgradeServer) handleAsset(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")
	if !fs.ValidPath(path) {
		writeError(w, http.StatusNotFound, "not_found", "unknown asset")
		return
	}
	s.serveAsset(w, "assets/"+path)
}

func (s *upgradeServer) serveAsset(w http.ResponseWriter, name string) {
	data, err := upgradeUI.ReadFile("upgrade_ui/" + name)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "unknown asset")
		return
	}
	switch {
	case len(name) >= 3 && name[len(name)-3:] == ".js":
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	case len(name) >= 4 && name[len(name)-4:] == ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case len(name) >= 5 && name[len(name)-5:] == ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(data)
}
