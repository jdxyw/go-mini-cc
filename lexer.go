package main

type TokenType string

const (
	INT        TokenType = "INT"
	CHAR       TokenType = "CHAR"
	VOID       TokenType = "VOID"
	RETURN     TokenType = "RETURN"
	IF         TokenType = "IF"
	ELSE       TokenType = "ELSE"
	WHILE      TokenType = "WHILE"
	FOR        TokenType = "FOR"
	
	LPAREN     TokenType = "LPAREN"
	RPAREN     TokenType = "RPAREN"
	LBRACE     TokenType = "LBRACE"
	RBRACE     TokenType = "RBRACE"
	LBRACKET   TokenType = "LBRACKET"
	RBRACKET   TokenType = "RBRACKET"
	SEMICOLON  TokenType = "SEMICOLON"
	COMMA      TokenType = "COMMA"
	ASSIGN     TokenType = "ASSIGN"
	
	PLUS       TokenType = "PLUS"
	MINUS      TokenType = "MINUS"
	ASTERISK   TokenType = "ASTERISK"
	SLASH      TokenType = "SLASH"
	AMPERSAND  TokenType = "AMPERSAND"
	
	EQ         TokenType = "EQ"     // ==
	NOT_EQ     TokenType = "NOT_EQ" // !=
	LT         TokenType = "LT"     // <
	GT         TokenType = "GT"     // >
	LTE        TokenType = "LTE"    // <=
	GTE        TokenType = "GTE"    // >=
	
	NUMBER     TokenType = "NUMBER"
	IDENT      TokenType = "IDENT"
	STRING     TokenType = "STRING"
	EOF        TokenType = "EOF"
)

type Token struct {
	Type  TokenType
	Value string
}

type Lexer struct {
	input   string
	pos     int
	readPos int
	ch      byte
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	switch l.ch {
	case '(':
		tok = newToken(LPAREN, l.ch)
	case ')':
		tok = newToken(RPAREN, l.ch)
	case '{':
		tok = newToken(LBRACE, l.ch)
	case '}':
		tok = newToken(RBRACE, l.ch)
	case '[':
		tok = newToken(LBRACKET, l.ch)
	case ']':
		tok = newToken(RBRACKET, l.ch)
	case ';':
		tok = newToken(SEMICOLON, l.ch)
	case ',':
		tok = newToken(COMMA, l.ch)
	case '+':
		tok = newToken(PLUS, l.ch)
	case '-':
		tok = newToken(MINUS, l.ch)
	case '*':
		tok = newToken(ASTERISK, l.ch)
	case '/':
		tok = newToken(SLASH, l.ch)
	case '&':
		tok = newToken(AMPERSAND, l.ch)
	case '"':
		tok.Type = STRING
		tok.Value = l.readString()
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: EQ, Value: string(ch) + string(l.ch)}
		} else {
			tok = newToken(ASSIGN, l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: NOT_EQ, Value: string(ch) + string(l.ch)}
		} else {
			tok = newToken(EOF, l.ch) 
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: LTE, Value: string(ch) + string(l.ch)}
		} else {
			tok = newToken(LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: GTE, Value: string(ch) + string(l.ch)}
		} else {
			tok = newToken(GT, l.ch)
		}
	case 0:
		tok.Type = EOF
		tok.Value = ""
	default:
		if isLetter(l.ch) {
			tok.Value = l.readIdentifier()
			tok.Type = LookupIdent(tok.Value)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = NUMBER
			tok.Value = l.readNumber()
			return tok
		} else {
			tok = newToken(EOF, l.ch)
		}
	}

	l.readChar()
	return tok
}

func newToken(tokenType TokenType, ch byte) Token {
	return Token{Type: tokenType, Value: string(ch)}
}

func (l *Lexer) readString() string {
	position := l.pos + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.pos]
}

func (l *Lexer) readIdentifier() string {
	position := l.pos
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.pos]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func (l *Lexer) readNumber() string {
	position := l.pos
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.pos]
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func LookupIdent(ident string) TokenType {
	keywords := map[string]TokenType{
		"int":    INT,
		"char":   CHAR,
		"void":   VOID,
		"return": RETURN,
		"if":     IF,
		"else":   ELSE,
		"while":  WHILE,
		"for":    FOR,
	}
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}