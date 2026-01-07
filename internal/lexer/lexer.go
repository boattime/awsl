// Package lexer implements lexical analysis for the AWSL language.
// It converts source code into a stream of tokens that can be
// consumed by the parser.
package lexer

import (
	"github.com/boattime/awsl/internal/token"
)

// Lexer performs lexical analysis on AWSL source code.
// It maintains position tracking for error reporting and
// converts the input string into a sequence of tokens.
type Lexer struct {
	input        string // source code being tokenized
	position     int    // current position in input (points to current char)
	readPosition int    // next reading position in input (after current char)
	ch           byte   // current character under examination
	line         int    // current line number (1-based)
	column       int    // current column number (1-based)
}

// New creates a new Lexer instance for the given input string.
// The lexer is initialized and ready to produce tokens via NextToken.
func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// NextToken scans and returns the next token from the input.
// It skips whitespace and comments, then identifies the token type
// based on the current character(s). Returns an EOF token when
// the input is exhausted.
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespaceAndComments()

	// Record position at the start of the token
	startLine := l.line
	startColumn := l.column

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: "==", Line: startLine, Column: startColumn}
		} else {
			tok = newToken(token.ASSIGN, l.ch, startLine, startColumn)
		}
	case '+':
		tok = newToken(token.PLUS, l.ch, startLine, startColumn)
	case '-':
		tok = newToken(token.MINUS, l.ch, startLine, startColumn)
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: "!=", Line: startLine, Column: startColumn}
		} else {
			tok = newToken(token.BANG, l.ch, startLine, startColumn)
		}
	case '*':
		tok = newToken(token.ASTERISK, l.ch, startLine, startColumn)
	case '/':
		tok = newToken(token.SLASH, l.ch, startLine, startColumn)
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.LTE, Literal: "<=", Line: startLine, Column: startColumn}
		} else {
			tok = newToken(token.LT, l.ch, startLine, startColumn)
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.GTE, Literal: ">=", Line: startLine, Column: startColumn}
		} else {
			tok = newToken(token.GT, l.ch, startLine, startColumn)
		}
	case ',':
		tok = newToken(token.COMMA, l.ch, startLine, startColumn)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch, startLine, startColumn)
	case ':':
		tok = newToken(token.COLON, l.ch, startLine, startColumn)
	case '.':
		tok = newToken(token.DOT, l.ch, startLine, startColumn)
	case '|':
		tok = newToken(token.PIPE, l.ch, startLine, startColumn)
	case '(':
		tok = newToken(token.LPAREN, l.ch, startLine, startColumn)
	case ')':
		tok = newToken(token.RPAREN, l.ch, startLine, startColumn)
	case '{':
		tok = newToken(token.LBRACE, l.ch, startLine, startColumn)
	case '}':
		tok = newToken(token.RBRACE, l.ch, startLine, startColumn)
	case '[':
		tok = newToken(token.LBRACKET, l.ch, startLine, startColumn)
	case ']':
		tok = newToken(token.RBRACKET, l.ch, startLine, startColumn)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		tok.Line = startLine
		tok.Column = startColumn
	case 0:
		tok.Type = token.EOF
		tok.Literal = ""
		tok.Line = startLine
		tok.Column = startColumn
		return tok
	default:
		if isLetter(l.ch) {
			literal := l.readIdentifier()
			return token.Token{
				Type:    token.LookupIdent(literal),
				Literal: literal,
				Line:    startLine,
				Column:  startColumn,
			}
		} else if isDigit(l.ch) {
			literal, tokenType := l.readNumber()
			return token.Token{
				Type:    tokenType,
				Literal: literal,
				Line:    startLine,
				Column:  startColumn,
			}
		} else {
			tok = newToken(token.ILLEGAL, l.ch, startLine, startColumn)
		}
	}

	l.readChar()
	return tok
}

// readChar advances the lexer to the next character in the input.
// It updates position tracking and handles line/column counting.
// When the end of input is reached, ch is set to 0 (NULL).
func (l *Lexer) readChar() {
	// Update line/column based on the character we're moving past
	if l.ch == '\n' {
		l.line++
		l.column = 0
	}

	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
}

// peekChar returns the next character without advancing the lexer position.
// Returns 0 if at end of input.
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// skipWhitespaceAndComments advances past whitespace and single-line comments.
// Comments start with // and continue to the end of the line.
func (l *Lexer) skipWhitespaceAndComments() {
	for {
		// Skip whitespace
		for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
			l.readChar()
		}

		// Skip single-line comments
		if l.ch == '/' && l.peekChar() == '/' {
			l.skipLineComment()
		} else {
			break
		}
	}
}

// skipLineComment advances past a single-line comment.
// It assumes the lexer is positioned at the first '/'.
func (l *Lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

// readIdentifier reads an identifier starting at the current position.
// Identifiers consist of letters, digits, and underscores, but must
// start with a letter or underscore.
func (l *Lexer) readIdentifier() string {
	startPosition := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[startPosition:l.position]
}

// readNumber reads a numeric literal (integer or float) starting at
// the current position. It returns the literal string and the
// appropriate token type (INT or FLOAT).
func (l *Lexer) readNumber() (string, token.TokenType) {
	startPosition := l.position
	tokenType := token.INT

	// Read integer part
	for isDigit(l.ch) {
		l.readChar()
	}

	// Check for decimal point followed by digits (float)
	if l.ch == '.' && isDigit(l.peekChar()) {
		tokenType = token.FLOAT
		l.readChar() // consume the '.'

		// Read fractional part
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[startPosition:l.position], tokenType
}

// readString reads a string literal, returning the content without
// the surrounding quotes. It assumes the lexer is positioned at the
// opening quote.
func (l *Lexer) readString() string {
	startPosition := l.position + 1 // Start after opening quote

	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}

	return l.input[startPosition:l.position]
}

// newToken creates a token from a single character.
func newToken(tokenType token.TokenType, ch byte, line, column int) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: string(ch),
		Line:    line,
		Column:  column,
	}
}

// isLetter reports whether the character is a letter or underscore.
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// isDigit reports whether the character is a decimal digit.
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
