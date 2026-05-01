package analyzer

import (
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type coverageResolver struct {
	matches map[string]coverageMatch
}

type coverageMatch struct {
	Key      string
	Pct      float64
	priority int
}

func newCoverageResolver(values map[string]float64, pkgs []LoadedPackage, repoRoot string) coverageResolver {
	resolver := coverageResolver{matches: map[string]coverageMatch{}}
	if len(values) == 0 || len(pkgs) == 0 {
		return resolver
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		type candidate struct {
			pkg      LoadedPackage
			priority int
		}
		var candidates []candidate
		for _, pkg := range pkgs {
			if priority := coverageKeyMatchPriority(key, pkg, repoRoot); priority > 0 {
				candidates = append(candidates, candidate{pkg: pkg, priority: priority})
			}
		}
		if len(candidates) == 1 {
			match := coverageMatch{Key: cleanCoverageKey(key), Pct: values[key], priority: candidates[0].priority}
			existing, exists := resolver.matches[candidates[0].pkg.ID]
			if !exists || betterCoverageMatch(match, existing) {
				resolver.matches[candidates[0].pkg.ID] = match
			}
		}
	}
	return resolver
}

func (r coverageResolver) ForPackage(pkg LoadedPackage) (coverageMatch, bool) {
	match, ok := r.matches[pkg.ID]
	return match, ok
}

func coverageKeyMatchPriority(key string, pkg LoadedPackage, repoRoot string) int {
	key = cleanCoverageKey(key)
	if key == "" {
		return 0
	}
	for _, candidate := range coverageKeyCandidates(pkg, repoRoot) {
		if key == candidate {
			return 2
		}
	}

	// Some coverprofile producers emit shortened package paths. Keep this as a
	// compatibility fallback, but newCoverageResolver only uses it when the key
	// maps to exactly one loaded package.
	if strings.HasSuffix(pkg.ImportPath, "/"+key) || strings.HasSuffix(filepath.ToSlash(pkg.Dir), "/"+key) {
		return 1
	}
	return 0
}

func betterCoverageMatch(candidate, existing coverageMatch) bool {
	if candidate.priority != existing.priority {
		return candidate.priority > existing.priority
	}
	if len(candidate.Key) != len(existing.Key) {
		return len(candidate.Key) > len(existing.Key)
	}
	return candidate.Key < existing.Key
}

func coverageKeyCandidates(pkg LoadedPackage, repoRoot string) []string {
	candidates := []string{
		pkg.ImportPath,
		pkg.ID,
		filepath.ToSlash(pkg.Dir),
		safeRel(repoRoot, pkg.Dir),
	}

	if pkg.ModuleRoot != "" {
		moduleRoot := filepath.Join(repoRoot, filepath.FromSlash(pkg.ModuleRoot))
		moduleRel := safeRel(moduleRoot, pkg.Dir)
		candidates = append(candidates, moduleRel)
		if pkg.ModulePath != "" {
			if moduleRel == "." {
				candidates = append(candidates, pkg.ModulePath)
			} else {
				candidates = append(candidates, path.Join(pkg.ModulePath, filepath.ToSlash(moduleRel)))
			}
		}
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		candidate = cleanCoverageKey(candidate)
		if candidate == "" {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		out = append(out, candidate)
	}
	return out
}

func cleanCoverageKey(key string) string {
	key = strings.TrimSpace(strings.ReplaceAll(key, `\`, "/"))
	key = filepath.ToSlash(key)
	for strings.Contains(key, "//") {
		key = strings.ReplaceAll(key, "//", "/")
	}
	key = strings.TrimPrefix(key, "./")
	if key != "." {
		key = strings.TrimRight(key, "/")
	}
	return key
}
