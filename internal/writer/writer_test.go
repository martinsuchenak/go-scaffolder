package writer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAll(t *testing.T) {
	dir := t.TempDir()

	files := map[string][]byte{
		"main.go":           []byte("package main\n"),
		"cmd/serve.go":      []byte("package cmd\n"),
		"internal/db/db.go": []byte("package db\n"),
	}

	if err := WriteAll(dir, files); err != nil {
		t.Fatalf("WriteAll error: %v", err)
	}

	for relPath, expected := range files {
		fullPath := filepath.Join(dir, relPath)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatalf("reading %s: %v", fullPath, err)
		}
		if string(data) != string(expected) {
			t.Errorf("file %s: got %q, want %q", relPath, data, expected)
		}
	}
}

func TestWriteAllCreatesDirectories(t *testing.T) {
	dir := t.TempDir()

	files := map[string][]byte{
		"a/b/c/d.txt": []byte("deep file"),
	}

	if err := WriteAll(dir, files); err != nil {
		t.Fatalf("WriteAll error: %v", err)
	}

	fullPath := filepath.Join(dir, "a/b/c/d.txt")
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Fatalf("expected file %s to exist", fullPath)
	}
}

func TestWriteAllInvalidDir(t *testing.T) {
	files := map[string][]byte{
		"test.txt": []byte("test"),
	}

	err := WriteAll("/nonexistent/path/that/cannot/exist", files)
	if err == nil {
		t.Fatal("expected error for invalid directory")
	}
}
