package cli

import (
	"fmt"

	"github.com/nhomble/canopy/internal/canopydir"
	"github.com/nhomble/canopy/internal/schema"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the existing .canopy/index.json",
	RunE: func(cmd *cobra.Command, args []string) error {
		ad, err := canopydir.Find(".")
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
