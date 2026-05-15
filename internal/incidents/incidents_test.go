package incidents

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadFileParsesIncidentsDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "incidents.json")
	data := `{
  "schema": "faultline.incidents.v1",
  "incidents": [{
    "id": "INC-1",
    "title": "Checkout outage",
    "severity": "p1",
    "started_at": "2026-05-14T12:00:00Z",
    "resolved_at": "2026-05-14T12:30:00Z",
    "affected_packages": ["example.com/acme/checkout"],
    "url": "https://status.example.test/incidents/1"
  }]
}`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}

	file, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	if file.Schema != SchemaVersion {
		t.Fatalf("schema = %q", file.Schema)
	}
	if len(file.Incidents) != 1 {
		t.Fatalf("incident count = %d", len(file.Incidents))
	}
	incident := file.Incidents[0]
	if incident.ID != "INC-1" || incident.Title != "Checkout outage" || incident.Severity != SeverityP1 {
		t.Fatalf("incident = %+v", incident)
	}
	if incident.ResolvedAt == nil || incident.ResolvedAt.Format(time.RFC3339) != "2026-05-14T12:30:00Z" {
		t.Fatalf("resolved_at = %+v", incident.ResolvedAt)
	}
}

func TestLoadFileRejectsUnsupportedSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "incidents.json")
	if err := os.WriteFile(path, []byte(`{"schema":"faultline.incidents.v0","incidents":[]}`), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPackageIndexFiltersByLookbackAndIndexesAffectedPackages(t *testing.T) {
	now := time.Now().UTC()
	recent := Incident{
		ID:               "INC-recent",
		Title:            "Recent incident",
		Severity:         SeverityHigh,
		StartedAt:        now.Add(-2 * time.Hour),
		AffectedPackages: []string{"example.com/acme/a", "example.com/acme/b"},
	}
	old := Incident{
		ID:               "INC-old",
		Title:            "Old incident",
		Severity:         SeverityLow,
		StartedAt:        now.AddDate(0, 0, -120),
		AffectedPackages: []string{"example.com/acme/a"},
	}

	index := PackageIndex([]Incident{recent, old}, 90)
	if got := len(index["example.com/acme/a"]); got != 1 {
		t.Fatalf("package a incident count = %d", got)
	}
	if got := len(index["example.com/acme/b"]); got != 1 {
		t.Fatalf("package b incident count = %d", got)
	}
	if index["example.com/acme/a"][0].ID != "INC-recent" {
		t.Fatalf("package a incidents = %+v", index["example.com/acme/a"])
	}
}

func TestPackageIndexDefaultsLookbackWhenNonPositive(t *testing.T) {
	recent := Incident{
		ID:               "INC-default-lookback",
		Title:            "Default lookback incident",
		Severity:         SeverityMedium,
		StartedAt:        time.Now().UTC().AddDate(0, 0, -30),
		AffectedPackages: []string{"example.com/acme/a"},
	}

	index := PackageIndex([]Incident{recent}, 0)
	if len(index["example.com/acme/a"]) != 1 {
		t.Fatalf("index = %+v", index)
	}
}

func TestToSnapshotIncidentsConvertsAndCopiesAffectedPackages(t *testing.T) {
	startedAt := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	resolvedAt := startedAt.Add(30 * time.Minute)
	input := []Incident{{
		ID:               "INC-1",
		Title:            "Checkout outage",
		Severity:         SeverityCritical,
		StartedAt:        startedAt,
		ResolvedAt:       &resolvedAt,
		AffectedPackages: []string{"example.com/acme/checkout"},
		URL:              "https://status.example.test/incidents/1",
	}}

	out := ToSnapshotIncidents(input)
	if len(out) != 1 {
		t.Fatalf("snapshot incident count = %d", len(out))
	}
	if out[0].ID != "INC-1" || out[0].Severity != "critical" || out[0].StartedAt != startedAt || out[0].ResolvedAt != &resolvedAt {
		t.Fatalf("snapshot incident = %+v", out[0])
	}
	input[0].AffectedPackages[0] = "example.com/acme/mutated"
	if out[0].AffectedPackages[0] != "example.com/acme/checkout" {
		t.Fatalf("affected packages were not copied: %+v", out[0].AffectedPackages)
	}
}
