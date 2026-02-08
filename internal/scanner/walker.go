package scanner

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// FileInfo holds metadata about a single discovered file.
type FileInfo struct {
	Path      string // Absolute path
	RelPath   string // Relative to scan root
	Extension string // e.g., ".ts", ".go", ".java"
	SizeBytes int64
}

// WalkResult contains all files discovered during a directory walk.
type WalkResult struct {
	Root  string
	Files []FileInfo
}

// DefaultMaxFileSize is the default maximum file size to include (1MB).
const DefaultMaxFileSize int64 = 1 << 20

// binaryExtensions contains file extensions considered binary.
var binaryExtensions = map[string]bool{
	".exe": true, ".bin": true, ".o": true, ".so": true, ".dll": true,
	".class": true, ".jar": true, ".png": true, ".jpg": true, ".jpeg": true,
	".gif": true, ".ico": true, ".woff": true, ".woff2": true, ".ttf": true,
	".eot": true, ".pdf": true, ".zip": true, ".tar": true, ".gz": true,
	".bz2": true, ".7z": true, ".rar": true, ".mp3": true, ".mp4": true,
	".avi": true, ".mov": true, ".webm": true, ".webp": true, ".svg": true,
	".bmp": true, ".tif": true, ".tiff": true, ".pyc": true, ".pyo": true,
	".a": true, ".lib": true, ".dylib": true,
}

// Walk traverses the directory tree rooted at root and returns all source
// files that pass the ignore, binary, and size filters.
func Walk(root string, ignorePatterns []string, maxFileSize int64) (*WalkResult, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if maxFileSize <= 0 {
		maxFileSize = DefaultMaxFileSize
	}
	ignoreSet := make(map[string]bool, len(ignorePatterns))
	for _, p := range ignorePatterns {
		ignoreSet[strings.ToLower(p)] = true
	}
	result := &WalkResult{Root: absRoot, Files: make([]FileInfo, 0, 256)}
	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if ignoreSet[strings.ToLower(name)] {
				return filepath.SkipDir
			}
			if strings.HasPrefix(name, ".") && path != absRoot {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(name, ".") {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(name))
		if binaryExtensions[ext] {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.Size() > maxFileSize {
			return nil
		}
		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			relPath = path
		}
		result.Files = append(result.Files, FileInfo{
			Path: path, RelPath: relPath, Extension: ext, SizeBytes: info.Size(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
