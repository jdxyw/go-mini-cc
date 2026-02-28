package main

import (
	"fmt"
)

// Assembler for ARM64
type Assembler struct {
	code []uint32
}

func NewAssembler() *Assembler {
	return &Assembler{
		code: []uint32{},
	}
}

func (a *Assembler) Emit(inst uint32) {
	a.code = append(a.code, inst)
}

// MOVZ xd, #imm16
func (a *Assembler) MovImm(reg int, val int) {
	if val < 0 || val > 65535 {
		// For simplicity, panic. Real assembler handles MOVK etc.
		panic(fmt.Sprintf("Immediate %d out of range for MOVZ", val))
	}
	// 0xD2800000 | (imm << 5) | reg
	inst := uint32(0xD2800000) | uint32(val)<<5 | uint32(reg)
	a.Emit(inst)
}

// ADD xd, xn, xm
func (a *Assembler) Add(dst, src1, src2 int) {
	// 0x8B000000 | rm << 16 | rn << 5 | rd
	inst := uint32(0x8B000000) | uint32(src2)<<16 | uint32(src1)<<5 | uint32(dst)
	a.Emit(inst)
}

// SUB xd, xn, xm
func (a *Assembler) Sub(dst, src1, src2 int) {
	// 0xCB000000 | rm << 16 | rn << 5 | rd
	inst := uint32(0xCB000000) | uint32(src2)<<16 | uint32(src1)<<5 | uint32(dst)
	a.Emit(inst)
}

// MUL xd, xn, xm (MADD xd, xn, xm, xzr)
// 1001 1011 000 rm 000000 rn rd
// 0x9B000000 | rm<<16 | rn<<5 | rd
func (a *Assembler) Mul(dst, src1, src2 int) {
	inst := uint32(0x9B000000) | uint32(src2)<<16 | uint32(src1)<<5 | uint32(dst)
	a.Emit(inst)
}

// SDIV xd, xn, xm
// 1001 1010 110 rm 000010 rn rd
// 0x9AC00000 | rm<<16 | rn<<5 | rd
func (a *Assembler) Sdiv(dst, src1, src2 int) {
	inst := uint32(0x9AC00000) | uint32(src2)<<16 | uint32(src1)<<5 | uint32(dst)
	a.Emit(inst)
}

// STR xd, [sp, #-16]! (Pre-indexed)
func (a *Assembler) Push(reg int) {
	// 0xF81F0FE0 | rt
	inst := uint32(0xF81F0FE0) | uint32(reg)
	a.Emit(inst)
}

// LDR xd, [sp], #16 (Post-indexed)
func (a *Assembler) Pop(reg int) {
	// 0xF84107E0 | rt
	inst := uint32(0xF84107E0) | uint32(reg)
	a.Emit(inst)
}

// LDUR rt, [rn, #simm9] (Load Unscaled)
// 1011 1000 010 imm9 00 rn rt
// 0xB8400000
func (a *Assembler) Ldur(dst, base, offset int) {
	// Offset is signed 9-bit (-256 to 255)
	if offset < -256 || offset > 255 {
		panic("Offset out of range for LDUR")
	}
	imm9 := uint32(offset) & 0x1FF
	inst := uint32(0xB8400000) | (imm9 << 12) | uint32(base)<<5 | uint32(dst)
	a.Emit(inst)
}

// RET
func (a *Assembler) Ret() {
	a.Emit(0xD65F03C0)
}

// SVC #imm
func (a *Assembler) Svc(imm int) {
	a.Emit(0xD4000001) // Fixed #0 for now
}