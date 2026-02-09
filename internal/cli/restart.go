package cli

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/nhomble/arch-index/internal/archdir"
	"github.com/nhomble/arch-index/internal/server"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Stop the running server and start a fresh one",
	RunE: func(cmd *cobra.Command, args []string) error {
		ad, err := archdir.Find(".")
		if err != nil {
			return err
		}

		port := servePort
		if port == 0 {
			repoRoot := filepath.Dir(ad.Root)
			port = server.DeterministicPort(repoRoot)
		}

		addr := fmt.Sprintf("%s:%d", serveHost, port)

		// Try to stop the existing server.
		if pid, err := findListenerPID(port); err == nil && pid > 0 {
			log.Printf("Stopping server (pid %d) on %s", pid, addr)
			proc, err := os.FindProcess(pid)
			if err == nil {
				proc.Signal(syscall.SIGTERM)
				// Wait briefly for it to exit.
				for i := 0; i < 20; i++ {
					time.Sleep(50 * time.Millisecond)
					if !isPortOpen(addr) {
						break
					}
				}
			}
		} else if isPortOpen(addr) {
			return fmt.Errorf("port %d is in use but could not find owner process", port)
		} else {
			log.Printf("No server running on %s", addr)
		}

		log.Printf("Starting server on %s", addr)
		return server.Run(ad.IndexPath(), serveHost, port)
	},
}

// findListenerPID uses lsof to find the PID listening on a port.
func findListenerPID(port int) (int, error) {
	// Try connecting first — if nothing is listening, skip the lookup.
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	if !isPortOpen(addr) {
		return 0, fmt.Errorf("nothing listening on port %d", port)
	}

	out, err := exec.Command("lsof", "-ti", fmt.Sprintf("tcp:%d", port)).Output()
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, fmt.Errorf("parsing pid: %w", err)
	}
	return pid, nil
}

func isPortOpen(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
