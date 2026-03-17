package diff

import (
	"fmt"
	"strings"
)

func UnifiedDiff(oldPath, newPath, oldContent, newContent string) string {
	oldLines := splitLines(oldContent)
	newLines := splitLines(newContent)

	hunks := computeHunks(oldLines, newLines)
	if len(hunks) == 0 {
		return ""
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "--- %s\n", oldPath)
	fmt.Fprintf(&sb, "+++ %s\n", newPath)

	for _, h := range hunks {
		fmt.Fprintf(&sb, "@@ -%d,%d +%d,%d @@\n", h.oldStart+1, h.oldCount, h.newStart+1, h.newCount)
		for _, l := range h.lines {
			sb.WriteString(l)
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

func NewFileDiff(path string, content string) string {
	lines := splitLines(content)
	var sb strings.Builder
	fmt.Fprintf(&sb, "--- /dev/null\n")
	fmt.Fprintf(&sb, "+++ %s\n", path)
	fmt.Fprintf(&sb, "@@ -0,0 +1,%d @@\n", len(lines))
	for _, l := range lines {
		fmt.Fprintf(&sb, "+%s\n", l)
	}
	return sb.String()
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.TrimRight(s, "\n")
	return strings.Split(s, "\n")
}

type hunk struct {
	oldStart int
	oldCount int
	newStart int
	newCount int
	lines    []string
}

func computeHunks(oldLines, newLines []string) []hunk {
	lcs := lcsTable(oldLines, newLines)
	edits := backtrack(lcs, oldLines, newLines)

	const contextLines = 3
	var hunks []hunk
	var current *hunk

	oi, ni := 0, 0
	for _, e := range edits {
		switch e {
		case editKeep:
			if current != nil {
				current.lines = append(current.lines, " "+oldLines[oi])
				current.oldCount++
				current.newCount++
			}
			oi++
			ni++
		case editDelete:
			if current == nil {
				start := oi - contextLines
				if start < 0 {
					start = 0
				}
				current = &hunk{oldStart: start, newStart: ni - (oi - start)}
				if current.newStart < 0 {
					current.newStart = 0
				}
				for i := start; i < oi; i++ {
					current.lines = append(current.lines, " "+oldLines[i])
					current.oldCount++
					current.newCount++
				}
			}
			current.lines = append(current.lines, "-"+oldLines[oi])
			current.oldCount++
			oi++
		case editInsert:
			if current == nil {
				start := oi - contextLines
				if start < 0 {
					start = 0
				}
				current = &hunk{oldStart: start, newStart: ni - (oi - start)}
				if current.newStart < 0 {
					current.newStart = 0
				}
				for i := start; i < oi; i++ {
					current.lines = append(current.lines, " "+oldLines[i])
					current.oldCount++
					current.newCount++
				}
			}
			current.lines = append(current.lines, "+"+newLines[ni])
			current.newCount++
			ni++
		}
	}

	if current != nil {
		end := oi + contextLines
		if end > len(oldLines) {
			end = len(oldLines)
		}
		for i := oi; i < end; i++ {
			current.lines = append(current.lines, " "+oldLines[i])
			current.oldCount++
			current.newCount++
		}
		hunks = append(hunks, *current)
	}

	return hunks
}

type editOp int

const (
	editKeep editOp = iota
	editDelete
	editInsert
)

func lcsTable(a, b []string) [][]int {
	m, n := len(a), len(b)
	table := make([][]int, m+1)
	for i := range table {
		table[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				table[i][j] = table[i-1][j-1] + 1
			} else if table[i-1][j] >= table[i][j-1] {
				table[i][j] = table[i-1][j]
			} else {
				table[i][j] = table[i][j-1]
			}
		}
	}
	return table
}

func backtrack(table [][]int, a, b []string) []editOp {
	var ops []editOp
	i, j := len(a), len(b)
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && a[i-1] == b[j-1] {
			ops = append(ops, editKeep)
			i--
			j--
		} else if j > 0 && (i == 0 || table[i][j-1] >= table[i-1][j]) {
			ops = append(ops, editInsert)
			j--
		} else {
			ops = append(ops, editDelete)
			i--
		}
	}
	for l, r := 0, len(ops)-1; l < r; l, r = l+1, r-1 {
		ops[l], ops[r] = ops[r], ops[l]
	}
	return ops
}
