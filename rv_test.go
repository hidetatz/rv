package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

// This test is not working for now.
func TestInstructions(t *testing.T) {
	type testcase struct {
		name string
		asm  []string
		// key: xreg index, value: value inside the register
		expectedXRegs map[uint8]uint64
	}

	testcases := []testcase{
		{
			name: "addi",
			asm: []string{
				"addi x16,x0,3",
			},
			expectedXRegs: map[uint8]uint64{
				16: 3,
			},
		},
	}

	for i, tt := range testcases {
		i := i
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			as := `
	.text
	.align 2
	.globl main
main:
`
			for _, s := range tt.asm {
				as += "	" + s + "\n"
			}

			dir := fmt.Sprintf("./test/instructions/%d", i)
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				t.Fatalf("make a directory: %v", err)
			}

			cmd := exec.Command(
				"riscv64-unknown-linux-gnu-gcc",
				"-O0",
				"-o",
				fmt.Sprintf("%s/a.out", dir),
				"-xassembler",
				"-Wl,-Ttext=0x80000000",
				"-",
			)
			stdin, err := cmd.StdinPipe()
			if err != nil {
				t.Fatalf("obtain stdin pipe: %v", err)
			}

			io.WriteString(stdin, as)
			stdin.Close()

			_, err = cmd.Output()
			if err != nil {
				t.Fatalf("obtain gcc output: %v", err)
			}

			prog, err := os.ReadFile(fmt.Sprintf("%s/a.out", dir))
			if err != nil {
				t.Fatalf("read test binary: %v", err)
			}

			rv, err := New(prog)
			if err != nil {
				t.Fatalf("initialize rv emulator: %v", err)
			}
			rv.StartForTest(len(as))
			if !reflect.DeepEqual(rv.cpu.XRegs, tt.expectedXRegs) {
				t.Fatalf("fail")
			}
		})
	}
}
