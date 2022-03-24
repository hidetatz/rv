package main

import "testing"

func TestCPU_Decompress(t *testing.T) {
	tests := map[string]struct {
		compressed   uint64
		expected     uint64
		expectedExcp Exception
	}{
		// CR
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
		"c.add": {
			compressed: 0b1001_01101_01111_10,
			expected:   0b0000000_01111_01101_000_01101_0110011,
		},
		// CI
		"c.nop": {
			compressed: 0b000_1_00000_01111_01,
			expected:   0b000000000000_00000_000_00000_0010011,
		},
		"c.addi (positive imm)": {
			compressed: 0b000_0_00101_01101_01,
			expected:   0b000000001101_00101_000_00101_0010011,
		},
		"c.addi (negative imm)": {
			compressed: 0b000_1_00101_01101_01,
			// internally this is treated as uint64
			expected: 0b11111111111111111111111111111111_111111101101_00101_000_00101_0010011,
		},
		"c.li (positive imm)": {
			compressed: 0b010_0_00101_01101_01,
			expected:   0b000000001101_00000_000_00101_0010011,
		},
		"c.li (negative imm)": {
			compressed: 0b010_1_00101_01101_01,
			// internally this is treated as uint64
			expected: 0b11111111111111111111111111111111_111111101101_00000_000_00101_0010011,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			cpu := &CPU{} // Decompression does not require any CPU state
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
