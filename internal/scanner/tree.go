package scanner

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// DefaultMaxDepth is the default maximum depth for tree rendering.
const DefaultMaxDepth = 4

// DefaultMaxFilesPerDir is the default max files shown per directory.
const DefaultMaxFilesPerDir = 10

// dirNode represents a node in the directory tree.
type dirNode struct {
	name     string
	children map[string]*dirNode
	files    []string
	// totalFiles counts all files recursively under this node.
	totalFiles int
}

func newDirNode(name string) *dirNode {
	return &dirNode{
		name:     name,
		children: make(map[string]*dirNode),
	}
}

// RenderTree produces a compact directory tree string from a list of FileInfo.
// maxDepth controls how deep into the tree to render (0 means use default).
func RenderTree(files []FileInfo, maxDepth int) string {
	if maxDepth <= 0 {
		maxDepth = DefaultMaxDepth
	}

	root := newDirNode("")

	// Build the tree structure.
	for _, f := range files {
		parts := strings.Split(filepath.ToSlash(f.RelPath), "/")
		node := root
		// Navigate/create directory nodes.
		for i := 0; i < len(parts)-1; i++ {
			child, ok := node.children[parts[i]]
			if !ok {
				child = newDirNode(parts[i])
				node.children[parts[i]] = child
			}
			node = child
		}
		// Add the file to the leaf directory.
		node.files = append(node.files, parts[len(parts)-1])
	}

	// Count total files recursively.
	countFiles(root)

	var sb strings.Builder
	renderNode(&sb, root, 0, maxDepth, "")
	return sb.String()
}

// countFiles recursively counts all files under a node.
func countFiles(n *dirNode) int {
	n.totalFiles = len(n.files)
	for _, child := range n.children {
		n.totalFiles += countFiles(child)
	}
	return n.totalFiles
}

// renderNode writes the tree representation of a node into the string builder.
func renderNode(sb *strings.Builder, n *dirNode, depth, maxDepth int, prefix string) {
	if depth >= maxDepth {
		// At max depth, just show summary if there are contents.
		if n.totalFiles > 0 {
			fmt.Fprintf(sb, "%s(%d files)\n", prefix, n.totalFiles)
		}
		return
	}

	// Sort child directory names for deterministic output.
	dirNames := make([]string, 0, len(n.children))
	for name := range n.children {
		dirNames = append(dirNames, name)
	}
	sort.Strings(dirNames)

	// Collapse single-child directory chains.
	// If this node has exactly one child directory and no files, collapse.
	for len(dirNames) == 1 && len(n.files) == 0 && depth > 0 {
		childName := dirNames[0]
		child := n.children[childName]
		// Write collapsed path segment.
		prefix = prefix + childName + "/"
		n = child
		dirNames = make([]string, 0, len(n.children))
		for name := range n.children {
			dirNames = append(dirNames, name)
		}
		sort.Strings(dirNames)
		depth++ // Collapsing still counts toward depth budget.
		if depth >= maxDepth {
			if n.totalFiles > 0 {
				fmt.Fprintf(sb, "%s(%d files)\n", prefix, n.totalFiles)
			}
			return
		}
	}

	indent := strings.Repeat("  ", depth)

	// Render child directories.
	for _, dirName := range dirNames {
		child := n.children[dirName]
		fmt.Fprintf(sb, "%s%s/\n", indent, prefix+dirName)
		renderNode(sb, child, depth+1, maxDepth, "")
	}

	// Render files.
	sort.Strings(n.files)
	shown := n.files
	remaining := 0
	if len(shown) > DefaultMaxFilesPerDir {
		remaining = len(shown) - DefaultMaxFilesPerDir
		shown = shown[:DefaultMaxFilesPerDir]
	}
	for _, f := range shown {
		fmt.Fprintf(sb, "%s%s%s\n", indent, prefix, f)
	}
	if remaining > 0 {
		fmt.Fprintf(sb, "%s%s... and %d more\n", indent, prefix, remaining)
	}
	// Clear prefix after first use at this level.
}
