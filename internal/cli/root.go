package cli

import (
	"github.com/spf13/cobra"
)

var archDir string

var rootCmd = &cobra.Command{
	Use:   "arch-index",
	Short: "Generate and serve architectural context for codebases",
	Long: `arch-index uses LLMs to analyze codebases and generate architectural
context that developers can explore interactively.

Workflow:
  1. arch-index init                    Create .arch/ directory
  2. arch-index prepare-analysis .      Generate analysis prompt
  3. Feed prompt to your LLM, save result as JSON
  4. arch-index import result.json      Validate and save analysis
  5. arch-index serve                   Start local query server`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&archDir, "arch-dir", ".arch", "path to .arch directory")
}

func Execute() error {
	return rootCmd.Execute()
}
