package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1broseidon/cymbal/index"
)

func TestRepoBoundFilePathRejectsOutsideRepo(t *testing.T) {
	t.Cleanup(index.CloseAll)

	repoDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\nfunc inside() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(t.TempDir(), "test.db")
	if _, err := index.Index(repoDir, dbPath, index.Options{Workers: 1}); err != nil {
		t.Fatal(err)
	}

	outside := filepath.Join(t.TempDir(), "secret.go")
	if err := os.WriteFile(outside, []byte("package secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := repoBoundFilePath(dbPath, outside)
	if err == nil || !strings.Contains(err.Error(), "outside repository") {
		t.Fatalf("expected outside-repository error, got %v", err)
	}
}

func TestRepoBoundFilePathRejectsSymlinkEscape(t *testing.T) {
	t.Cleanup(index.CloseAll)

	repoDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoDir, "main.go"), []byte("package main\nfunc inside() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(t.TempDir(), "test.db")
	if _, err := index.Index(repoDir, dbPath, index.Options{Workers: 1}); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(t.TempDir(), "secret.go")
	if err := os.WriteFile(target, []byte("package secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(repoDir, "leak.go")
	if err := os.Symlink(target, linkPath); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}

	_, err := repoBoundFilePath(dbPath, linkPath)
	if err == nil || !strings.Contains(err.Error(), "outside repository") {
		t.Fatalf("expected symlink escape error, got %v", err)
	}
}
