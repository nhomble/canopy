package cli

import (
	"github.com/spf13/cobra"
)

var archDir string

var rootCmd = &cobra.Command{
	Use:   "canopy",
	Short: "Generate and serve architectural context for codebases",
	Long: `canopy uses LLMs to analyze codebases and generate architectural
context that developers can explore interactively.

Workflow:
  1. canopy init                    Create .canopy/ directory
  2. canopy prepare-analysis .      Generate analysis prompt
  3. Feed prompt to your LLM, save result as JSON
  4. canopy import result.json      Validate and save analysis
  5. canopy serve                   Start local query server`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&archDir, "canopy-dir", ".canopy", "path to .canopy directory")
}

func Execute() error {
	return rootCmd.Execute()
}
