package cmd

import (
	gopath "path"
	"path/filepath"
	"strings"
)

func normalizeRelPath(rel string) string {
	return filepath.ToSlash(rel)
}

func matchAnyPath(rel string, globs []string) bool {
	rel = normalizeRelPath(rel)
	for _, glob := range globs {
		glob = strings.TrimSpace(glob)
		if glob == "" {
			continue
		}
		if ok, err := gopath.Match(glob, rel); err == nil && ok {
			return true
		}
		if strings.Contains(rel, glob) {
			return true
		}
	}
	return false
}

func widenPathFilterLimit(limit int, hasFilters bool) int {
	if !hasFilters {
		return limit
	}
	if limit <= 0 {
		return 500
	}
	w := limit * 5
	if w < 100 {
		w = 100
	}
	if w > 1000 {
		w = 1000
	}
	return w
}

func filterByPath[T any](items []T, relPath func(T) string, includes, excludes []string) []T {
	if len(includes) == 0 && len(excludes) == 0 {
		return items
	}
	out := make([]T, 0, len(items))
	for _, item := range items {
		rel := normalizeRelPath(relPath(item))
		if len(includes) > 0 && !matchAnyPath(rel, includes) {
			continue
		}
		if len(excludes) > 0 && matchAnyPath(rel, excludes) {
			continue
		}
		out = append(out, item)
	}
	return out
}
