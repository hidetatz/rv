package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

// TestE2E runs riscv-tests (https://github.com/riscv-software-src/riscv-tests) and make sure
// every test suite passes.
// Before running this test, test binary must locate in "./tests/" directory.
func TestE2E(t *testing.T) {
	tests := []string{
		"rv64ui-p-add",
	}

	for _, tc := range tests {
		f, err := os.Open(filepath.Join("./tests/", tc))
		if err != nil {
			t.Fatalf("open file: %s, %s", tc, err)
		}
		defer f.Close()

		buff, err := io.ReadAll(f)
		if err != nil {
			t.Fatalf("read file: %s, %s", tc, err)
		}

		emulator, err := New(buff)
		if err != nil {
			t.Fatalf("initialize RV: %s, %s", tc, err)
		}

		if emulator.tohost == 0 {
			t.Fatalf("unexpected error: tohost is 0 but expected some address in the binary! %s", tc)
		}

		if err := emulator.Start(); err != nil {
			t.Errorf("fail to run: %s, %s", tc, err)
		}
	}
}
