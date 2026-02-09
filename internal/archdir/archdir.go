package archdir

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nhomble/arch-index/internal/schema"
)

const DefaultDirName = ".arch"

// ArchDir manages the .arch/ directory for a project.
type ArchDir struct {
	Root string // Absolute path to the .arch/ directory
}

// Find walks up from startDir looking for a .arch/ directory,
// similar to how git finds .git/.
func Find(startDir string) (*ArchDir, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	for {
		candidate := filepath.Join(dir, DefaultDirName)
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return &ArchDir{Root: candidate}, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return nil, fmt.Errorf("no %s directory found (run 'arch-index init' first)", DefaultDirName)
}

// Init creates a new .arch/ directory structure inside parentDir.
func Init(parentDir string) (*ArchDir, error) {
	absParent, err := filepath.Abs(parentDir)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	root := filepath.Join(absParent, DefaultDirName)

	// Check if already exists
	if info, err := os.Stat(root); err == nil && info.IsDir() {
		return nil, fmt.Errorf("%s already exists", root)
	}

	// Create directory structure
	dirs := []string{
		root,
		filepath.Join(root, "components"),
		filepath.Join(root, "prompts"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, fmt.Errorf("creating directory %s: %w", d, err)
		}
	}

	// Infer repo ID from directory name
	repoID := filepath.Base(absParent)
	cfg := schema.DefaultConfig(repoID)

	// Write config.json
	ad := &ArchDir{Root: root}
	cfgData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(ad.ConfigPath(), cfgData, 0o644); err != nil {
		return nil, fmt.Errorf("writing config: %w", err)
	}

	return ad, nil
}

// Path helpers

func (a *ArchDir) ConfigPath() string {
	return filepath.Join(a.Root, "config.json")
}

func (a *ArchDir) IndexPath() string {
	return filepath.Join(a.Root, "index.json")
}

func (a *ArchDir) PromptPath(name string) string {
	return filepath.Join(a.Root, "prompts", name)
}

func (a *ArchDir) ComponentPath(id string) string {
	return filepath.Join(a.Root, "components", id+".json")
}

// LoadConfig reads and parses the config.json file.
func (a *ArchDir) LoadConfig() (*schema.Config, error) {
	data, err := os.ReadFile(a.ConfigPath())
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg schema.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}
