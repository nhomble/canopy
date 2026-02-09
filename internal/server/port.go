package server

import (
	"hash/fnv"
)

const (
	portBase  = 10000
	portRange = 50000
)

// DeterministicPort computes a stable port number from a repo root path.
// The same absolute path always produces the same port in [10000, 59999].
// Uses FNV-1a 32-bit, matching the Lua implementation in the Neovim plugin.
func DeterministicPort(repoRoot string) int {
	h := fnv.New32a()
	h.Write([]byte(repoRoot))
	return portBase + int(h.Sum32()%uint32(portRange))
}
