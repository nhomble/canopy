package cli

import (
	"github.com/nicolas/arch-index/internal/archdir"
	"github.com/nicolas/arch-index/internal/server"
	"github.com/spf13/cobra"
)

var (
	servePort int
	serveHost string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the local HTTP server for architectural queries",
	RunE: func(cmd *cobra.Command, args []string) error {
		ad, err := archdir.Find(".")
		if err != nil {
			return err
		}

		return server.Run(ad.IndexPath(), serveHost, servePort)
	},
}

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 3451, "port to listen on")
	serveCmd.Flags().StringVar(&serveHost, "host", "127.0.0.1", "host to bind to")
	rootCmd.AddCommand(serveCmd)
}
