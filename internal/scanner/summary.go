package scanner

import (
	"path/filepath"
)

// ScanStats holds aggregate statistics from the scan.
type ScanStats struct {
	TotalFiles       int
	TotalDirs        int
	FilesByExtension map[string]int
}

// CodebaseSummary is the complete output of scanning a codebase.
type CodebaseSummary struct {
	RepoID string
	Root   string
	Tree   string
	Stats  ScanStats
}

// Scan walks the codebase and produces a summary with tree and stats.
func Scan(root string, ignorePatterns []string, maxFileSize int64) (*CodebaseSummary, error) {
	walkResult, err := Walk(root, ignorePatterns, maxFileSize)
	if err != nil {
		return nil, err
	}

	stats := computeStats(walkResult.Files)
	tree := RenderTree(walkResult.Files, 0)
	repoID := filepath.Base(walkResult.Root)

	return &CodebaseSummary{
		RepoID: repoID,
		Root:   walkResult.Root,
		Tree:   tree,
		Stats:  stats,
	}, nil
}

func computeStats(files []FileInfo) ScanStats {
	byExt := make(map[string]int)
	dirs := make(map[string]bool)
	for _, f := range files {
		byExt[f.Extension]++
		dirs[filepath.Dir(f.RelPath)] = true
	}
	return ScanStats{
		TotalFiles:       len(files),
		TotalDirs:        len(dirs),
		FilesByExtension: byExt,
	}
}
