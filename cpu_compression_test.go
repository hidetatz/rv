package main

import "testing"

func TestCPU_Decompress(t *testing.T) {
	tests := map[string]struct {
		compressed   uint64
		expected     uint64
		expectedExcp Exception
	}{
		"c.sub": {
			compressed:   0b100011_101_00_111_01,
			expected:     0b0100000_01111_01101_000_01101_0110011,
			expectedExcp: ExcpNone,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cpu := &CPU{} // Decompression does not require any CPu state
			got, gotExcp := cpu.Decompress(tc.compressed)
			if gotExcp != tc.expectedExcp {
				t.Errorf("got %v, want %v", gotExcp, tc.expectedExcp)
			}

			if got != tc.expected {
				t.Errorf("\ngot:  %032b\nwant: %032b", got, tc.expected)
			}
		})
	}
}
