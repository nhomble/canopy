package scanner

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// TechHint describes a technology detected from manifest files.
type TechHint struct {
	Source string // "pom.xml", "package.json", "go.mod"
	Name   string // "spring-boot", "express", etc.
	Type   string // "framework", "library", "build-tool"
}

// ScanStats holds aggregate statistics from the scan.
type ScanStats struct {
	TotalFiles       int
	TotalDirs        int
	FilesByExtension map[string]int
}

// CodebaseSummary is the complete output of scanning a codebase.
type CodebaseSummary struct {
	RepoID     string
	Root       string
	Tree       string
	Stats      ScanStats
	TechStack  []TechHint
	Signatures []FileSignatures
}

// importantPatterns are filename substrings that indicate architecturally
// significant files, used to prioritize sampling.
var importantPatterns = []string{
	"controller", "service", "repository", "handler", "model",
	"entity", "port", "usecase", "adapter", "middleware",
	"factory", "gateway", "client", "provider",
}

// entryPoints are filenames that typically represent application entry points.
var entryPoints = map[string]bool{
	"main.go": true, "main.java": true, "main.py": true, "main.rs": true,
	"index.ts": true, "index.js": true, "app.ts": true, "app.js": true,
	"app.py": true, "application.java": true, "server.go": true,
	"server.ts": true, "server.js": true,
}

// manifestFiles are filenames that indicate project tech stack.
var manifestFiles = map[string]bool{
	"package.json": true, "go.mod": true, "cargo.toml": true,
	"pom.xml": true, "build.gradle": true, "build.gradle.kts": true,
	"requirements.txt": true, "pyproject.toml": true, "gemfile": true,
	"composer.json": true, "mix.exs": true,
}

// Scan walks the codebase, extracts structural info, and produces a summary.
func Scan(root string, ignorePatterns []string, maxFileSize int64, sampleCount int) (*CodebaseSummary, error) {
	if sampleCount <= 0 {
		sampleCount = 5
	}

	// Walk the directory
	walkResult, err := Walk(root, ignorePatterns, maxFileSize)
	if err != nil {
		return nil, err
	}

	// Build stats
	stats := computeStats(walkResult.Files)

	// Render tree
	tree := RenderTree(walkResult.Files, 0)

	// Detect tech stack
	techStack := detectTechStack(walkResult)

	// Sample files and extract signatures
	sampled := sampleFiles(walkResult.Files, sampleCount)
	var signatures []FileSignatures
	for _, f := range sampled {
		sig, err := ExtractSignatures(f.Path, f.RelPath)
		if err != nil || sig == nil {
			continue
		}
		signatures = append(signatures, *sig)
	}

	repoID := filepath.Base(walkResult.Root)

	return &CodebaseSummary{
		RepoID:     repoID,
		Root:       walkResult.Root,
		Tree:       tree,
		Stats:      stats,
		TechStack:  techStack,
		Signatures: signatures,
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

func sampleFiles(files []FileInfo, perDir int) []FileInfo {
	// Group files by directory
	byDir := make(map[string][]FileInfo)
	for _, f := range files {
		dir := filepath.Dir(f.RelPath)
		byDir[dir] = append(byDir[dir], f)
	}

	var sampled []FileInfo

	for _, dirFiles := range byDir {
		// Always include manifest files
		var manifests, important, rest []FileInfo
		for _, f := range dirFiles {
			name := strings.ToLower(filepath.Base(f.RelPath))
			if manifestFiles[name] {
				manifests = append(manifests, f)
			} else if isImportant(name) {
				important = append(important, f)
			} else {
				rest = append(rest, f)
			}
		}

		sampled = append(sampled, manifests...)

		// Sort important files by size descending (larger = more content)
		sort.Slice(important, func(i, j int) bool {
			return important[i].SizeBytes > important[j].SizeBytes
		})

		remaining := perDir
		for _, f := range important {
			if remaining <= 0 {
				break
			}
			sampled = append(sampled, f)
			remaining--
		}
		for _, f := range rest {
			if remaining <= 0 {
				break
			}
			sampled = append(sampled, f)
			remaining--
		}
	}

	return sampled
}

func isImportant(name string) bool {
	if entryPoints[name] {
		return true
	}
	lower := strings.ToLower(name)
	for _, p := range importantPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func detectTechStack(wr *WalkResult) []TechHint {
	var hints []TechHint
	for _, f := range wr.Files {
		name := strings.ToLower(filepath.Base(f.RelPath))
		switch name {
		case "package.json":
			hints = append(hints, detectFromPackageJSON(f.Path)...)
		case "pom.xml":
			hints = append(hints, detectFromPomXML(f.Path)...)
		case "go.mod":
			hints = append(hints, detectFromGoMod(f.Path)...)
		case "cargo.toml":
			hints = append(hints, TechHint{Source: "Cargo.toml", Name: "rust", Type: "language"})
		case "requirements.txt", "pyproject.toml":
			hints = append(hints, TechHint{Source: name, Name: "python", Type: "language"})
		case "gemfile":
			hints = append(hints, TechHint{Source: "Gemfile", Name: "ruby", Type: "language"})
		case "build.gradle", "build.gradle.kts":
			hints = append(hints, TechHint{Source: name, Name: "gradle", Type: "build-tool"})
		}
	}
	return hints
}

func detectFromPackageJSON(path string) []TechHint {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}

	var hints []TechHint
	known := map[string]string{
		"express": "framework", "fastify": "framework", "koa": "framework",
		"next": "framework", "react": "framework", "vue": "framework",
		"angular": "framework", "nestjs": "framework", "@nestjs/core": "framework",
		"typeorm": "library", "prisma": "library", "mongoose": "library",
		"sequelize": "library",
	}
	for dep := range pkg.Dependencies {
		if typ, ok := known[dep]; ok {
			hints = append(hints, TechHint{Source: "package.json", Name: dep, Type: typ})
		}
	}
	if len(pkg.Dependencies) > 0 || len(pkg.DevDependencies) > 0 {
		hints = append(hints, TechHint{Source: "package.json", Name: "node.js", Type: "runtime"})
	}
	return hints
}

func detectFromPomXML(path string) []TechHint {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	hints := []TechHint{{Source: "pom.xml", Name: "maven", Type: "build-tool"}}

	// Simple XML parsing for dependencies
	var pom struct {
		Dependencies struct {
			Dependency []struct {
				GroupID    string `xml:"groupId"`
				ArtifactID string `xml:"artifactId"`
			} `xml:"dependency"`
		} `xml:"dependencies"`
		Parent struct {
			GroupID    string `xml:"groupId"`
			ArtifactID string `xml:"artifactId"`
		} `xml:"parent"`
	}
	if xml.Unmarshal(data, &pom) == nil {
		if strings.Contains(pom.Parent.GroupID, "springframework") ||
			strings.Contains(pom.Parent.ArtifactID, "spring-boot") {
			hints = append(hints, TechHint{Source: "pom.xml", Name: "spring-boot", Type: "framework"})
		}
		for _, dep := range pom.Dependencies.Dependency {
			if strings.Contains(dep.GroupID, "springframework") {
				hints = append(hints, TechHint{Source: "pom.xml", Name: "spring-" + dep.ArtifactID, Type: "framework"})
			}
			if strings.Contains(dep.GroupID, "mongodb") {
				hints = append(hints, TechHint{Source: "pom.xml", Name: "mongodb", Type: "database"})
			}
			if strings.Contains(dep.GroupID, "postgresql") || strings.Contains(dep.ArtifactID, "postgresql") {
				hints = append(hints, TechHint{Source: "pom.xml", Name: "postgresql", Type: "database"})
			}
		}
	}

	return hints
}

func detectFromGoMod(path string) []TechHint {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	hints := []TechHint{{Source: "go.mod", Name: "go", Type: "language"}}
	content := string(data)
	known := map[string]string{
		"github.com/gin-gonic/gin":    "framework",
		"github.com/labstack/echo":    "framework",
		"github.com/gofiber/fiber":    "framework",
		"github.com/gorilla/mux":      "library",
		"gorm.io/gorm":                "library",
		"github.com/spf13/cobra":      "library",
	}
	for mod, typ := range known {
		if strings.Contains(content, mod) {
			name := filepath.Base(mod)
			hints = append(hints, TechHint{Source: "go.mod", Name: name, Type: typ})
		}
	}
	return hints
}
