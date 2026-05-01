package analyzer

import (
	"path/filepath"
	"testing"
)

func TestCoverageResolverDisambiguatesModules(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "repo")
	pkgs := []LoadedPackage{
		{
			ID:         "service-a|example.com/acme/pkg",
			ImportPath: "example.com/acme/pkg",
			Dir:        filepath.Join(repo, "service-a", "pkg"),
			ModulePath: "example.com/acme",
			ModuleRoot: "service-a",
		},
		{
			ID:         "service-b|example.com/acme/pkg",
			ImportPath: "example.com/acme/pkg",
			Dir:        filepath.Join(repo, "service-b", "pkg"),
			ModulePath: "example.com/acme",
			ModuleRoot: "service-b",
		},
	}

	resolver := newCoverageResolver(map[string]float64{
		"service-a/pkg": 81.25,
		"pkg":           12.50,
	}, pkgs, repo)

	match, ok := resolver.ForPackage(pkgs[0])
	if !ok {
		t.Fatal("expected service-a package coverage")
	}
	if match.Pct != 81.25 || match.Key != "service-a/pkg" {
		t.Fatalf("service-a match = %+v, want key service-a/pkg pct 81.25", match)
	}
	if _, ok := resolver.ForPackage(pkgs[1]); ok {
		t.Fatal("ambiguous suffix coverage key should not map to service-b package")
	}
}

func TestCoverageResolverCompatibilityMatchesUniqueSuffix(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "repo")
	pkg := LoadedPackage{
		ID:         "service-a|example.com/acme/internal/orders",
		ImportPath: "example.com/acme/internal/orders",
		Dir:        filepath.Join(repo, "service-a", "internal", "orders"),
		ModulePath: "example.com/acme",
		ModuleRoot: "service-a",
	}

	resolver := newCoverageResolver(map[string]float64{
		"internal/orders": 62.5,
	}, []LoadedPackage{pkg}, repo)

	match, ok := resolver.ForPackage(pkg)
	if !ok {
		t.Fatal("expected unique suffix coverage match")
	}
	if match.Pct != 62.5 || match.Key != "internal/orders" {
		t.Fatalf("match = %+v, want key internal/orders pct 62.5", match)
	}
}

func TestCoverageResolverExactImportPath(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "repo")
	pkg := LoadedPackage{
		ID:         "root|example.com/acme/internal/orders",
		ImportPath: "example.com/acme/internal/orders",
		Dir:        filepath.Join(repo, "internal", "orders"),
		ModulePath: "example.com/acme",
		ModuleRoot: ".",
	}

	resolver := newCoverageResolver(map[string]float64{
		"example.com/acme/internal/orders": 93,
	}, []LoadedPackage{pkg}, repo)

	match, ok := resolver.ForPackage(pkg)
	if !ok {
		t.Fatal("expected exact import path coverage match")
	}
	if match.Pct != 93 || match.Key != "example.com/acme/internal/orders" {
		t.Fatalf("match = %+v, want exact import path pct 93", match)
	}
}

func TestCoverageResolverPrefersPreciseKey(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "repo")
	pkg := LoadedPackage{
		ID:         "service-a|example.com/acme/internal/orders",
		ImportPath: "example.com/acme/internal/orders",
		Dir:        filepath.Join(repo, "service-a", "internal", "orders"),
		ModulePath: "example.com/acme",
		ModuleRoot: "service-a",
	}

	resolver := newCoverageResolver(map[string]float64{
		"internal/orders":                  25,
		"example.com/acme/internal/orders": 80,
	}, []LoadedPackage{pkg}, repo)

	match, ok := resolver.ForPackage(pkg)
	if !ok {
		t.Fatal("expected coverage match")
	}
	if match.Pct != 80 || match.Key != "example.com/acme/internal/orders" {
		t.Fatalf("match = %+v, want precise import path pct 80", match)
	}
}

func TestCleanCoverageKey(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "trims whitespace", in: "  example.com/acme/orders  ", want: "example.com/acme/orders"},
		{name: "converts backslashes", in: `example.com\acme\orders`, want: "example.com/acme/orders"},
		{name: "trims leading dot slash", in: "./internal/orders", want: "internal/orders"},
		{name: "collapses repeated slashes", in: "example.com//acme///orders", want: "example.com/acme/orders"},
		{name: "trims trailing slash", in: "example.com/acme/orders/", want: "example.com/acme/orders"},
		{name: "keeps root dot stable", in: ".", want: "."},
		{name: "keeps empty stable", in: "  ", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanCoverageKey(tt.in); got != tt.want {
				t.Fatalf("cleanCoverageKey(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
