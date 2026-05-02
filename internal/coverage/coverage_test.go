package coverage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		wantPkg     string
		wantStmts   int64
		wantCovered int64
		wantErr     bool
	}{
		{
			name:        "covered block",
			line:        "github.com/example/pkg/file.go:10.32,14.2 4 1",
			wantPkg:     "github.com/example/pkg",
			wantStmts:   4,
			wantCovered: 4,
		},
		{
			name:        "uncovered block",
			line:        "github.com/example/pkg/file.go:20.10,25.5 3 0",
			wantPkg:     "github.com/example/pkg",
			wantStmts:   3,
			wantCovered: 0,
		},
		{
			name:        "nested package",
			line:        "github.com/example/pkg/sub/file.go:1.1,2.2 1 2",
			wantPkg:     "github.com/example/pkg/sub",
			wantStmts:   1,
			wantCovered: 1,
		},
		{
			name:        "windows separators",
			line:        `github.com\example\pkg\file.go:10.32,14.2 4 1`,
			wantPkg:     "github.com/example/pkg",
			wantStmts:   4,
			wantCovered: 4,
		},
		{
			name:        "repeated separators",
			line:        "github.com/example//pkg/file.go:10.32,14.2 4 1",
			wantPkg:     "github.com/example/pkg",
			wantStmts:   4,
			wantCovered: 4,
		},
		{
			name:    "malformed",
			line:    "not a valid line",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, pkg, err := parseLine(tt.line)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pkg != tt.wantPkg {
				t.Errorf("pkg = %q, want %q", pkg, tt.wantPkg)
			}
			if b.stmts != tt.wantStmts {
				t.Errorf("stmts = %d, want %d", b.stmts, tt.wantStmts)
			}
			if b.covered != tt.wantCovered {
				t.Errorf("covered = %d, want %d", b.covered, tt.wantCovered)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	content := `mode: set
github.com/example/pkg/a.go:1.10,5.2 4 1
github.com/example/pkg/a.go:6.10,10.2 4 0
github.com/example/other/b.go:1.10,3.2 2 1
`
	dir := t.TempDir()
	f := filepath.Join(dir, "coverage.out")
	if err := os.WriteFile(f, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	result, err := ParseFile(f)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	// pkg: 4 covered, 4 uncovered → 50%
	if got := result["github.com/example/pkg"]; got != 50.0 {
		t.Errorf("pkg coverage = %.1f, want 50.0", got)
	}

	// other: 2 covered, 0 uncovered → 100%
	if got := result["github.com/example/other"]; got != 100.0 {
		t.Errorf("other coverage = %.1f, want 100.0", got)
	}
}

func TestParseFile_MissingMode(t *testing.T) {
	content := "github.com/example/pkg/a.go:1.10,5.2 4 1\n"
	dir := t.TempDir()
	f := filepath.Join(dir, "coverage.out")
	if err := os.WriteFile(f, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := ParseFile(f)
	if err == nil {
		t.Fatal("expected error for missing mode line")
	}
}

func TestParseProfile_PartialMalformed(t *testing.T) {
	content := `mode: count
github.com/example/pkg/a.go:1.10,5.2 4 1
this is malformed
github.com/example/other/b.go:1.10,3.2 2 0
github.com/example/pkg/a.go:6.10,10.2 -1 0
`
	dir := t.TempDir()
	f := filepath.Join(dir, "coverage.out")
	if err := os.WriteFile(f, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	profile, err := ParseProfile(f)
	if err != nil {
		t.Fatalf("ParseProfile: %v", err)
	}
	if got := profile.Packages["github.com/example/pkg"]; got != 100.0 {
		t.Fatalf("pkg coverage = %.1f, want 100.0", got)
	}
	if got := profile.Packages["github.com/example/other"]; got != 0.0 {
		t.Fatalf("other coverage = %.1f, want 0.0", got)
	}
	if len(profile.Warnings) != 2 {
		t.Fatalf("warnings = %d, want 2: %+v", len(profile.Warnings), profile.Warnings)
	}
}

func TestParseProfile_InvalidMode(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "coverage.out")
	if err := os.WriteFile(f, []byte("mode: bad\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseProfile(f); err == nil {
		t.Fatal("expected invalid mode error")
	}
}

func TestParseFile_NotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/path/coverage.out")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}
