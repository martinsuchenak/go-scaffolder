package patcher

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestApplyPatches_MarkerFound(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.go")
	if err := os.WriteFile(file, []byte("line1\n// go-scaffolder:marker\nline3\n"), 0644); err != nil {
		t.Fatal(err)
	}

	patches := []Patch{
		{
			File:        "test.go",
			Marker:      "// go-scaffolder:marker",
			InsertAbove: false,
			Content:     "inserted-line",
			Description: "test insert below",
		},
	}

	results := ApplyPatches(dir, patches)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Applied {
		t.Fatal("expected patch to be applied")
	}

	data, _ := os.ReadFile(file)
	content := string(data)
	if !strings.Contains(content, "inserted-line") {
		t.Fatalf("expected inserted line in file, got:\n%s", content)
	}
	if !strings.Contains(content, "// go-scaffolder:marker\ninserted-line") {
		t.Fatalf("expected insert below marker, got:\n%s", content)
	}
}

func TestApplyPatches_InsertAbove(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.go")
	if err := os.WriteFile(file, []byte("line1\n// go-scaffolder:marker\nline3\n"), 0644); err != nil {
		t.Fatal(err)
	}

	patches := []Patch{
		{
			File:        "test.go",
			Marker:      "// go-scaffolder:marker",
			InsertAbove: true,
			Content:     "inserted-above",
			Description: "test insert above",
		},
	}

	results := ApplyPatches(dir, patches)
	if !results[0].Applied {
		t.Fatal("expected patch to be applied")
	}

	data, _ := os.ReadFile(file)
	content := string(data)
	if !strings.Contains(content, "inserted-above\n// go-scaffolder:marker") {
		t.Fatalf("expected insert above marker, got:\n%s", content)
	}
}

func TestApplyPatches_MarkerNotFound(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.go")
	if err := os.WriteFile(file, []byte("line1\nline2\nline3\n"), 0644); err != nil {
		t.Fatal(err)
	}

	patches := []Patch{
		{
			File:        "test.go",
			Marker:      "// go-scaffolder:missing",
			Content:     "will-not-insert",
			Description: "test missing marker",
		},
	}

	results := ApplyPatches(dir, patches)
	if results[0].Applied {
		t.Fatal("expected patch to NOT be applied")
	}
}

func TestApplyPatches_FileNotFound(t *testing.T) {
	dir := t.TempDir()

	patches := []Patch{
		{
			File:        "nonexistent.go",
			Marker:      "// go-scaffolder:marker",
			Content:     "will-not-insert",
			Description: "test missing file",
		},
	}

	results := ApplyPatches(dir, patches)
	if results[0].Applied {
		t.Fatal("expected patch to NOT be applied for missing file")
	}
}

func TestApplyPatches_MultiplePatches(t *testing.T) {
	dir := t.TempDir()

	file1 := filepath.Join(dir, "a.go")
	if err := os.WriteFile(file1, []byte("// go-scaffolder:m1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	file2 := filepath.Join(dir, "b.go")
	if err := os.WriteFile(file2, []byte("no marker here\n"), 0644); err != nil {
		t.Fatal(err)
	}

	patches := []Patch{
		{File: "a.go", Marker: "// go-scaffolder:m1", Content: "patched-a", Description: "patch a"},
		{File: "b.go", Marker: "// go-scaffolder:m2", Content: "patched-b", Description: "patch b"},
	}

	results := ApplyPatches(dir, patches)
	if !results[0].Applied {
		t.Error("expected first patch to be applied")
	}
	if results[1].Applied {
		t.Error("expected second patch to NOT be applied")
	}
}

func TestApplyPatches_ReplaceBlock(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.go")
	if err := os.WriteFile(file, []byte("before\n// start-marker\nold content\n// end-marker\nafter\n"), 0644); err != nil {
		t.Fatal(err)
	}

	patches := []Patch{
		{
			File:        "test.go",
			Description: "replace block",
			Replace: &ReplaceBlock{
				StartMarker: "// start-marker",
				EndMarker:   "// end-marker",
				Content:     "// start-marker\nnew content\n// end-marker",
			},
		},
	}

	results := ApplyPatches(dir, patches)
	if !results[0].Applied {
		t.Fatal("expected replace patch to be applied")
	}

	data, _ := os.ReadFile(file)
	content := string(data)
	if strings.Contains(content, "old content") {
		t.Fatal("old content should have been replaced")
	}
	if !strings.Contains(content, "new content") {
		t.Fatalf("new content should be present, got:\n%s", content)
	}
	if !strings.Contains(content, "before") || !strings.Contains(content, "after") {
		t.Fatal("surrounding content should be preserved")
	}
}

func TestComputePatches_InsertBelow(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.go")
	if err := os.WriteFile(file, []byte("line1\n// marker\nline3\n"), 0644); err != nil {
		t.Fatal(err)
	}

	patches := []Patch{
		{File: "test.go", Marker: "// marker", Content: "inserted", Description: "test"},
	}

	results := ComputePatches(dir, patches)
	if !results[0].Applied {
		t.Fatal("expected patch to be computable")
	}
	if results[0].OldContent != "line1\n// marker\nline3\n" {
		t.Fatalf("unexpected old content: %q", results[0].OldContent)
	}
	if !strings.Contains(results[0].NewContent, "inserted") {
		t.Fatal("expected new content to contain inserted text")
	}

	data, _ := os.ReadFile(file)
	if strings.Contains(string(data), "inserted") {
		t.Fatal("ComputePatches should NOT write to disk")
	}
}

func TestComputePatches_ReplaceBlock(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.go")
	if err := os.WriteFile(file, []byte("before\n// start\nold\n// end\nafter\n"), 0644); err != nil {
		t.Fatal(err)
	}

	patches := []Patch{
		{
			File:        "test.go",
			Description: "replace",
			Replace: &ReplaceBlock{
				StartMarker: "// start",
				EndMarker:   "// end",
				Content:     "// start\nnew\n// end",
			},
		},
	}

	results := ComputePatches(dir, patches)
	if !results[0].Applied {
		t.Fatal("expected replace patch to be computable")
	}
	if strings.Contains(results[0].NewContent, "old") {
		t.Fatal("expected old content to be replaced")
	}
	if !strings.Contains(results[0].NewContent, "new") {
		t.Fatal("expected new content in result")
	}
}

func TestComputePatches_MissingMarker(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.go")
	writeFile(t, file, []byte("no marker\n"))

	patches := []Patch{
		{File: "test.go", Marker: "// missing", Content: "x", Description: "test"},
	}

	results := ComputePatches(dir, patches)
	if results[0].Applied {
		t.Fatal("expected patch to NOT be computable when marker missing")
	}
}

func TestApplyPatches_ReplaceBlockMissingMarker(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.go")
	writeFile(t, file, []byte("no markers here\n"))

	patches := []Patch{
		{
			File:        "test.go",
			Description: "replace block missing",
			Replace: &ReplaceBlock{
				StartMarker: "// start-marker",
				EndMarker:   "// end-marker",
				Content:     "new content",
			},
		},
	}

	results := ApplyPatches(dir, patches)
	if results[0].Applied {
		t.Fatal("expected replace patch to NOT be applied when markers missing")
	}
}
