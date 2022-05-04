package main

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestE2E runs riscv-tests (https://github.com/riscv-software-src/riscv-tests) and make sure
// every test suite passes.
// Before running this test, test binary must locate in "./tests/" directory.
func TestE2E(t *testing.T) {
	tests := []string{
		"rv64ui-p-add",
		"rv64ui-p-addi",
		"rv64ui-p-addiw",
		"rv64ui-p-addw",
		"rv64ui-p-and",
		"rv64ui-p-andi",
		"rv64ui-p-auipc",
		"rv64ui-p-beq",
		"rv64ui-p-bge",
		"rv64ui-p-bgeu",
		"rv64ui-p-blt",
		"rv64ui-p-bltu",
		"rv64ui-p-bne",
		//"rv64ui-p-fence_i",
		//"rv64ui-p-jal",
		//"rv64ui-p-jalr",
		"rv64ui-p-lb",
		"rv64ui-p-lbu",
		"rv64ui-p-ld",
		"rv64ui-p-lh",
		"rv64ui-p-lhu",
		"rv64ui-p-lui",
		"rv64ui-p-lw",
		"rv64ui-p-lwu",
		"rv64ui-p-or",
		"rv64ui-p-ori",
		"rv64ui-p-sb",
		"rv64ui-p-sd",
		"rv64ui-p-sh",
		"rv64ui-p-simple",
		//"rv64ui-p-sll",
		//"rv64ui-p-slli",
		"rv64ui-p-slliw",
		//"rv64ui-p-sllw",
		"rv64ui-p-slt",
		"rv64ui-p-slti",
		"rv64ui-p-sltiu",
		//"rv64ui-p-sltu",
		"rv64ui-p-sra",
		"rv64ui-p-srai",
		"rv64ui-p-sraiw",
		//"rv64ui-p-sraw",
		//"rv64ui-p-srl",
		//"rv64ui-p-srli",
		"rv64ui-p-srliw",
		//"rv64ui-p-srlw",
		"rv64ui-p-sub",
		"rv64ui-p-subw",
		"rv64ui-p-sw",
		"rv64ui-p-xor",
		"rv64ui-p-xori",
	}

	for _, tc := range tests {
		t.Run(tc, func(t *testing.T) {
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

		})

		// Each test runs rv emulator which internally contains 4GiB space for DRAM emulation.
		// Because GitHub Actions single hosted runner only has 7GB memory,
		// sometimes the test is killed by OOM without this manual GC.
		// TODO: optimize somehow
		runtime.GC()
	}
}
