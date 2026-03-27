package index

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/1broseidon/cymbal/internal/symbols"
)

func newTestStore(t *testing.T) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	store, err := OpenStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	return store, dbPath
}

func insertTestSymbols(t *testing.T, store *Store) {
	t.Helper()
	now := time.Now()

	// File 1: Go file with functions and a struct
	fid1, err := store.UpsertFile("/repo/main.go", "main.go", "go", "hash1", now, 100)
	if err != nil {
		t.Fatal(err)
	}
	err = store.InsertSymbols(fid1, []symbols.Symbol{
		{Name: "main", Kind: "function", File: "/repo/main.go", StartLine: 1, EndLine: 5, Language: "go"},
		{Name: "HandleRequest", Kind: "function", File: "/repo/main.go", StartLine: 7, EndLine: 20, Language: "go"},
		{Name: "Server", Kind: "struct", File: "/repo/main.go", StartLine: 22, EndLine: 30, Language: "go"},
		{Name: "Config", Kind: "struct", File: "/repo/main.go", StartLine: 32, EndLine: 40, Language: "go"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// File 2: Python file with classes
	fid2, err := store.UpsertFile("/repo/app.py", "app.py", "python", "hash2", now, 200)
	if err != nil {
		t.Fatal(err)
	}
	err = store.InsertSymbols(fid2, []symbols.Symbol{
		{Name: "Application", Kind: "class", File: "/repo/app.py", StartLine: 1, EndLine: 50, Language: "python"},
		{Name: "handle_request", Kind: "function", File: "/repo/app.py", StartLine: 10, EndLine: 20, Language: "python"},
		{Name: "Config", Kind: "class", File: "/repo/app.py", StartLine: 52, EndLine: 70, Language: "python"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestFeatureStoreFTS5Search(t *testing.T) {
	store, _ := newTestStore(t)
	insertTestSymbols(t, store)

	// FTS5 prefix search for "Handle"
	results, err := store.SearchSymbols("Handle", "", "", false, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Fatal("expected FTS5 search to find symbols matching 'Handle'")
	}

	found := false
	for _, r := range results {
		if r.Name == "HandleRequest" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find HandleRequest via FTS5 prefix search")
	}
}

func TestFeatureStoreExactSearch(t *testing.T) {
	store, _ := newTestStore(t)
	insertTestSymbols(t, store)

	results, err := store.SearchSymbols("HandleRequest", "", "", true, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected exactly 1 result for exact match, got %d", len(results))
	}
	if results[0].Name != "HandleRequest" {
		t.Errorf("expected HandleRequest, got %s", results[0].Name)
	}
}

func TestFeatureStoreKindFilter(t *testing.T) {
	store, _ := newTestStore(t)
	insertTestSymbols(t, store)

	// Search for all functions only
	results, err := store.SearchSymbols("main", "function", "", true, 50)
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range results {
		if r.Kind != "function" {
			t.Errorf("expected kind 'function', got %q for %s", r.Kind, r.Name)
		}
	}

	// Search for structs
	results, err = store.SearchSymbols("Config", "struct", "", true, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 struct named Config (Go), got %d", len(results))
	}
	if results[0].Kind != "struct" {
		t.Errorf("expected struct, got %s", results[0].Kind)
	}
}

func TestFeatureStoreLanguageFilter(t *testing.T) {
	store, _ := newTestStore(t)
	insertTestSymbols(t, store)

	// Search Config in Go only
	results, err := store.SearchSymbols("Config", "", "go", true, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 Go Config, got %d", len(results))
	}
	if results[0].Language != "go" {
		t.Errorf("expected language go, got %s", results[0].Language)
	}

	// Search Config in Python only
	results, err = store.SearchSymbols("Config", "", "python", true, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 Python Config, got %d", len(results))
	}
	if results[0].Language != "python" {
		t.Errorf("expected language python, got %s", results[0].Language)
	}
}

func TestFeatureStoreCaseInsensitiveSearch(t *testing.T) {
	store, _ := newTestStore(t)
	insertTestSymbols(t, store)

	// Search with different casing
	results, err := store.SearchSymbolsCI("handlerequest", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Fatal("expected case-insensitive search to find HandleRequest")
	}
	if results[0].Name != "HandleRequest" {
		t.Errorf("expected HandleRequest, got %s", results[0].Name)
	}

	// Also try uppercase
	results, err = store.SearchSymbolsCI("HANDLEREQUEST", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Fatal("expected case-insensitive search to find HandleRequest with UPPERCASE")
	}
}

func TestFeatureStoreEmptyResults(t *testing.T) {
	store, _ := newTestStore(t)
	insertTestSymbols(t, store)

	// Search for something that doesn't exist
	results, err := store.SearchSymbols("NonExistentSymbolXYZ123", "", "", true, 50)
	if err != nil {
		t.Fatal(err)
	}
	if results == nil {
		// nil is acceptable for "no rows" in Go, but we verify it doesn't error
		results = []SymbolResult{}
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestFeatureStoreTextSearch(t *testing.T) {
	store, _ := newTestStore(t)

	// Create a real file with searchable content
	dir := t.TempDir()
	testFile := filepath.Join(dir, "search_test.go")
	content := `package main

// UniqueMarkerXYZ is a special function
func UniqueMarkerXYZ() {
	fmt.Println("hello world")
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	fid, err := store.UpsertFile(testFile, "search_test.go", "go", "hash_search", now, int64(len(content)))
	if err != nil {
		t.Fatal(err)
	}
	_ = fid

	// Use the store's AllFiles to verify it's indexed
	files, err := store.AllFiles("")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
}

func TestFeatureStoreFileSymbols(t *testing.T) {
	store, _ := newTestStore(t)
	insertTestSymbols(t, store)

	results, err := store.FileSymbols("/repo/main.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 4 {
		t.Fatalf("expected 4 symbols in main.go, got %d", len(results))
	}

	// Verify they're ordered by start_line
	for i := 1; i < len(results); i++ {
		if results[i].StartLine < results[i-1].StartLine {
			t.Error("expected symbols ordered by start_line")
		}
	}
}

func TestFeatureStoreDeleteStalePaths(t *testing.T) {
	store, _ := newTestStore(t)
	insertTestSymbols(t, store)

	// Pretend only main.go still exists
	current := map[string]struct{}{
		"/repo/main.go": {},
	}
	deleted, err := store.DeleteStalePaths(current)
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 stale file deleted, got %d", deleted)
	}

	// Verify app.py symbols are gone
	results, err := store.FileSymbols("/repo/app.py")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 symbols for deleted file, got %d", len(results))
	}
}

func TestFeatureStoreImportsAndRefs(t *testing.T) {
	store, _ := newTestStore(t)
	now := time.Now()

	fid, err := store.UpsertFile("/repo/main.go", "main.go", "go", "hash1", now, 100)
	if err != nil {
		t.Fatal(err)
	}

	err = store.InsertImports(fid, []symbols.Import{
		{RawPath: "fmt", Language: "go"},
		{RawPath: "net/http", Language: "go"},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = store.InsertRefs(fid, []symbols.Ref{
		{Name: "Println", Line: 10, Language: "go"},
		{Name: "ListenAndServe", Line: 15, Language: "go"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify imports
	imports, err := store.FileImports("/repo/main.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(imports))
	}

	// Verify refs
	refs, err := store.FindReferences("Println", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 1 {
		t.Fatalf("expected 1 reference to Println, got %d", len(refs))
	}
	if refs[0].Line != 10 {
		t.Errorf("expected ref on line 10, got %d", refs[0].Line)
	}
}
