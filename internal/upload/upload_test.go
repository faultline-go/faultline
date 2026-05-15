package upload

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConfigIsConfigured(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want bool
	}{
		{
			name: "all fields set",
			cfg:  Config{BaseURL: "https://api.example.test", Token: "flt_token", OrgID: "org-123"},
			want: true,
		},
		{
			name: "missing base url",
			cfg:  Config{Token: "flt_token", OrgID: "org-123"},
		},
		{
			name: "missing token",
			cfg:  Config{BaseURL: "https://api.example.test", OrgID: "org-123"},
		},
		{
			name: "missing org id",
			cfg:  Config{BaseURL: "https://api.example.test", Token: "flt_token"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.IsConfigured(); got != tt.want {
				t.Fatalf("IsConfigured() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUploadSnapshotRequestAndSuccessResponse(t *testing.T) {
	var gotPath, gotAuth, gotContentType, gotUserAgent string
	var gotBody map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		gotUserAgent = r.Header.Get("User-Agent")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"snapshot_id":"snap-1","repo_id":"repo-1","package_count":3,"finding_count":2,"created_at":"2026-05-15T00:00:00Z"}`))
	}))
	defer server.Close()

	result, err := UploadSnapshot(context.Background(), Config{
		BaseURL: server.URL + "/",
		Token:   "flt_token",
		OrgID:   "org-123",
	}, []byte(`{"schema_version":"faultline.snapshot.v1"}`))
	if err != nil {
		t.Fatalf("UploadSnapshot() error = %v", err)
	}
	if result.SnapshotID != "snap-1" || result.RepoID != "repo-1" || result.PackageCount != 3 || result.FindingCount != 2 {
		t.Fatalf("result = %+v", result)
	}
	if gotPath != "/v1/orgs/org-123/snapshots" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotAuth != "Bearer flt_token" {
		t.Fatalf("authorization header = %q", gotAuth)
	}
	if gotContentType != "application/json" {
		t.Fatalf("content-type header = %q", gotContentType)
	}
	if gotUserAgent != "faultline-cli/upload" {
		t.Fatalf("user-agent header = %q", gotUserAgent)
	}
	if gotBody["schema_version"] != "faultline.snapshot.v1" {
		t.Fatalf("request body = %+v", gotBody)
	}
}

func TestUploadSnapshotStatusHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		want       string
	}{
		{name: "unauthorized", statusCode: http.StatusUnauthorized, body: "nope", want: "invalid or expired API token (401)"},
		{name: "forbidden", statusCode: http.StatusForbidden, body: "nope", want: "token lacks permission to upload snapshots (403)"},
		{name: "rate limited", statusCode: http.StatusTooManyRequests, body: "slow down", want: "rate limit exceeded"},
		{name: "too large", statusCode: http.StatusRequestEntityTooLarge, body: "too large", want: "snapshot too large"},
		{name: "generic failure includes body", statusCode: http.StatusBadGateway, body: "upstream unavailable", want: "HTTP 502: upstream unavailable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			_, err := UploadSnapshot(context.Background(), Config{
				BaseURL: server.URL,
				Token:   "flt_token",
				OrgID:   "org-123",
			}, []byte(`{}`))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want substring %q", err.Error(), tt.want)
			}
		})
	}
}

func TestUploadSnapshotRejectsIncompleteConfig(t *testing.T) {
	_, err := UploadSnapshot(context.Background(), Config{BaseURL: "https://api.example.test", Token: "flt_token"}, []byte(`{}`))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--enterprise-url") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestUploadSnapshotMalformedSuccessJSONIsNotFatal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"snapshot_id":`))
	}))
	defer server.Close()

	result, err := UploadSnapshot(context.Background(), Config{
		BaseURL: server.URL,
		Token:   "flt_token",
		OrgID:   "org-123",
	}, []byte(`{}`))
	if err != nil {
		t.Fatalf("UploadSnapshot() error = %v", err)
	}
	if result != (Result{}) {
		t.Fatalf("result = %+v, want zero result", result)
	}
}
