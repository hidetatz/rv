package main

import (
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
		"rv64ui-p-and",
		"rv64ui-p-andi",
		"rv64ui-p-auipc",
		"rv64ui-p-beq",
		"rv64ui-p-bge",
		"rv64ui-p-bgeu",
		"rv64ui-p-blt",
		"rv64ui-p-bltu",
		"rv64ui-p-bne",
		"rv64ui-p-fence_i",
		"rv64ui-p-jal",
		"rv64ui-p-jalr",
		"rv64ui-p-lb",
		"rv64ui-p-lbu",
		"rv64ui-p-ld",
		"rv64ui-p-lh",
		"rv64ui-p-lhu",
		"rv64ui-p-lui",
		"rv64ui-p-lwu",
		"rv64ui-p-or",
		"rv64ui-p-ori",
		"rv64ui-p-sb",
		"rv64ui-p-sd",
		"rv64ui-p-sh",
		"rv64ui-p-simple",
		"rv64ui-p-sll",
		"rv64ui-p-slli",
		"rv64ui-p-slt",
		"rv64ui-p-slti",
		"rv64ui-p-sltiu",
		"rv64ui-p-sltu",
		"rv64ui-p-sra",
		"rv64ui-p-srai",
		"rv64ui-p-srl",
		"rv64ui-p-srli",
		"rv64ui-p-sub",
		"rv64ui-p-xor",
		"rv64ui-p-xori",

		"rv64ui-p-addiw",
		"rv64ui-p-addw",
		"rv64ui-p-lw",
		"rv64ui-p-slliw",
		"rv64ui-p-sllw",
		"rv64ui-p-sraiw",
		"rv64ui-p-sraw",
		"rv64ui-p-srliw",
		"rv64ui-p-srlw",
		"rv64ui-p-subw",
		"rv64ui-p-sw",

		"rv64ua-p-lrsc",
		"rv64ua-p-amoadd_w",
		"rv64ua-p-amoand_w",
		"rv64ua-p-amomaxu_w",
		"rv64ua-p-amomax_w",
		"rv64ua-p-amominu_w",
		"rv64ua-p-amomin_w",
		"rv64ua-p-amoor_w",
		"rv64ua-p-amoswap_w",
		"rv64ua-p-amoxor_w",

		"rv64ua-p-amoadd_d",
		"rv64ua-p-amoand_d",
		"rv64ua-p-amomax_d",
		"rv64ua-p-amomaxu_d",
		"rv64ua-p-amomin_d",
		"rv64ua-p-amominu_d",
		"rv64ua-p-amoor_d",
		"rv64ua-p-amoswap_d",
		"rv64ua-p-amoxor_d",

		"rv64um-p-mul",
		"rv64um-p-mulh",
		"rv64um-p-mulhsu",
		"rv64um-p-mulhu",
		"rv64um-p-div",
		"rv64um-p-divu",
		"rv64um-p-rem",
		"rv64um-p-remu",

		"rv64um-p-mulw",
		"rv64um-p-divuw",
		"rv64um-p-divw",
		"rv64um-p-remuw",
		"rv64um-p-remw",

		"rv64ui-v-add",
		"rv64ui-v-addi",
		"rv64ui-v-addiw",
		"rv64ui-v-addw",
		"rv64ui-v-and",
		"rv64ui-v-andi",
		"rv64ui-v-auipc",
		"rv64ui-v-beq",
		"rv64ui-v-bge",
		"rv64ui-v-bgeu",
		"rv64ui-v-blt",
		"rv64ui-v-bltu",
		"rv64ui-v-bne",
		"rv64ui-v-fence_i",
		"rv64ui-v-jal",
		"rv64ui-v-jalr",
		"rv64ui-v-lb",
		"rv64ui-v-lbu",
		"rv64ui-v-ld",
		"rv64ui-v-lh",
		"rv64ui-v-lhu",
		"rv64ui-v-lui",
		"rv64ui-v-lw",
		"rv64ui-v-lwu",
		"rv64ui-v-or",
		"rv64ui-v-ori",
		//"rv64ui-v-sb",
		//"rv64ui-v-sd",
		//"rv64ui-v-sh",
		"rv64ui-v-simple",
		"rv64ui-v-sll",
		"rv64ui-v-slli",
		"rv64ui-v-slliw",
		"rv64ui-v-sllw",
		"rv64ui-v-slt",
		"rv64ui-v-slti",
		"rv64ui-v-sltiu",
		"rv64ui-v-sltu",
		"rv64ui-v-sra",
		"rv64ui-v-srai",
		"rv64ui-v-sraiw",
		"rv64ui-v-sraw",
		"rv64ui-v-srl",
		"rv64ui-v-srli",
		"rv64ui-v-srliw",
		"rv64ui-v-srlw",
		"rv64ui-v-sub",
		"rv64ui-v-subw",
		//"rv64ui-v-sw",
		"rv64ui-v-xor",
		"rv64ui-v-xori",
	}

	for _, tc := range tests {
		t.Run(tc, func(t *testing.T) {
			cpu, err := initCPU(filepath.Join("./tests/", tc))
			if err != nil {
				t.Fatalf("initialize RV: %s, %s", tc, err)
			}

			if cpu.tohost == 0 {
				t.Fatalf("unexpected error: tohost is 0 but expected some address in the binary! %s", tc)
			}

			if err := cpu.Start(); err != nil {
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
