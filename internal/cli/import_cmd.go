package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/nhomble/arch-index/internal/archdir"
	"github.com/nhomble/arch-index/internal/schema"
	"github.com/spf13/cobra"
)

var importForce bool

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Validate and import LLM analysis output into .arch/index.json",
	Long: `Import reads JSON output from an LLM analysis, validates it against
the expected schema, and saves it to .arch/index.json.

The input can be a file path or piped via stdin. The tool handles
messy LLM output: markdown code fences, surrounding commentary,
and trailing commas are automatically cleaned up.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read input
		var input []byte
		var err error

		if len(args) > 0 {
			input, err = os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("reading file: %w", err)
			}
		} else {
			input, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}
		}

		if len(input) == 0 {
			return fmt.Errorf("empty input")
		}

		// Extract JSON from potentially messy input
		jsonData, err := schema.ExtractJSON(string(input))
		if err != nil {
			return fmt.Errorf("extracting JSON: %w", err)
		}

		// Parse into ArchIndex
		var idx schema.ArchIndex
		if err := json.Unmarshal(jsonData, &idx); err != nil {
			return fmt.Errorf("parsing JSON: %w", err)
		}

		// Validate
		result := schema.ValidateIndex(&idx)
		if !result.Valid {
			fmt.Fprint(os.Stderr, result.FormatResult())
			return fmt.Errorf("validation failed with %d errors", len(result.Errors))
		}

		if len(result.Warnings) > 0 {
			for _, w := range result.Warnings {
				fmt.Fprintf(os.Stderr, "WARNING: %s\n", w)
			}
		}

		// Find .arch directory
		ad, err := archdir.Find(".")
		if err != nil {
			return err
		}

		// Check if index already exists
		indexPath := ad.IndexPath()
		if !importForce {
			if _, err := os.Stat(indexPath); err == nil {
				return fmt.Errorf("index already exists at %s (use --force to overwrite)", indexPath)
			}
		}

		// Save
		if err := schema.SaveIndex(indexPath, &idx); err != nil {
			return err
		}

		// Print summary
		archetypeCount := 0
		for _, a := range idx.Archetypes {
			archetypeCount += len(a)
		}
		fmt.Fprintf(os.Stderr, "Imported to %s\n", indexPath)
		fmt.Fprintf(os.Stderr, "  Components:    %d\n", len(idx.Components))
		fmt.Fprintf(os.Stderr, "  Archetypes:    %d\n", archetypeCount)
		fmt.Fprintf(os.Stderr, "  Relationships: %d\n", len(idx.Relationships))
		fmt.Fprintf(os.Stderr, "  Flows:         %d\n", len(idx.Flows))

		return nil
	},
}

func init() {
	importCmd.Flags().BoolVar(&importForce, "force", false, "overwrite existing index.json")
	rootCmd.AddCommand(importCmd)
}
