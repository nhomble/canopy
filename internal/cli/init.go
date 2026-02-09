package cli

import (
	"fmt"

	"github.com/nhomble/canopy/internal/canopydir"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize .canopy/ directory in the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		ad, err := canopydir.Init(dir)
		if err != nil {
			return err
		}

		fmt.Printf("Initialized %s\n", ad.Root)
		fmt.Println("\nNext steps:")
		fmt.Println("  canopy prepare-analysis . | <your-llm-cli> | canopy import --force")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
