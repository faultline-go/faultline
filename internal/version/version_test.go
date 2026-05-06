package version

import (
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
)

func TestDefaultsAreSane(t *testing.T) {
	info := Get()
	if info.Version == "" || info.Commit == "" || info.Date == "" {
		t.Fatalf("metadata defaults should be populated: %+v", info)
	}
	if info.GoVersion != runtime.Version() {
		t.Fatalf("GoVersion = %q, want %q", info.GoVersion, runtime.Version())
	}
	if info.OS != runtime.GOOS || info.Arch != runtime.GOARCH {
		t.Fatalf("os/arch = %s/%s, want %s/%s", info.OS, info.Arch, runtime.GOOS, runtime.GOARCH)
	}
}

func TestFullStringHasRequiredFields(t *testing.T) {
	out := FullString()
	for _, field := range []string{"version:", "commit:", "date:", "go:", "os/arch:"} {
		if !strings.Contains(out, field) {
			t.Fatalf("FullString missing %q:\n%s", field, out)
		}
	}
}

func TestGetUsesModuleBuildInfoWhenLdflagsAreDefaults(t *testing.T) {
	restoreVersionState(t, "0.1.1", "unknown", "unknown", &debug.BuildInfo{
		Main: debug.Module{Version: "v0.2.6"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "abcdef1234567890"},
			{Key: "vcs.time", Value: "2026-05-06T17:30:00Z"},
		},
	})

	info := Get()
	if info.Version != "v0.2.6" {
		t.Fatalf("Version = %q, want module build version", info.Version)
	}
	if info.Commit != "abcdef1234567890" {
		t.Fatalf("Commit = %q, want vcs revision", info.Commit)
	}
	if info.Date != "2026-05-06T17:30:00Z" {
		t.Fatalf("Date = %q, want vcs time", info.Date)
	}
}

func TestGetPrefersLdflagsOverBuildInfo(t *testing.T) {
	restoreVersionState(t, "v9.9.9", "commit-from-ldflags", "date-from-ldflags", &debug.BuildInfo{
		Main: debug.Module{Version: "v0.2.6"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "abcdef1234567890"},
			{Key: "vcs.time", Value: "2026-05-06T17:30:00Z"},
		},
	})

	info := Get()
	if info.Version != "v9.9.9" || info.Commit != "commit-from-ldflags" || info.Date != "date-from-ldflags" {
		t.Fatalf("ldflags metadata not preferred: %+v", info)
	}
}

func restoreVersionState(t *testing.T, version, commit, date string, buildInfo *debug.BuildInfo) {
	t.Helper()
	oldVersion, oldCommit, oldBuildDate := Version, Commit, BuildDate
	oldReadBuildInfo := readBuildInfo
	Version, Commit, BuildDate = version, commit, date
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return buildInfo, buildInfo != nil
	}
	t.Cleanup(func() {
		Version, Commit, BuildDate = oldVersion, oldCommit, oldBuildDate
		readBuildInfo = oldReadBuildInfo
	})
}
