package patcher

import (
	"fmt"
	"os"
	"strings"
)

type Patch struct {
	File        string
	Marker      string
	InsertAbove bool
	Content     string
	Description string
	Replace     *ReplaceBlock
}

type ReplaceBlock struct {
	StartMarker string
	EndMarker   string
	Content     string
}

type PatchResult struct {
	File        string
	Applied     bool
	Description string
	Content     string
}

func ApplyPatches(baseDir string, patches []Patch) []PatchResult {
	var results []PatchResult

	for _, p := range patches {
		path := baseDir + "/" + p.File
		displayContent := p.Content
		if p.Replace != nil {
			displayContent = p.Replace.Content
		}
		result := PatchResult{
			File:        p.File,
			Description: p.Description,
			Content:     displayContent,
		}

		data, err := os.ReadFile(path)
		if err != nil {
			result.Applied = false
			results = append(results, result)
			continue
		}

		content := string(data)

		if p.Replace != nil {
			startIdx := strings.Index(content, p.Replace.StartMarker)
			endIdx := strings.Index(content, p.Replace.EndMarker)
			if startIdx == -1 || endIdx == -1 {
				result.Applied = false
				results = append(results, result)
				continue
			}
			newContent := content[:startIdx] + p.Replace.Content + content[endIdx+len(p.Replace.EndMarker):]
			if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
				result.Applied = false
				results = append(results, result)
				continue
			}
			result.Applied = true
			results = append(results, result)
			continue
		}

		markerLine := p.Marker

		idx := strings.Index(content, markerLine)
		if idx == -1 {
			result.Applied = false
			results = append(results, result)
			continue
		}

		var newContent string
		if p.InsertAbove {
			newContent = content[:idx] + p.Content + "\n" + content[idx:]
		} else {
			endOfMarker := idx + len(markerLine)
			newContent = content[:endOfMarker] + "\n" + p.Content + content[endOfMarker:]
		}

		if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
			result.Applied = false
			results = append(results, result)
			continue
		}

		result.Applied = true
		results = append(results, result)
	}

	return results
}

func ReportResults(results []PatchResult) {
	var failed []PatchResult
	var succeeded []PatchResult

	for _, r := range results {
		if r.Applied {
			succeeded = append(succeeded, r)
		} else {
			failed = append(failed, r)
		}
	}

	for _, r := range succeeded {
		fmt.Printf("  updated: %s (%s)\n", r.File, r.Description)
	}

	if len(failed) > 0 {
		fmt.Println("\nThe following patches could not be applied automatically.")
		fmt.Println("Please add the following code manually:")
		for _, r := range failed {
			fmt.Printf("\n--- %s (%s) ---\n", r.File, r.Description)
			fmt.Println(r.Content)
		}
	}
}
