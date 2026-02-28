package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	outputFlag := flag.String("o", "a.out", "Output file name")
	assemblyOnly := flag.Bool("S", false, "Generate assembly text only")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: mini-gcc [options] <file.c>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	inputFile := args[0]
	code, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		os.Exit(1)
	}

	l := NewLexer(string(code))
	p := NewParser(l)
	prog := p.ParseProgram()
	
	if prog == nil {
		fmt.Println("Parser returned nil.")
		os.Exit(1)
	}

	if len(p.errors) > 0 {
		fmt.Println("Parser errors:")
		for _, msg := range p.errors {
			fmt.Printf("\t%s\n", msg)
		}
		os.Exit(1)
	}

	// 1. Text Assembly Generation (-S)
	if *assemblyOnly {
		cg := NewCodeGenerator()
		asm := cg.Generate(prog)
		
		asmFile := *outputFlag
		if asmFile == "a.out" {
			asmFile = strings.TrimSuffix(inputFile, filepath.Ext(inputFile)) + ".s"
		}
		
		err = os.WriteFile(asmFile, []byte(asm), 0644)
		if err != nil {
			fmt.Printf("Error writing assembly: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Assembly generated: %s\n", asmFile)
		return
	}

	// 2. Native Compilation (Machine Code + Mach-O)
	bcg := NewBinaryCodeGenerator()
	binaryCode := bcg.Generate(prog)
	
	outputFile := *outputFlag
	err = WriteMacho(outputFile, binaryCode)
	if err != nil {
		fmt.Printf("Error writing executable: %s\n", err)
		os.Exit(1)
	}

	// Ad-hoc code signing for ARM64 macOS
	cmd := exec.Command("codesign", "-s", "-", "--entitlements", "entitlements.plist", outputFile)
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Warning: Failed to sign binary: %s\nOutput: %s\n", err, out)
	}

	fmt.Printf("Compilation successful! Output: %s\n", outputFile)
}