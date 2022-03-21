package main

import "testing"

func TestCPU_Decompress(t *testing.T) {
	tests := map[string]struct {
		compressed   uint64
		expected     uint64
		expectedExcp Exception
	}{
		"c.sub": {
			compressed: 0b1000_11101_00111_01,
			expected:   0b0100000_01111_01101_000_01101_0110011,
		},
		"c.xor": {
			compressed: 0b1000_11101_01111_01,
			expected:   0b0000000_01111_01101_100_01101_0110011,
		},
		"c.or": {
			compressed: 0b1000_11101_10111_01,
			expected:   0b0000000_01111_01101_110_01101_0110011,
		},
		"c.and": {
			compressed: 0b1000_11101_11111_01,
			expected:   0b0000000_01111_01101_111_01101_0110011,
		},
		"c.mv": {
			compressed: 0b1000_01101_01111_10,
			expected:   0b0000000_01111_00000_000_01101_0110011,
		},
		//"c.add": {},
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
