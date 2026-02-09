package cli

import (
	"fmt"
	"os"

	"github.com/nhomble/canopy/internal/canopydir"
	"github.com/nhomble/canopy/internal/patterns"
	"github.com/nhomble/canopy/internal/prompt"
	"github.com/nhomble/canopy/internal/scanner"
	"github.com/spf13/cobra"
)

var prepareOutput string

var prepareCmd = &cobra.Command{
	Use:   "prepare-analysis [directory]",
	Short: "Scan a codebase and generate an analysis prompt for an LLM",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := "."
		if len(args) > 0 {
			target = args[0]
		}

		ad, err := canopydir.Find(".")
		if err != nil {
			return err
		}

		cfg, err := ad.LoadConfig()
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Scanning %s...\n", target)
		summary, err := scanner.Scan(target, cfg.IgnorePatterns, cfg.MaxFileSizeBytes)
		if err != nil {
			return fmt.Errorf("scanning codebase: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Found %d files across %d directories\n",
			summary.Stats.TotalFiles, summary.Stats.TotalDirs)

		pats, err := patterns.LoadAll()
		if err != nil {
			return fmt.Errorf("loading patterns: %w", err)
		}

		data := prompt.PromptData{
			RepoID: cfg.RepoID,
			Tree:   summary.Tree,
			Stats: prompt.ScanStats{
				TotalFiles:       summary.Stats.TotalFiles,
				TotalDirs:        summary.Stats.TotalDirs,
				FilesByExtension: summary.Stats.FilesByExtension,
			},
			Patterns: pats,
		}

		rendered, err := prompt.RenderAnalysisPrompt(data)
		if err != nil {
			return fmt.Errorf("rendering prompt: %w", err)
		}

		if prepareOutput != "" {
			if err := os.WriteFile(prepareOutput, []byte(rendered), 0o644); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Prompt written to %s\n", prepareOutput)
		} else {
			fmt.Print(rendered)
		}

		promptPath := ad.PromptPath("analyze-root.md")
		os.WriteFile(promptPath, []byte(rendered), 0o644)

		return nil
	},
}

func init() {
	prepareCmd.Flags().StringVarP(&prepareOutput, "output", "o", "", "write prompt to file instead of stdout")
	rootCmd.AddCommand(prepareCmd)
}
