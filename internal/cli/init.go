package cli

import (
	"fmt"

	"github.com/nhomble/arch-index/internal/archdir"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize .arch/ directory in the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		ad, err := archdir.Init(dir)
		if err != nil {
			return err
		}

		fmt.Printf("Initialized %s\n", ad.Root)
		fmt.Println("\nNext steps:")
		fmt.Println("  arch-index prepare-analysis . | <your-llm-cli> | arch-index import --force")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
