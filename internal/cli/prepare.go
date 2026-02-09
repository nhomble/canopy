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

		// Find .canopy directory
		ad, err := canopydir.Find(".")
		if err != nil {
			return err
		}

		// Load config
		cfg, err := ad.LoadConfig()
		if err != nil {
			return err
		}

		// Scan the codebase
		fmt.Fprintf(os.Stderr, "Scanning %s...\n", target)
		summary, err := scanner.Scan(target, cfg.IgnorePatterns, cfg.MaxFileSizeBytes, cfg.SampleFileCount)
		if err != nil {
			return fmt.Errorf("scanning codebase: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Found %d files across %d directories\n",
			summary.Stats.TotalFiles, summary.Stats.TotalDirs)

		// Load patterns
		pats, err := patterns.LoadAll()
		if err != nil {
			return fmt.Errorf("loading patterns: %w", err)
		}

		// Convert scanner types to prompt types
		data := prompt.PromptData{
			RepoID:   cfg.RepoID,
			Tree:     summary.Tree,
			Stats:    convertStats(summary.Stats),
			TechStack: convertTechStack(summary.TechStack),
			Signatures: convertSignatures(summary.Signatures),
			Patterns: pats,
		}

		// Render prompt
		rendered, err := prompt.RenderAnalysisPrompt(data)
		if err != nil {
			return fmt.Errorf("rendering prompt: %w", err)
		}

		// Write output
		if prepareOutput != "" {
			if err := os.WriteFile(prepareOutput, []byte(rendered), 0o644); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Prompt written to %s\n", prepareOutput)
		} else {
			fmt.Print(rendered)
		}

		// Also save to .canopy/prompts/
		promptPath := ad.PromptPath("analyze-root.md")
		os.WriteFile(promptPath, []byte(rendered), 0o644)

		return nil
	},
}

func init() {
	prepareCmd.Flags().StringVarP(&prepareOutput, "output", "o", "", "write prompt to file instead of stdout")
	rootCmd.AddCommand(prepareCmd)
}

func convertStats(s scanner.ScanStats) prompt.ScanStats {
	return prompt.ScanStats{
		TotalFiles:       s.TotalFiles,
		TotalDirs:        s.TotalDirs,
		FilesByExtension: s.FilesByExtension,
	}
}

func convertTechStack(ts []scanner.TechHint) []prompt.TechHint {
	result := make([]prompt.TechHint, len(ts))
	for i, t := range ts {
		result[i] = prompt.TechHint{Source: t.Source, Name: t.Name, Type: t.Type}
	}
	return result
}

func convertSignatures(sigs []scanner.FileSignatures) []prompt.FileSignatures {
	result := make([]prompt.FileSignatures, len(sigs))
	for i, s := range sigs {
		imports := make([]prompt.ImportStatement, len(s.Imports))
		for j, imp := range s.Imports {
			imports[j] = prompt.ImportStatement{Raw: imp.Raw, Source: imp.Source}
		}
		signatures := make([]prompt.Signature, len(s.Signatures))
		for j, sig := range s.Signatures {
			signatures[j] = prompt.Signature{Kind: sig.Kind, Name: sig.Name, Line: sig.Line, Raw: sig.Raw}
		}
		result[i] = prompt.FileSignatures{
			Path:       s.Path,
			RelPath:    s.RelPath,
			Language:   s.Language,
			Imports:    imports,
			Signatures: signatures,
			Decorators: s.Decorators,
		}
	}
	return result
}
