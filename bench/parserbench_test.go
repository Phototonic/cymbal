package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// This file measures tree-sitter parser init cost vs parse cost.
// Run: CGO_CFLAGS="-DSQLITE_ENABLE_FTS5" go test -v -run=TestParserCost ./bench/

var langs = map[string]*sitter.Language{
	"go":         golang.GetLanguage(),
	"python":     python.GetLanguage(),
	"typescript": typescript.GetLanguage(),
}

// collectFiles walks a directory and returns paths matching the given extension.
func collectFiles(dir, ext string, limit int) []string {
	var files []string
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ext {
			files = append(files, path)
			if len(files) >= limit {
				return filepath.SkipAll
			}
		}
		return nil
	})
	return files
}

func TestParserCost(t *testing.T) {
	corpus := filepath.Join(".corpus")
	if _, err := os.Stat(corpus); err != nil {
		t.Skipf("corpus not found at %s — run: go run ./bench setup", corpus)
	}

	tests := []struct {
		name string
		dir  string
		ext  string
		lang string
	}{
		{"cli/go", filepath.Join(corpus, "cli"), ".go", "go"},
		{"fastapi/python", filepath.Join(corpus, "fastapi"), ".py", "python"},
		{"vite/typescript", filepath.Join(corpus, "vite"), ".ts", "typescript"},
	}

	for _, tt := range tests {
		files := collectFiles(tt.dir, tt.ext, 500)
		if len(files) == 0 {
			t.Logf("SKIP %s: no %s files found", tt.name, tt.ext)
			continue
		}

		// Pre-read all files into memory so we're measuring parser, not I/O.
		sources := make([][]byte, len(files))
		for i, f := range files {
			sources[i], _ = os.ReadFile(f)
		}

		lang := langs[tt.lang]

		// --- Measure: new parser per file (current behavior) ---
		var totalInit, totalParse time.Duration

		for i, src := range sources {
			_ = i

			t0 := time.Now()
			p := sitter.NewParser()
			p.SetLanguage(lang)
			initDone := time.Now()

			p.Parse(nil, src)
			parseDone := time.Now()

			totalInit += initDone.Sub(t0)
			totalParse += parseDone.Sub(initDone)

			p.Close()
		}

		// --- Measure: reuse one parser (pooled behavior) ---
		poolParser := sitter.NewParser()
		poolParser.SetLanguage(lang)
		var totalPoolParse time.Duration

		for _, src := range sources {
			t0 := time.Now()
			poolParser.Parse(nil, src)
			totalPoolParse += time.Since(t0)
		}
		poolParser.Close()

		// --- Report ---
		n := len(sources)
		fmt.Printf("\n=== %s (%d files) ===\n", tt.name, n)
		fmt.Printf("  New parser per file:\n")
		fmt.Printf("    init (new+setlang):  %v total, %v/file\n", totalInit, totalInit/time.Duration(n))
		fmt.Printf("    parse:               %v total, %v/file\n", totalParse, totalParse/time.Duration(n))
		fmt.Printf("    init %% of total:     %.1f%%\n", 100*float64(totalInit)/float64(totalInit+totalParse))
		fmt.Printf("  Pooled parser (reuse):\n")
		fmt.Printf("    parse:               %v total, %v/file\n", totalPoolParse, totalPoolParse/time.Duration(n))
		fmt.Printf("  Speedup from pooling:  %.2fx\n", float64(totalInit+totalParse)/float64(totalPoolParse))
	}
}
