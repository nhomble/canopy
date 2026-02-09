package cli

import (
	"log"
	"path/filepath"

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

		port := servePort
		if port == 0 {
			repoRoot := filepath.Dir(ad.Root)
			port = server.DeterministicPort(repoRoot)
			log.Printf("Auto-assigned port %d for %s", port, repoRoot)
		}

		return server.Run(ad.IndexPath(), serveHost, port)
	},
}

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 0,
		"port to listen on (0 = auto-assign from repo path)")
	serveCmd.Flags().StringVar(&serveHost, "host", "127.0.0.1", "host to bind to")
	rootCmd.AddCommand(serveCmd)
}
