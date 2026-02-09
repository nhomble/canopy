package server

import "testing"

func TestDeterministicPort(t *testing.T) {
	// These vectors MUST match the Lua implementation in port.lua.
	tests := []struct {
		path string
		want int
	}{
		{"/home/user/my-project", 50556},
		{"/Users/nicolas/dev/codebase-viz", 35385},
		{"/tmp/hexagonal-ddd", 25612},
		{"/tmp/test", 17752},
	}
	for _, tt := range tests {
		got := DeterministicPort(tt.path)
		if got != tt.want {
			t.Errorf("DeterministicPort(%q) = %d, want %d", tt.path, got, tt.want)
		}
		if got < 10000 || got > 59999 {
			t.Errorf("DeterministicPort(%q) = %d, out of range [10000, 59999]", tt.path, got)
		}
	}
}

func TestDeterministicPortStability(t *testing.T) {
	p1 := DeterministicPort("/some/path")
	p2 := DeterministicPort("/some/path")
	if p1 != p2 {
		t.Errorf("not stable: %d != %d", p1, p2)
	}
}
