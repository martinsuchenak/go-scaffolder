package diff

import (
	"strings"
	"testing"
)

func TestNewFileDiff(t *testing.T) {
	content := "line1\nline2\nline3\n"
	result := NewFileDiff("cmd/test.go", content)

	if !strings.Contains(result, "--- /dev/null") {
		t.Error("expected --- /dev/null header")
	}
	if !strings.Contains(result, "+++ cmd/test.go") {
		t.Error("expected +++ header with file path")
	}
	if !strings.Contains(result, "+line1") {
		t.Error("expected +line1")
	}
	if !strings.Contains(result, "+line2") {
		t.Error("expected +line2")
	}
	if !strings.Contains(result, "@@ -0,0 +1,3 @@") {
		t.Error("expected @@ -0,0 +1,3 @@ hunk header")
	}
}

func TestNewFileDiffEmpty(t *testing.T) {
	result := NewFileDiff("empty.go", "")
	if !strings.Contains(result, "--- /dev/null") {
		t.Error("expected --- /dev/null for empty file")
	}
}

func TestUnifiedDiff_Insertion(t *testing.T) {
	old := "line1\nline2\nline3\n"
	new_ := "line1\nline2\nnew-line\nline3\n"
	result := UnifiedDiff("a/file.go", "b/file.go", old, new_)

	if !strings.Contains(result, "--- a/file.go") {
		t.Error("expected --- header")
	}
	if !strings.Contains(result, "+++ b/file.go") {
		t.Error("expected +++ header")
	}
	if !strings.Contains(result, "+new-line") {
		t.Error("expected +new-line in diff")
	}
}

func TestUnifiedDiff_Deletion(t *testing.T) {
	old := "line1\nline2\nline3\n"
	new_ := "line1\nline3\n"
	result := UnifiedDiff("a/file.go", "b/file.go", old, new_)

	if !strings.Contains(result, "-line2") {
		t.Error("expected -line2 in diff")
	}
}

func TestUnifiedDiff_Replacement(t *testing.T) {
	old := "line1\nold-line\nline3\n"
	new_ := "line1\nnew-line\nline3\n"
	result := UnifiedDiff("a/file.go", "b/file.go", old, new_)

	if !strings.Contains(result, "-old-line") {
		t.Error("expected -old-line")
	}
	if !strings.Contains(result, "+new-line") {
		t.Error("expected +new-line")
	}
}

func TestUnifiedDiff_NoChange(t *testing.T) {
	old := "line1\nline2\n"
	result := UnifiedDiff("a/file.go", "b/file.go", old, old)

	if result != "" {
		t.Errorf("expected empty diff for identical content, got:\n%s", result)
	}
}
