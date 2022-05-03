package main

import (
	"reflect"
	"testing"
)

func TestParseR(t *testing.T) {
	expected := &InstructionR{
		Opcode: 0b1100100,
		Rd:     0b01101,
		Funct3: 0b010,
		Rs1:    0b01101,
		Rs2:    0b00101,
		Funct7: 0b0011011,
	}
	got := ParseR(0b0011011_00101_01101_010_01101_1100100)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}

func TestParseI(t *testing.T) {
	expected := &InstructionI{
		Opcode: 0b1100100,
		Rd:     0b01101,
		Funct3: 0b010,
		Rs1:    0b01101,
		Imm:    0b001101100101,
	}
	got := ParseI(0b001101100101_01101_010_01101_1100100)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected: %v, got: %v", expected, got)
	}

	// sign-extend is required
	expected = &InstructionI{
		Opcode: 0b1100100,
		Rd:     0b01101,
		Funct3: 0b010,
		Rs1:    0b01101,
		Imm:    0b11111111_11111111_11111111_11111111_11111111_11111111_1111_101101100101,
	}
	got = ParseI(0b101101100101_01101_010_01101_1100100)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}

func TestParseS(t *testing.T) {
	expected := &InstructionS{
		Opcode: 0b1100100,
		Funct3: 0b010,
		Rs1:    0b01101,
		Rs2:    0b00101,
		Imm:    0b0011011_01101,
	}
	got := ParseS(0b0011011_00101_01101_010_01101_1100100)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected: %v, got: %v", expected, got)
	}

	// sign-extend is required
	expected = &InstructionS{
		Opcode: 0b1100100,
		Funct3: 0b010,
		Rs1:    0b01101,
		Rs2:    0b00101,
		Imm:    0b11111111_11111111_11111111_11111111_11111111_11111111_1111_1011011_01101,
	}
	got = ParseS(0b1011011_00101_01101_010_01101_1100100)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}

func TestParseB(t *testing.T) {
	expected := &InstructionB{
		Opcode: 0b1100100,
		Funct3: 0b010,
		Rs1:    0b01101,
		Rs2:    0b00101,
		Imm:    0b0_1_011011_0110_0,
	}
	got := ParseB(0b0011011_00101_01101_010_01101_1100100)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected: %v, got: %v", expected, got)
	}

	// sign-extend is required
	expected = &InstructionB{
		Opcode: 0b1100100,
		Funct3: 0b010,
		Rs1:    0b01101,
		Rs2:    0b00101,
		Imm:    0b11111111_11111111_11111111_11111111_11111111_11111111_111_1_1_011011_0110_0,
	}
	got = ParseB(0b1011011_00101_01101_010_01101_1100100)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}

func TestParseJ(t *testing.T) {
	expected := &InstructionJ{
		Opcode: 0b1100100,
		Rd:     0b01101,
		Imm:    0b0_01101010_1_0110110010_0,
	}
	got := ParseJ(0b00110110010101101010_01101_1100100)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected: %v, got: %v", expected, got)
	}

	// sign-extend is required
	expected = &InstructionJ{
		Opcode: 0b1100100,
		Rd:     0b01101,
		Imm:    0b11111111_11111111_11111111_11111111_11111111_111_1_01101010_1_0110110010_0,
	}
	got = ParseJ(0b10110110010101101010_01101_1100100)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}
