package rv

import (
	"reflect"
	"testing"
)

func TestDecode32(t *testing.T) {
	// sra
	cpu := NewCPU()
	got := cpu.Decode32(0b0100000_01010_10100_101_00001_0110011)
	expected := &Instruction{
		Raw:    0b0100000_01010_10100_101_00001_0110011,
		Op:     InsSRA,
		Opcode: 0b0110011,
		Rd:     0b00001,
		Funct3: 0b101,
		Rs1:    0b10100,
		Rs2:    0b01010,
		Funct7: 0b0100000,
	}
	if !reflect.DeepEqual(expected, got) {
		t.Fatalf("expected: %v, got: %v", expected, got)
	}
	//	tests := struct {
	//		ins    uint32
	//		Op     uint32
	//		Format uint8
	//	}{
	//		{
	//			//
	//			ins: 0b0000_0000_0000_0000_0000_0000_0000_0000,
	//			out: nil,
	//		},
	//	}
}
