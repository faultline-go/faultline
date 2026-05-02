// Package coverage parses Go coverage profiles and computes per-package coverage percentages.
package coverage

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// PackageCoverage holds the coverage percentage for a single package.
type PackageCoverage struct {
	ImportPath  string
	CoveragePct float64
}

type Profile struct {
	Packages map[string]float64
	Warnings []string
}

// block represents a single line from a Go coverage profile.
type block struct {
	stmts   int64
	covered int64
}

// ParseFile parses a Go coverage profile produced by `go test -coverprofile`.
// It returns a map from import path to coverage percentage.
// If the file is empty or the mode line is missing, an error is returned.
func ParseFile(path string) (map[string]float64, error) {
	profile, err := ParseProfile(path)
	if err != nil {
		return nil, err
	}
	return profile.Packages, nil
}

// ParseProfile parses a Go coverage profile and preserves malformed data-line
// warnings. A malformed block should not make the whole scan fail when the mode
// line is valid.
func ParseProfile(path string) (Profile, error) {
	f, err := os.Open(path)
	if err != nil {
		return Profile{}, fmt.Errorf("open coverage file %s: %w", path, err)
	}
	defer f.Close()

	pkgBlocks := make(map[string][]block)
	var warnings []string
	scanner := bufio.NewScanner(f)

	first := true
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if first {
			// First line: "mode: set|count|atomic"
			if !validModeLine(line) {
				return Profile{}, fmt.Errorf("coverage file %s: missing or invalid mode line", path)
			}
			first = false
			continue
		}

		b, pkg, err := parseLine(line)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s:%d: %v", path, lineNo, err))
			continue
		}
		pkgBlocks[pkg] = append(pkgBlocks[pkg], b)
	}
	if err := scanner.Err(); err != nil {
		return Profile{}, fmt.Errorf("read coverage file %s: %w", path, err)
	}
	if first {
		return Profile{}, fmt.Errorf("coverage file %s: missing mode line", path)
	}

	result := make(map[string]float64, len(pkgBlocks))
	for pkg, blocks := range pkgBlocks {
		var totalStmts, coveredStmts int64
		for _, bl := range blocks {
			totalStmts += bl.stmts
			coveredStmts += bl.covered
		}
		if totalStmts == 0 {
			result[pkg] = 0
		} else {
			result[pkg] = float64(coveredStmts) / float64(totalStmts) * 100.0
		}
	}
	return Profile{Packages: result, Warnings: warnings}, nil
}

func validModeLine(line string) bool {
	switch line {
	case "mode: set", "mode: count", "mode: atomic":
		return true
	default:
		return false
	}
}

// parseLine parses one data line from a coverage profile.
// Format: <file>:<startLine>.<startCol>,<endLine>.<endCol> <numStmts> <count>
// Example: github.com/example/pkg/file.go:10.32,14.2 4 1
func parseLine(line string) (block, string, error) {
	// Split off the count at the end.
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return block{}, "", fmt.Errorf("unexpected format: %q", line)
	}

	fileCoords := parts[0] // e.g. "github.com/example/pkg/file.go:10.32,14.2"
	numStmts, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return block{}, "", fmt.Errorf("parse numStmts: %w", err)
	}
	if numStmts < 0 {
		return block{}, "", fmt.Errorf("numStmts must be non-negative")
	}
	count, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return block{}, "", fmt.Errorf("parse count: %w", err)
	}
	if count < 0 {
		return block{}, "", fmt.Errorf("count must be non-negative")
	}

	// Extract the file path (before the colon that separates path from line numbers).
	colonIdx := strings.LastIndex(fileCoords, ":")
	if colonIdx < 0 {
		return block{}, "", fmt.Errorf("no colon in file coords: %q", fileCoords)
	}
	filePath := fileCoords[:colonIdx] // e.g. "github.com/example/pkg/file.go"

	// The package import path is the directory of the file path.
	pkg := packageFromFilePath(filePath)

	covered := int64(0)
	if count > 0 {
		covered = numStmts
	}

	return block{stmts: numStmts, covered: covered}, pkg, nil
}

// packageFromFilePath extracts the package import path from a coverage file path.
// e.g. "github.com/example/pkg/file.go" → "github.com/example/pkg"
func packageFromFilePath(filePath string) string {
	filePath = cleanPackagePath(filePath)
	idx := strings.LastIndex(filePath, "/")
	if idx < 0 {
		return filePath
	}
	return filePath[:idx]
}

func cleanPackagePath(path string) string {
	path = strings.TrimSpace(strings.ReplaceAll(path, `\`, "/"))
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	return strings.TrimRight(path, "/")
}
