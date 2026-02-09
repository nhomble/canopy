package canopydir

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nhomble/canopy/internal/schema"
)

const DefaultDirName = ".canopy"

// CanopyDir manages the .canopy/ directory for a project.
type CanopyDir struct {
	Root string // Absolute path to the .canopy/ directory
}

// Find walks up from startDir looking for a .canopy/ directory,
// similar to how git finds .git/.
func Find(startDir string) (*CanopyDir, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	for {
		candidate := filepath.Join(dir, DefaultDirName)
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return &CanopyDir{Root: candidate}, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return nil, fmt.Errorf("no %s directory found (run 'canopy init' first)", DefaultDirName)
}

// Init creates a new .canopy/ directory structure inside parentDir.
func Init(parentDir string) (*CanopyDir, error) {
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
	ad := &CanopyDir{Root: root}
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

func (a *CanopyDir) ConfigPath() string {
	return filepath.Join(a.Root, "config.json")
}

func (a *CanopyDir) IndexPath() string {
	return filepath.Join(a.Root, "index.json")
}

func (a *CanopyDir) PromptPath(name string) string {
	return filepath.Join(a.Root, "prompts", name)
}

func (a *CanopyDir) ComponentPath(id string) string {
	return filepath.Join(a.Root, "components", id+".json")
}

// LoadConfig reads and parses the config.json file.
func (a *CanopyDir) LoadConfig() (*schema.Config, error) {
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
