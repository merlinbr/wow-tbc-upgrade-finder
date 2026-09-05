package cmd

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wowsims/tbc/cmd/wowsimcli/cmd/upgrades"
	"github.com/wowsims/tbc/sim"
)

func init() {
	sim.RegisterAll()
}

func fixtureLink(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("upgrades/testdata/fixed_individual_link.txt")
	if err != nil {
		t.Fatalf("failed to read fixture link: %v", err)
	}
	return strings.TrimSpace(string(data))
}

func optionalSettingsFixtureLink(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("upgrades/testdata/retribution_no_settings_link.txt")
	if err != nil {
		t.Fatalf("failed to read optional-settings fixture: %v", err)
	}
	return strings.TrimSpace(string(data))
}


type fakeRanker struct {
	mu    sync.Mutex
	runs  int
	block bool
}

func (f *fakeRanker) RankUpgrades(ctx context.Context, request upgrades.RankRequest, progress func(upgrades.Progress)) (*upgrades.UpgradeReport, error) {
	f.mu.Lock()
	f.runs++
	f.mu.Unlock()

	if f.block {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	if progress != nil {
		progress(upgrades.Progress{Stage: "screening", Completed: 1, Total: 2})
		progress(upgrades.Progress{Stage: "confirmation", Completed: 2, Total: 2})
	}
	return &upgrades.UpgradeReport{
		Baseline:               upgrades.BaselineSummary{Dps: 1200, DpsStdev: 10, Iterations: 100},
		AssumptionsFingerprint: "fingerprint",
		SimulatorRevision:      "sim",
		DatabaseRevision:       "db",
		Confirmed: []upgrades.ConfirmedUpgrade{
			{Rank: 1, Item: upgrades.UIItemSummary{ID: 2001, Name: "Chest 2001"}, TargetSlot: 4, Iterations: 100, DpsDelta: 15},
		},
	}, nil
}

func (f *fakeRanker) Runs() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.runs
}

func newTestServer(t *testing.T, ranker Ranker) *httptest.Server {
	t.Helper()
	s := newUpgradeServer("test")
	s.ranker = ranker
	server := httptest.NewServer(s.routes())
	t.Cleanup(server.Close)
	return server
}

func postJSON(t *testing.T, url string, body any) *http.Response {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.Post(url, "application/json", strings.NewReader(string(payload)))
	if err != nil {
		t.Fatalf("POST %s failed: %v", url, err)
	}
	t.Cleanup(func() { response.Body.Close() })
	return response
}

func decodeBody(t *testing.T, response *http.Response) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return payload
}

func fixtureImportBody(t *testing.T) map[string]string {
	t.Helper()
	return map[string]string{"link": fixtureLink(t)}
}

func fixtureRankBody(t *testing.T) map[string]any {
	t.Helper()
	return map[string]any{
		"link": fixtureLink(t),
		"filters": map[string]any{
			"maxPhase":       2,
			"sourceKinds":    []string{},
			"sourceNames":    []string{},
			"includeUnknown": false,
		},
		"policy":  map[string]any{},
		"options": map[string]any{"screeningIterations": 10, "confirmationIterations": 20},
	}
}

func TestImportReturnsSummaryWithoutStartingJob(t *testing.T) {
	fake := &fakeRanker{}
	server := newTestServer(t, fake)

	response := postJSON(t, server.URL+"/api/import", fixtureImportBody(t))
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}
	if fake.Runs() != 0 {
		t.Fatalf("runs = %d, want 0", fake.Runs())
	}

	payload := decodeBody(t, response)
	character, ok := payload["character"].(map[string]any)
	if !ok {
		t.Fatal("import response has no character summary")
	}
	if character["class"] != "ClassMage" {
		t.Fatalf("class = %v, want ClassMage", character["class"])
	}
	if digest, _ := payload["settingsDigest"].(string); digest == "" {
		t.Fatal("settingsDigest missing")
	}
	gear, ok := payload["gear"].([]any)
	if !ok || len(gear) != 17 {
		t.Fatalf("gear = %#v, want array of 17 slots", payload["gear"])
	}
	if _, ok := payload["talentsString"].(string); !ok {
		t.Fatalf("talentsString = %#v, want string", payload["talentsString"])
	}
	firstGear, ok := gear[0].(map[string]any)
	if !ok {
		t.Fatal("gear[0] is not an object")
	}
	if ilvl, _ := firstGear["ilvl"].(float64); ilvl <= 0 {
		t.Fatalf("gear[0].ilvl = %v, want > 0", firstGear["ilvl"])
	}
	if _, ok := payload["stats"].(map[string]any); !ok {
		t.Fatalf("stats = %#v, want object", payload["stats"])
	}
	if _, ok := payload["derivedStats"].(map[string]any); !ok {
		t.Fatalf("derivedStats = %#v, want object", payload["derivedStats"])
	}
}

func TestImportDefaultsPhaseWhenExportOmitsSimSettings(t *testing.T) {
	server := newTestServer(t, &fakeRanker{})

	response := postJSON(t, server.URL+"/api/import", map[string]string{
		"link": optionalSettingsFixtureLink(t),
	})
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}

	defaults, ok := decodeBody(t, response)["defaults"].(map[string]any)
	if !ok || defaults["maxPhase"] != float64(2) {
		t.Fatalf("defaults = %#v, want maxPhase 2 (highest equipped-item phase)", defaults)
	}
}

func TestCreateJobPollsCompletedReport(t *testing.T) {
	fake := &fakeRanker{}
	server := newTestServer(t, fake)

	id := createJob(t, server.URL, fixtureRankBody(t))
	job := waitForJob(t, server.URL, id)
	if job["status"] != "completed" {
		t.Fatalf("job status = %v, want completed", job["status"])
	}
	if job["report"] == nil {
		t.Fatal("completed job has no report")
	}
}

func TestInvalidImportIsSpecificAndDoesNotCreateJob(t *testing.T) {
	fake := &fakeRanker{}
	server := newTestServer(t, fake)

	response := postJSON(t, server.URL+"/api/import", map[string]string{"link": "not-a-link"})
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.StatusCode)
	}
	payload := decodeBody(t, response)
	errObj, ok := payload["error"].(map[string]any)
	if !ok || errObj["code"] != "invalid_link" {
		t.Fatalf("error = %#v, want code invalid_link", payload["error"])
	}
	if fake.Runs() != 0 {
		t.Fatalf("runs = %d, want 0", fake.Runs())
	}
}

func TestCancelJobReturnsCanceledWithoutPartialReport(t *testing.T) {
	fake := &fakeRanker{block: true}
	server := newTestServer(t, fake)

	id := createJob(t, server.URL, fixtureRankBody(t))

	// Wait until the job is actually running, then cancel.
	waitForStatus(t, server.URL, id, "running")
	deleteResponse := deleteJob(t, server.URL, id)
	payload := decodeBody(t, deleteResponse)
	if payload["status"] != "canceled" {
		t.Fatalf("cancel status = %v, want canceled", payload["status"])
	}
	if payload["report"] != nil {
		t.Fatal("canceled job has a report")
	}

	job := waitForJob(t, server.URL, id)
	if job["status"] != "canceled" || job["report"] != nil {
		t.Fatalf("job = %#v, want canceled with no report", job)
	}
}

func TestJobRejectsUnknownFieldsAndOversizedBodies(t *testing.T) {
	fake := &fakeRanker{}
	server := newTestServer(t, fake)

	response := postJSON(t, server.URL+"/api/jobs", map[string]any{
		"link": fixtureLink(t), "bogus": 1,
	})
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("unknown field status = %d, want 400", response.StatusCode)
	}

	huge := strings.Repeat("a", 2<<20)
	response = postJSON(t, server.URL+"/api/import", map[string]string{"link": huge})
	if response.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized body status = %d, want 413", response.StatusCode)
	}
}

func TestUnknownPathAndMethodAreStructuredErrors(t *testing.T) {
	fake := &fakeRanker{}
	server := newTestServer(t, fake)

	response, err := http.Get(server.URL + "/api/unknown")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown path status = %d, want 404", response.StatusCode)
	}

	response, err = http.Get(server.URL + "/api/import")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("wrong method status = %d, want 405", response.StatusCode)
	}
}

func createJob(t *testing.T, baseURL string, body map[string]any) string {
	t.Helper()
	response := postJSON(t, baseURL+"/api/jobs", body)
	if response.StatusCode != http.StatusAccepted {
		payload := decodeBody(t, response)
		t.Fatalf("create job status = %d, payload = %#v", response.StatusCode, payload)
	}
	payload := decodeBody(t, response)
	id, _ := payload["id"].(string)
	if id == "" {
		t.Fatalf("job id missing: %#v", payload)
	}
	return id
}

func waitForJob(t *testing.T, baseURL, id string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		job := getJob(t, baseURL, id)
		switch job["status"] {
		case "completed", "failed", "canceled":
			return job
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("job did not reach a terminal state")
	return nil
}

func waitForStatus(t *testing.T, baseURL, id, status string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		job := getJob(t, baseURL, id)
		if job["status"] == status {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("job did not reach status %q", status)
}

func getJob(t *testing.T, baseURL, id string) map[string]any {
	t.Helper()
	response, err := http.Get(baseURL + "/api/jobs/" + id)
	if err != nil {
		t.Fatalf("GET job failed: %v", err)
	}
	defer response.Body.Close()
	return decodeBody(t, response)
}

func deleteJob(t *testing.T, baseURL, id string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(http.MethodDelete, baseURL+"/api/jobs/"+id, nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("DELETE job failed: %v", err)
	}
	t.Cleanup(func() { response.Body.Close() })
	return response
}


var localAssetURLPattern = regexp.MustCompile(`(?:src|href)="((?:/|\./)assets/[^"]+)"`)

func localAssetURLs(page string) []string {
	matches := localAssetURLPattern.FindAllStringSubmatch(page, -1)
	assets := make([]string, 0, len(matches))
	for _, match := range matches {
		assets = append(assets, strings.TrimPrefix(match[1], "."))
	}
	return assets
}

func TestUpgradePageAndAssetsAreServed(t *testing.T) {
	fake := &fakeRanker{}
	server := newTestServer(t, fake)

	response, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", response.StatusCode)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	page := string(body)
	if !strings.Contains(page, `<div id="app"></div>`) {
		t.Fatal(`GET / missing Vite app mount "<div id="app"></div>"`)
	}

	assets := localAssetURLs(page)
	if len(assets) == 0 {
		t.Fatal("GET / has no local asset URLs")
	}
	for _, assetURL := range assets {
		assetResponse, err := http.Get(server.URL + assetURL)
		if err != nil {
			t.Fatalf("GET %s failed: %v", assetURL, err)
		}
		if assetResponse.StatusCode != http.StatusOK {
			assetResponse.Body.Close()
			t.Fatalf("GET %s status = %d, want 200", assetURL, assetResponse.StatusCode)
		}
		contentType := assetResponse.Header.Get("Content-Type")
		switch {
		case strings.HasSuffix(assetURL, ".js"):
			if !strings.Contains(contentType, "javascript") {
				assetResponse.Body.Close()
				t.Fatalf("%s content type = %q, want javascript", assetURL, contentType)
			}
		case strings.HasSuffix(assetURL, ".css"):
			if !strings.Contains(contentType, "css") {
				assetResponse.Body.Close()
				t.Fatalf("%s content type = %q, want css", assetURL, contentType)
			}
		default:
			assetResponse.Body.Close()
			t.Fatalf("unexpected local asset URL %q", assetURL)
		}
		assetResponse.Body.Close()
	}

	missingResponse, err := http.Get(server.URL + "/assets/not/a/real-file.js")
	if err != nil {
		t.Fatal(err)
	}
	if missingResponse.StatusCode != http.StatusNotFound {
		missingResponse.Body.Close()
		t.Fatalf("unknown asset status = %d, want 404", missingResponse.StatusCode)
	}
	missingPayload := decodeBody(t, missingResponse)
	missingResponse.Body.Close()
	errObj, ok := missingPayload["error"].(map[string]any)
	if !ok || errObj["code"] != "not_found" {
		t.Fatalf("unknown asset error = %#v, want code not_found", missingPayload["error"])
	}
}