package main

import (
	"fmt"
	"strings"
	"sort"
)

type Symbol struct {
	Offset   int
	IsArray  bool
	IsGlobal bool
}

type CodeGenerator struct {
	output           strings.Builder
	symbolTable      map[string]Symbol
	globalSymbolTable map[string]bool 
	stringTable      map[string]string 
	stackOffset      int
	labelCount       int
}

func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		symbolTable:       make(map[string]Symbol),
		globalSymbolTable: make(map[string]bool),
		stringTable:       make(map[string]string),
		stackOffset:       0,
		labelCount:        0,
	}
}

func (cg *CodeGenerator) newLabel() string {
	cg.labelCount++
	return fmt.Sprintf("L%d", cg.labelCount)
}

func (cg *CodeGenerator) newStringLabel() string {
	return fmt.Sprintf("LC%d", len(cg.stringTable))
}

func (cg *CodeGenerator) Generate(prog *Program) string {
	cg.output.Reset()
	cg.globalSymbolTable = make(map[string]bool)

	// Globals
	if len(prog.Globals) > 0 {
		cg.emit(".data")
		cg.emit(".align 2")
		for _, g := range prog.Globals {
			cg.emit(fmt.Sprintf("_%s:", g.Name))
			if g.Size > 0 {
				cg.emit(fmt.Sprintf("    .zero %d", g.Size * 8))
				cg.symbolTable[g.Name] = Symbol{Offset: 0, IsArray: true, IsGlobal: true}
			} else {
				val := 0
				if lit, ok := g.Value.(*IntegerLiteral); ok {
					val = lit.Value
				}
				cg.emit(fmt.Sprintf("    .word %d", val))
				cg.symbolTable[g.Name] = Symbol{Offset: 0, IsArray: false, IsGlobal: true}
			}
		}
		cg.emit("")
	}
	
	// Functions code
	var funcOutput strings.Builder
	mainOutput := cg.output
	cg.output = funcOutput
	
cg.emit(".text")
cg.emit(".global _main")
cg.emit(".align 2")
	
	for _, fn := range prog.Functions {
		cg.generateFunction(fn)
	}
	
	funcCode := cg.output.String()
	cg.output = mainOutput 
	
	// Strings
	if len(cg.stringTable) > 0 {
		cg.emit(".section __TEXT,__cstring,cstring_literals")
		var labels []string
		labelToContent := make(map[string]string)
		for c, l := range cg.stringTable {
			labels = append(labels, l)
			labelToContent[l] = c
		}
		sort.Strings(labels)
		
		for _, l := range labels {
			cg.emit(fmt.Sprintf("%s:", l))
			cg.emit(fmt.Sprintf("    .asciz \"%s\"", labelToContent[l]))
		}
		cg.emit("")
	}
	
cg.output.WriteString(funcCode)
	
	return cg.output.String()
}

func (cg *CodeGenerator) generateFunction(fn *Function) {
	cg.symbolTable = make(map[string]Symbol)
	cg.stackOffset = 0
	
cg.emit("")
cg.emit(fmt.Sprintf("_%s:", fn.Name))
	
cg.emit("    stp x29, x30, [sp, #-16]!")
cg.emit("    mov x29, sp")

	for i, param := range fn.Parameters {
		if i >= 8 {
			panic("Only support up to 8 parameters")
		}
		cg.emit(fmt.Sprintf("    str x%d, [sp, #-16]!", i))
		cg.stackOffset -= 16
		cg.symbolTable[param.Name] = Symbol{Offset: cg.stackOffset, IsArray: false}
	}

	if fn.Body != nil {
		cg.generateBlock(fn.Body)
	}

	cg.emit("    mov x0, #0") 
	cg.emit("    mov sp, x29") 
	cg.emit("    ldp x29, x30, [sp], #16")
cg.emit("    ret")
}

func (cg *CodeGenerator) generateBlock(block *Block) {
	for _, stmt := range block.Statements {
		cg.generateStatement(stmt)
	}
}

func (cg *CodeGenerator) generateStatement(stmt Statement) {
	switch s := stmt.(type) {
	case *ReturnStatement:
		cg.generateExpression(s.Value)
		cg.emit("    mov sp, x29")
		cg.emit("    ldp x29, x30, [sp], #16")
		cg.emit("    ret")
	
	case *VarStatement:
		if s.Size > 0 {
			size := s.Size * 8
			if size % 16 != 0 {
				size += 8
			}
			cg.stackOffset -= size
			cg.emit(fmt.Sprintf("    sub sp, sp, #%d", size))
			cg.symbolTable[s.Name] = Symbol{Offset: cg.stackOffset, IsArray: true}
		} else {
			cg.generateExpression(s.Value)
			cg.emit("    str x0, [sp, #-16]!")
			cg.stackOffset -= 16
			cg.symbolTable[s.Name] = Symbol{Offset: cg.stackOffset, IsArray: false}
		}

	case *ExpressionStatement:
		cg.generateExpression(s.Expression)

	case *IfStatement:
		elseLabel := cg.newLabel()
		endLabel := cg.newLabel()

		cg.generateExpression(s.Condition)
		cg.emit("    cmp x0, #0")
		cg.emit(fmt.Sprintf("    b.eq %s", elseLabel))
		
		cg.generateBlock(s.Consequence)
		cg.emit(fmt.Sprintf("    b %s", endLabel))
		
cg.emit(fmt.Sprintf("%s:", elseLabel))
		if s.Alternative != nil {
			cg.generateBlock(s.Alternative)
		}
		cg.emit(fmt.Sprintf("%s:", endLabel))

	case *WhileStatement:
		startLabel := cg.newLabel()
		endLabel := cg.newLabel()

		cg.emit(fmt.Sprintf("%s:", startLabel))
		cg.generateExpression(s.Condition)
		cg.emit("    cmp x0, #0")
		cg.emit(fmt.Sprintf("    b.eq %s", endLabel))

		cg.generateBlock(s.Body)
		cg.emit(fmt.Sprintf("    b %s", startLabel))
		
cg.emit(fmt.Sprintf("%s:", endLabel))

	case *ForStatement:
		startLabel := cg.newLabel()
		endLabel := cg.newLabel()
		
		if s.Init != nil {
			cg.generateStatement(s.Init)
		}
		
cg.emit(fmt.Sprintf("%s:", startLabel))
		
		if s.Condition != nil {
			cg.generateExpression(s.Condition)
			cg.emit("    cmp x0, #0")
			cg.emit(fmt.Sprintf("    b.eq %s", endLabel))
		}
		
cg.generateBlock(s.Body)
		
		if s.Post != nil {
			cg.generateStatement(s.Post) 
		}
		
cg.emit(fmt.Sprintf("    b %s", startLabel))
		cg.emit(fmt.Sprintf("%s:", endLabel))
	}
}

func (cg *CodeGenerator) generateExpression(expr Expression) {
	switch e := expr.(type) {
	case *IntegerLiteral:
		cg.emit(fmt.Sprintf("    mov x0, #%d", e.Value))
	
	case *StringLiteral:
		label, ok := cg.stringTable[e.Value]
		if !ok {
			label = cg.newStringLabel()
			cg.stringTable[e.Value] = label
		}
		cg.emit(fmt.Sprintf("    adrp x0, %s@PAGE", label))
		cg.emit(fmt.Sprintf("    add x0, x0, %s@PAGEOFF", label))

	case *Identifier:
		offset, ok := cg.symbolTable[e.Name]
		if ok {
			if offset.IsArray {
				cg.emit(fmt.Sprintf("    add x0, x29, #%d", offset.Offset))
			} else {
				cg.emit(fmt.Sprintf("    ldr x0, [x29, #%d]", offset.Offset))
			}
		} else {
			if cg.globalSymbolTable[e.Name] {
				cg.emit(fmt.Sprintf("    adrp x0, _%s@PAGE", e.Name))
				cg.emit(fmt.Sprintf("    add x0, x0, _%s@PAGEOFF", e.Name))
				// Hack: Assuming scalar global
				cg.emit("    ldrsw x0, [x0]")
			} else {
				panic(fmt.Sprintf("Undefined variable: %s", e.Name))
			}
		}
	
	case *AssignExpression:
		if ident, ok := e.Left.(*Identifier); ok {
			cg.generateExpression(e.Value)
			sym, ok := cg.symbolTable[ident.Name]
			if ok {
				cg.emit(fmt.Sprintf("    str x0, [x29, #%d]", sym.Offset))
			} else {
				if cg.globalSymbolTable[ident.Name] {
					cg.emit("    str x0, [sp, #-16]!") 
					cg.emit(fmt.Sprintf("    adrp x1, _%s@PAGE", ident.Name))
					cg.emit(fmt.Sprintf("    add x1, x1, _%s@PAGEOFF", ident.Name))
					cg.emit("    ldr x0, [sp], #16") 
					cg.emit("    str w0, [x1]") 
				} else {
					panic(fmt.Sprintf("Undefined variable in assignment: %s", ident.Name))
				}
			}
		} else if prefix, ok := e.Left.(*PrefixExpression); ok && prefix.Operator == "*" {
			cg.generateExpression(e.Value)
			cg.emit("    str x0, [sp, #-16]!") 
			cg.generateExpression(prefix.Right)
			cg.emit("    mov x1, x0")
			cg.emit("    ldr x0, [sp], #16")
			cg.emit("    str x0, [x1]")
		} else if idx, ok := e.Left.(*IndexExpression); ok {
			cg.generateExpression(e.Value)
			cg.emit("    str x0, [sp, #-16]!") 
			cg.generateExpression(idx.Left)
			cg.emit("    str x0, [sp, #-16]!") 
			cg.generateExpression(idx.Index)
			cg.emit("    lsl x0, x0, #3") 
			cg.emit("    ldr x1, [sp], #16") 
			cg.emit("    add x1, x1, x0") 
			cg.emit("    ldr x0, [sp], #16") 
			cg.emit("    str x0, [x1]")
		} else {
			panic("Invalid assignment target")
		}

	case *IndexExpression:
		cg.generateExpression(e.Left)
		cg.emit("    str x0, [sp, #-16]!") 
		cg.generateExpression(e.Index)
		cg.emit("    lsl x0, x0, #3") 
		cg.emit("    ldr x1, [sp], #16") 
		cg.emit("    add x0, x1, x0") 
		cg.emit("    ldr x0, [x0]")

	case *PrefixExpression:
		if e.Operator == "-" {
			cg.generateExpression(e.Right)
			cg.emit("    neg x0, x0")
		} else if e.Operator == "&" {
			ident, ok := e.Right.(*Identifier)
			if !ok {
				panic("Can only take address of identifier")
			}
			sym, ok := cg.symbolTable[ident.Name]
			if ok {
				cg.emit(fmt.Sprintf("    add x0, x29, #%d", sym.Offset))
			} else {
				if cg.globalSymbolTable[ident.Name] {
					cg.emit(fmt.Sprintf("    adrp x0, _%s@PAGE", ident.Name))
					cg.emit(fmt.Sprintf("    add x0, x0, _%s@PAGEOFF", ident.Name))
				} else {
					panic("Undefined variable for &")
				}
			}
		} else if e.Operator == "*" {
			cg.generateExpression(e.Right)
			cg.emit("    ldr x0, [x0]")
		}

	case *CallExpression:
		ident, ok := e.Function.(*Identifier)
		if ok {
			if ident.Name == "__ldrb" {
				// __ldrb(addr)
				cg.generateExpression(e.Arguments[0])
				cg.emit("    ldrb w0, [x0]")
				return
			}
			if ident.Name == "__strb" {
				// __strb(addr, val)
				cg.generateExpression(e.Arguments[1])
				cg.emit("    str x0, [sp, #-16]!") // Save val
				cg.generateExpression(e.Arguments[0]) // Addr
				cg.emit("    mov x1, x0") // Addr in x1
				cg.emit("    ldr x0, [sp], #16") // Val in x0
				cg.emit("    strb w0, [x1]")
				return
			}
		}

		for _, arg := range e.Arguments {
			cg.generateExpression(arg)
			cg.emit("    str x0, [sp, #-16]!")
		}
		for i := len(e.Arguments) - 1; i >= 0; i-- {
			cg.emit(fmt.Sprintf("    ldr x%d, [sp], #16", i))
		}
		if !ok {
			panic("Function call must be identifier")
		}
		cg.emit(fmt.Sprintf("    bl _%s", ident.Name))

	case *InfixExpression:
		cg.generateExpression(e.Left)
		cg.emit("    str x0, [sp, #-16]!")
		
cg.generateExpression(e.Right)
		cg.emit("    ldr x1, [sp], #16")
		
		switch e.Operator {
		case "+":
			cg.emit("    add x0, x1, x0")
		case "-":
			cg.emit("    sub x0, x1, x0")
		case "*":
			cg.emit("    mul x0, x1, x0")
		case "/":
			cg.emit("    sdiv x0, x1, x0")
		case "==":
			cg.emit("    cmp x1, x0")
			cg.emit("    cset x0, eq")
		case "!=":
			cg.emit("    cmp x1, x0")
			cg.emit("    cset x0, ne")
		case "<":
			cg.emit("    cmp x1, x0")
			cg.emit("    cset x0, lt")
		case ">":
			cg.emit("    cmp x1, x0")
			cg.emit("    cset x0, gt")
		case "<=":
			cg.emit("    cmp x1, x0")
			cg.emit("    cset x0, le")
		case ">=":
			cg.emit("    cmp x1, x0")
			cg.emit("    cset x0, ge")
		}
	}
}

func (cg *CodeGenerator) emit(s string) {
	cg.output.WriteString(s + "\n")
}