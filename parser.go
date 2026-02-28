package main

import (
	"fmt"
	"strconv"
)

// Precedence levels for the Pratt Parser.
// Higher values bind more tightly.
const (
	_ int = iota
	LOWEST
	ASSIGN_PREC // =
	EQUALS      // == !=
	LESSGREATER // > < >= <=
	SUM         // + -
	PRODUCT     // * /
	PREFIX      // -X or !X or *X or &X
	CALL        // myFunction(x)
	INDEX       // array[index]
)

// ============================================================================
// Abstract Syntax Tree (AST) Nodes
// ============================================================================

// Program is the root node, containing a list of functions and global variables.
type Program struct {
	Functions []*Function
	Globals   []*VarStatement
}

// Function represents a function definition (name, return type, params, body).
type Function struct {
	Name       string
	ReturnType string 
	Parameters []*VarStatement 
	Body       *Block
}

// Block represents a { ... } block containing statements.
type Block struct {
	Statements []Statement
}

// Statement Interface
type Statement interface {
	statementNode()
}

type ReturnStatement struct {
	Value Expression
}
func (rs *ReturnStatement) statementNode() {}

type VarStatement struct {
	Name  string
	Type  string 
	Size  int        // If > 0, it's an array
	Value Expression // Optional initial value
}
func (vs *VarStatement) statementNode() {}

type IfStatement struct {
	Condition   Expression
	Consequence *Block
	Alternative *Block
}
func (is *IfStatement) statementNode() {}

type WhileStatement struct {
	Condition Expression
	Body      *Block
}
func (ws *WhileStatement) statementNode() {}

type ForStatement struct {
	Init      Statement
	Condition Expression
	Post      Statement
	Body      *Block
}
func (fs *ForStatement) statementNode() {}

// ExpressionStatement allows expressions (like assignments or calls) to stand alone.
type ExpressionStatement struct {
	Expression Expression
}
func (es *ExpressionStatement) statementNode() {}

// Expression Interface
type Expression interface {
	expressionNode()
}

type IntegerLiteral struct {
	Value int
}
func (il *IntegerLiteral) expressionNode() {}

type StringLiteral struct {
	Value string
}
func (sl *StringLiteral) expressionNode() {}

type Identifier struct {
	Name string
}
func (id *Identifier) expressionNode() {}

// PrefixExpression handles unary operators: -x, !x, *p, &x
type PrefixExpression struct {
	Operator string
	Right    Expression
}
func (pe *PrefixExpression) expressionNode() {}

// CallExpression handles function calls: foo(a, b)
type CallExpression struct {
	Function  Expression 
	Arguments []Expression
}
func (ce *CallExpression) expressionNode() {}

// IndexExpression handles array access: a[i]
type IndexExpression struct {
	Left  Expression
	Index Expression
}
func (ie *IndexExpression) expressionNode() {}

// InfixExpression handles binary operators: a + b, a == b
type InfixExpression struct {
	Left     Expression
	Operator string
	Right    Expression
}
func (ie *InfixExpression) expressionNode() {}

// AssignExpression handles assignment: a = b
type AssignExpression struct {
	Left  Expression // Can be Identifier or Dereference (*p) or Index (a[i])
	Value Expression
}
func (ae *AssignExpression) expressionNode() {}

// ============================================================================
// Parser
// ============================================================================

type Parser struct {
	l         *Lexer
	curToken  Token
	peekToken Token
	errors    []string
}

func NewParser(l *Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func isType(t TokenType) bool {
	return t == INT || t == CHAR || t == VOID
}

// ParseProgram entry point. Loops until EOF.
func (p *Parser) ParseProgram() *Program {
	prog := &Program{}
	prog.Functions = []*Function{}
	prog.Globals = []*VarStatement{}
	
	for p.curToken.Type != EOF {
		// All top-level declarations (func or var) start with a Type (int, char, void)
		if !isType(p.curToken.Type) {
			p.errors = append(p.errors, fmt.Sprintf("expected type at top level, got %s", p.curToken.Type))
			p.nextToken()
			continue
		}
		p.parseTopLevel(prog)
	}
	return prog
}

// parseTopLevel distinguishes between Global Variables and Function Definitions.
func (p *Parser) parseTopLevel(prog *Program) {
	typeName := p.parseType()
	
	if !p.expectPeek(IDENT) {
		return
	}
	name := p.curToken.Value
	
	if p.peekToken.Type == LPAREN {
		// Function Definition
		fn := p.parseFunctionRest(name, typeName)
		if fn != nil {
			prog.Functions = append(prog.Functions, fn)
		}
	} else if p.peekToken.Type == ASSIGN || p.peekToken.Type == SEMICOLON || p.peekToken.Type == LBRACKET {
		// Global Variable Declaration
		stmt := p.parseVarStatementRest(name, typeName)
		if stmt != nil {
			prog.Globals = append(prog.Globals, stmt)
		}
	} else {
		p.errors = append(p.errors, fmt.Sprintf("unexpected token at top level: %s", p.peekToken.Type))
		p.nextToken()
	}
	p.nextToken()
}

// parseType parses types like "int", "char*", "int**".
func (p *Parser) parseType() string {
	t := p.curToken.Value
	for p.peekToken.Type == ASTERISK {
		p.nextToken()
		t += "*"
	}
	return t
}

func (p *Parser) parseFunctionRest(name, returnType string) *Function {
	fn := &Function{Name: name, ReturnType: returnType}
	
	if !p.expectPeek(LPAREN) {
		return nil
	}
	
	fn.Parameters = p.parseFunctionParameters()
	
	// Check for prototype (semicolon instead of brace)
	if p.peekToken.Type == SEMICOLON {
		p.nextToken() // consume ;
		return nil // Ignore prototype
	}
	
	if !p.expectPeek(LBRACE) {
		return nil
	}
	
	fn.Body = p.parseBlock()
	return fn
}

func (p *Parser) parseVarStatementRest(name, typeName string) *VarStatement {
	stmt := &VarStatement{Name: name, Type: typeName}
	
	// Array? int a[10]
	if p.peekToken.Type == LBRACKET {
		p.nextToken() 
		p.nextToken() 
		if p.curToken.Type != NUMBER {
			p.errors = append(p.errors, "expected array size")
			return nil
		}
		size, _ := strconv.Atoi(p.curToken.Value)
		stmt.Size = size
		
		if !p.expectPeek(RBRACKET) {
			return nil
		}
	}
	
	// Initialization? int a = 5
	if p.peekToken.Type == ASSIGN {
		p.nextToken() 
		p.nextToken() 
		stmt.Value = p.parseExpression(LOWEST)
	} else if stmt.Size == 0 && stmt.Value == nil {
		// Default init for global scalar
		stmt.Value = &IntegerLiteral{Value: 0}
	}
	
	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
	}
	
	return stmt
}

func (p *Parser) parseFunctionParameters() []*VarStatement {
	params := []*VarStatement{}
	
	if p.peekToken.Type == RPAREN {
		p.nextToken()
		return params
	}
	
	p.nextToken()
	if isType(p.curToken.Type) {
		t := p.parseType()
		if !p.expectPeek(IDENT) {
			return nil
		}
		params = append(params, &VarStatement{Name: p.curToken.Value, Type: t})
	}
	
	for p.peekToken.Type == COMMA {
		p.nextToken()
		p.nextToken()
		if !isType(p.curToken.Type) {
			return nil
		}
		t := p.parseType()
		if !p.expectPeek(IDENT) {
			return nil
		}
		params = append(params, &VarStatement{Name: p.curToken.Value, Type: t})
	}
	
	if !p.expectPeek(RPAREN) {
		return nil
	}
	
	return params
}

func (p *Parser) parseBlock() *Block {
	block := &Block{}
	block.Statements = []Statement{}
	p.nextToken() 
	
	for p.curToken.Type != RBRACE && p.curToken.Type != EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseStatement() Statement {
	switch p.curToken.Type {
	case RETURN:
		return p.parseReturnStatement()
	case INT, CHAR, VOID:
		return p.parseVarStatement()
	case IF:
		return p.parseIfStatement()
	case WHILE:
		return p.parseWhileStatement()
	case FOR:
		return p.parseForStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseReturnStatement() *ReturnStatement {
	stmt := &ReturnStatement{}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseVarStatement() *VarStatement {
	t := p.parseType()
	
	stmt := &VarStatement{Type: t}
	if !p.expectPeek(IDENT) {
		return nil
	}
	stmt.Name = p.curToken.Value
	
	if p.peekToken.Type == LBRACKET {
		p.nextToken()
		p.nextToken()
		if p.curToken.Type != NUMBER {
			return nil
		}
		size, _ := strconv.Atoi(p.curToken.Value)
		stmt.Size = size
		if !p.expectPeek(RBRACKET) {
			return nil
		}
	}
	
	if p.peekToken.Type == ASSIGN {
		p.nextToken()
		p.nextToken()
		stmt.Value = p.parseExpression(LOWEST)
	}
	
	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseIfStatement() *IfStatement {
	stmt := &IfStatement{}
	if !p.expectPeek(LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(RPAREN) {
		return nil
	}
	if !p.expectPeek(LBRACE) {
		return nil
	}
	stmt.Consequence = p.parseBlock()
	
	if p.peekToken.Type == ELSE {
		p.nextToken()
		if !p.expectPeek(LBRACE) {
			return nil
		}
		stmt.Alternative = p.parseBlock()
	}
	return stmt
}

func (p *Parser) parseWhileStatement() *WhileStatement {
	stmt := &WhileStatement{}
	if !p.expectPeek(LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(RPAREN) {
		return nil
	}
	if !p.expectPeek(LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlock()
	return stmt
}

func (p *Parser) parseForStatement() *ForStatement {
	stmt := &ForStatement{}
	if !p.expectPeek(LPAREN) {
		return nil
	}
	p.nextToken() 
	
	// Init part (optional)
	if p.curToken.Type != SEMICOLON {
		stmt.Init = p.parseStatement()
		// If parseStatement consumed semicolon, we need to check if we are at semicolon now.
		// parseVarStatement consumes semicolon.
		// So curToken is semicolon.
		if p.curToken.Type == SEMICOLON {
			p.nextToken()
		}
	} else {
		p.nextToken() // consume empty init semicolon
	}
	
	// Condition part (optional)
	if p.curToken.Type != SEMICOLON {
		stmt.Condition = p.parseExpression(LOWEST)
	}
	if !p.expectPeek(SEMICOLON) {
		return nil
	}
	p.nextToken() 
	
	// Post part (optional)
	if p.curToken.Type != RPAREN {
		exp := p.parseExpression(LOWEST)
		stmt.Post = &ExpressionStatement{Expression: exp}
	}
	
	if !p.expectPeek(RPAREN) {
		return nil
	}
	if !p.expectPeek(LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlock()
	
	return stmt
}

func (p *Parser) parseExpressionStatement() *ExpressionStatement {
	stmt := &ExpressionStatement{}
	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
	}
	return stmt
}

// parseExpression uses Pratt Parsing.
func (p *Parser) parseExpression(precedence int) Expression {
	// 1. Parse prefix (e.g. -x, !x, IDENT, NUMBER)
	left := p.parsePrefix()
	if left == nil {
		return nil
	}
	// 2. Loop while next operator has higher precedence
	for p.peekToken.Type != SEMICOLON && precedence < p.peekPrecedence() {
		p.nextToken()
		left = p.parseInfix(left)
	}
	return left
}

func (p *Parser) parsePrefix() Expression {
	switch p.curToken.Type {
	case NUMBER:
		val, _ := strconv.Atoi(p.curToken.Value)
		return &IntegerLiteral{Value: val}
	case STRING:
		return &StringLiteral{Value: p.curToken.Value}
	case IDENT:
		return &Identifier{Name: p.curToken.Value}
	case LPAREN:
		p.nextToken()
		exp := p.parseExpression(LOWEST)
		if !p.expectPeek(RPAREN) {
			return nil
		}
		return exp
	case MINUS, AMPERSAND, ASTERISK:
		// Prefix operators: - (negate), & (addr), * (deref)
		op := p.curToken.Value
		p.nextToken()
		right := p.parseExpression(PREFIX)
		return &PrefixExpression{Operator: op, Right: right}
	default:
		return nil
	}
}

func (p *Parser) parseInfix(left Expression) Expression {
	if p.curToken.Type == LPAREN {
		return p.parseCallExpression(left)
	}
	
	if p.curToken.Type == LBRACKET {
		return p.parseIndexExpression(left)
	}

	op := p.curToken.Value
	
	if p.curToken.Type == ASSIGN {
		p.nextToken()
		val := p.parseExpression(LOWEST)
		return &AssignExpression{Left: left, Value: val}
	}

	precedence := p.curPrecedence()
	p.nextToken()
	right := p.parseExpression(precedence)
	return &InfixExpression{Left: left, Operator: op, Right: right}
}

func (p *Parser) parseIndexExpression(left Expression) Expression {
	exp := &IndexExpression{Left: left}
	p.nextToken() // [
	exp.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(RBRACKET) {
		return nil
	}
	return exp
}

func (p *Parser) parseCallExpression(function Expression) *CallExpression {
	exp := &CallExpression{Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []Expression {
	args := []Expression{}
	if p.peekToken.Type == RPAREN {
		p.nextToken()
		return args
	}
	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))
	for p.peekToken.Type == COMMA {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(RPAREN) {
		return nil
	}
	return args
}

func (p *Parser) peekPrecedence() int {
	return getPrecedence(p.peekToken.Type)
}

func (p *Parser) curPrecedence() int {
	return getPrecedence(p.curToken.Type)
}

func getPrecedence(t TokenType) int {
	switch t {
	case LBRACKET:
		return INDEX
	case LPAREN:
		return CALL
	case ASSIGN:
		return ASSIGN_PREC
	case EQ, NOT_EQ:
		return EQUALS
	case LT, GT, LTE, GTE:
		return LESSGREATER
	case PLUS, MINUS:
		return SUM
	case ASTERISK, SLASH:
		return PRODUCT
	default:
		return LOWEST
	}
}

func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	} else {
		p.errors = append(p.errors, fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type))
		return false
	}
}