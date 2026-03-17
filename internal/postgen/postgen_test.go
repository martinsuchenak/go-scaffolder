package postgen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunGoModTidy(t *testing.T) {
	dir := t.TempDir()

	gomod := `module testmod

go 1.26
`
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0644); err != nil {
		t.Fatal(err)
	}

	mainGo := `package main

func main() {}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := RunGoModTidy(dir)
	if err != nil {
		t.Fatalf("RunGoModTidy error: %v, output: %s", err, output)
	}
}

func TestRunGoModTidyInvalidDir(t *testing.T) {
	_, err := RunGoModTidy("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for invalid directory")
	}
}
