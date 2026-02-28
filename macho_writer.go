package main

import (
	"encoding/binary"
	"os"
)

const (
	MH_MAGIC_64 = 0xfeedfacf
	MH_EXECUTE  = 0x2
	
	CPU_TYPE_ARM64 = 0x0100000c
	CPU_SUBTYPE_ARM64_ALL = 0

	LC_SEGMENT_64 = 0x19
	LC_UNIXTHREAD = 0x5
)

func WriteMacho(filename string, code []uint32) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Helpers
	write32 := func(val uint32) {
		binary.Write(f, binary.LittleEndian, val)
	}
	write64 := func(val uint64) {
		binary.Write(f, binary.LittleEndian, val)
	}
	writeString16 := func(s string) {
		var b [16]byte
		copy(b[:], s)
		f.Write(b[:])
	}

	// Calculate Sizes
	// Header: 32 bytes
	// LC_SEGMENT_64 (72) + Section_64 (80) = 152 bytes
	// LC_UNIXTHREAD (16 + 272) = 288 bytes
	// Total Cmds = 440 bytes
	// Total Header Size = 32 + 440 = 472 bytes
	codeOffset := uint64(472)
	codeSize := uint64(len(code) * 4)
	
	totalFileSize := codeOffset + codeSize
	
	// VM Size must be at least 16KB (page size on M1)
	vmSize := uint64(16384)
    if totalFileSize > vmSize {
        vmSize = (totalFileSize + 16383) / 16384 * 16384
    }

	// 1. Mach-O Header
	write32(MH_MAGIC_64)
	write32(CPU_TYPE_ARM64)
	write32(CPU_SUBTYPE_ARM64_ALL)
	write32(MH_EXECUTE)
	write32(2) // ncmds
	write32(440) // sizeofcmds
	write32(0) // flags
	write32(0) // reserved

	// 2. LC_SEGMENT_64 (__TEXT)
	write32(LC_SEGMENT_64)
	write32(152) // cmdsize (72+80)
	writeString16("__TEXT")
	write64(0) // vmaddr
	write64(vmSize) // vmsize (Page Aligned)
	write64(0) // fileoff
	write64(totalFileSize) // filesize (Actual bytes on disk)
	write32(7) // maxprot (RWX)
	write32(7) // initprot (RWX)
	write32(1) // nsects
	write32(0) // flags

	// Section 64 (__text)
	writeString16("__text")
	writeString16("__TEXT")
	write64(codeOffset) // addr
	write64(codeSize) // size
	write32(uint32(codeOffset)) // offset
	write32(2) // align
	write32(0) // reloff
	write32(0) // nreloc
	write32(0x80000400) // flags
	write32(0) // reserved1
	write32(0) // reserved2
	write32(0) // reserved3

	// 3. LC_UNIXTHREAD
	write32(LC_UNIXTHREAD)
	write32(288) // cmdsize
	write32(6) // ARM_THREAD_STATE64
	write32(68) // count (34 uint64s * 2)

	// Registers
	for i := 0; i < 33; i++ {
		if i == 32 { // PC
			write64(codeOffset)
		} else {
			write64(0)
		}
	}
	write32(0) // cpsr
	write32(0) // pad

	// 4. Code
	current, _ := f.Seek(0, 1)
	padHeader := int64(codeOffset) - current
	if padHeader > 0 {
		f.Write(make([]byte, padHeader))
	}

	for _, inst := range code {
		write32(inst)
	}
	
	// No trailing padding needed on disk if filesize matches.
	
	f.Chmod(0755)
	return nil
}