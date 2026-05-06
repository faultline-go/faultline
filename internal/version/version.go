package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// These are set via -ldflags at build time.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

var readBuildInfo = debug.ReadBuildInfo

type Info struct {
	Version   string
	Commit    string
	Date      string
	GoVersion string
	OS        string
	Arch      string
}

func Get() Info {
	version := strings.TrimSpace(Version)
	commit := strings.TrimSpace(Commit)
	date := strings.TrimSpace(BuildDate)

	if info, ok := readBuildInfo(); ok {
		if isUnsetVersion(version) {
			if moduleVersion := buildInfoVersion(info); moduleVersion != "" {
				version = moduleVersion
			}
		}
		if isUnsetBuildValue(commit) {
			commit = buildSetting(info, "vcs.revision")
		}
		if isUnsetBuildValue(date) {
			date = buildSetting(info, "vcs.time")
		}
	}

	return Info{
		Version:   sane(version, "dev"),
		Commit:    sane(commit, "unknown"),
		Date:      sane(date, "unknown"),
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

func String() string {
	info := Get()
	if info.Commit == "unknown" {
		return info.Version
	}
	return info.Version + " (" + info.Commit + ")"
}

func FullString() string {
	info := Get()
	return fmt.Sprintf("version: %s\ncommit: %s\ndate: %s\ngo: %s\nos/arch: %s/%s",
		info.Version, info.Commit, info.Date, info.GoVersion, info.OS, info.Arch)
}

func sane(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func buildInfoVersion(info *debug.BuildInfo) string {
	if info == nil {
		return ""
	}
	version := strings.TrimSpace(info.Main.Version)
	if version == "(devel)" {
		return ""
	}
	return version
}

func buildSetting(info *debug.BuildInfo, key string) string {
	if info == nil {
		return ""
	}
	for _, setting := range info.Settings {
		if setting.Key == key {
			return strings.TrimSpace(setting.Value)
		}
	}
	return ""
}

func isUnsetVersion(value string) bool {
	switch strings.TrimSpace(value) {
	case "", "dev", "0.1.1":
		return true
	default:
		return false
	}
}

func isUnsetBuildValue(value string) bool {
	switch strings.TrimSpace(value) {
	case "", "unknown":
		return true
	default:
		return false
	}
}
