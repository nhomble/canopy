package cli

import (
	"fmt"

	"github.com/nicolas/arch-index/internal/archdir"
	"github.com/nicolas/arch-index/internal/schema"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the existing .arch/index.json",
	RunE: func(cmd *cobra.Command, args []string) error {
		ad, err := archdir.Find(".")
		if err != nil {
			return err
		}

		idx, err := schema.LoadIndex(ad.IndexPath())
		if err != nil {
			return err
		}

		result := schema.ValidateIndex(idx)
		fmt.Print(result.FormatResult())

		if !result.Valid {
			return fmt.Errorf("validation failed")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
