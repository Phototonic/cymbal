package index

import (
	"sort"
	"strings"
)

// rankFetchWindow returns the over-fetch size: at least limit*5, capped at 500.
func rankFetchWindow(userLimit int) int {
	if userLimit <= 0 {
		return 500
	}
	w := userLimit * 5
	if w > 500 {
		w = 500
	}
	return w
}

// RankSymbols sorts a slice of SymbolResults so the canonical definition
// appears first. Identical logic is used by both the store (over-fetch window)
// and the cmd layer so they are always consistent.
func RankSymbols(results []SymbolResult) {
	sort.SliceStable(results, func(i, j int) bool {
		return SymbolScore(results[i]) > SymbolScore(results[j])
	})
}

// SymbolScore returns a heuristic relevance score for a symbol result.
// Higher is better / more canonical.
func SymbolScore(r SymbolResult) int {
	score := 0
	p := strings.ToLower(r.RelPath)

	// Kind priority.
	switch r.Kind {
	case "class", "struct", "interface", "type":
		score += 60
	case "function":
		score += 50
	case "method":
		score += 40
	case "enum":
		score += 30
	case "constructor":
		score += 20
	case "impl":
		score += 15
	case "variable", "constant":
		score += 10
	}

	// Penalise test paths.
	for _, seg := range []string{
		"/test/", "/tests/", "/testing/",
		"_test.go", "_test.", "_spec.", ".test.", ".spec.",
		"/testdata/", "/testutil/", "/testutils/",
	} {
		if strings.Contains(p, seg) {
			score -= 80
			break
		}
	}
	// Penalise playground / example paths.
	for _, seg := range []string{
		"/playground/", "/example/", "/examples/",
		"/demo/", "/demos/", "/sample/", "/samples/",
		"/fixture/", "/fixtures/",
	} {
		if strings.Contains(p, seg) {
			score -= 70
			break
		}
	}
	// Penalise doc paths.
	for _, seg := range []string{
		"/docs/", "/docs_src/", "/doc/", "/documentation/",
	} {
		if strings.Contains(p, seg) {
			score -= 60
			break
		}
	}
	// Penalise vendored / third-party paths.
	for _, seg := range []string{
		"/vendor/", "/node_modules/", "/third_party/",
		"/external/", "/deps/",
	} {
		if strings.Contains(p, seg) {
			score -= 90
			break
		}
	}
	// Penalise mirror / alternate build trees (e.g. guava android/).
	for _, prefix := range []string{
		"android/", "guava-gwt/",
	} {
		if strings.HasPrefix(p, prefix) || strings.Contains(p, "/"+prefix) {
			score -= 50
			break
		}
	}

	// Prefer well-known source roots.
	for _, seg := range []string{
		"/src/", "/pkg/", "/lib/", "/crates/",
		"/packages/", "/internal/", "/cmd/",
	} {
		if strings.Contains(p, seg) {
			score += 15
			break
		}
	}

	// Shallower paths are more likely canonical.
	score -= strings.Count(r.RelPath, "/") * 3
	// Shorter path as minor tiebreaker.
	score -= len(r.RelPath) / 10

	return score
}
