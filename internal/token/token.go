// Package token defines the token types and structures used by the lexer
// and parser for lexical analysis.
package token

// TokenType represents the type of a lexical token as a string.
type TokenType string

// Token represents a lexical token with its type, literal value,
// and position information in the source code.
type Token struct {
	// Type indicates what kind of token this is.
	Type TokenType
	// Literal contains the actual string value from the source.
	Literal string
	// Line is the 1-based line number where the token appears.
	Line int
	// Column is the 1-based column number where the token starts.
	Column int
}

// Token types for special tokens.
const (
	// ILLEGAL represents an unknown or invalid token.
	ILLEGAL TokenType = "ILLEGAL"
	// EOF represents the end of the input.
	EOF TokenType = "EOF"
)

// Token types for identifiers and literals.
const (
	// IDENT represents an identifier (variable name, function name, etc.).
	IDENT TokenType = "IDENT"
	// INT represents an integer literal.
	INT TokenType = "INT"
	// FLOAT represents a floating-point literal.
	FLOAT TokenType = "FLOAT"
	// STRING represents a string literal.
	STRING TokenType = "STRING"
)

// Token types for operators.
const (
	ASSIGN   TokenType = "="  // Assignment operator
	PLUS     TokenType = "+"  // Addition operator
	MINUS    TokenType = "-"  // Subtraction operator
	BANG     TokenType = "!"  // Logical NOT operator
	ASTERISK TokenType = "*"  // Multiplication operator
	SLASH    TokenType = "/"  // Division operator
	LT       TokenType = "<"  // Less than operator
	GT       TokenType = ">"  // Greater than operator
	EQ       TokenType = "==" // Equality operator
	NOT_EQ   TokenType = "!=" // Inequality operator
	LTE      TokenType = "<=" // Less than or equal operator
	GTE      TokenType = ">=" // Greater than or equal operator
	OR       TokenType = "||" // Logical OR operator
	AND      TokenType = "&&" // Logical AND operator
)

// Token types for delimiters.
const (
	COMMA     TokenType = ","
	SEMICOLON TokenType = ";"
	COLON     TokenType = ":"
	DOT       TokenType = "."
	PIPE      TokenType = "|"

	LPAREN   TokenType = "("
	RPAREN   TokenType = ")"
	LBRACE   TokenType = "{"
	RBRACE   TokenType = "}"
	LBRACKET TokenType = "["
	RBRACKET TokenType = "]"
)

// Token types for keywords.
const (
	FUNCTION TokenType = "FUNCTION"
	TRUE     TokenType = "TRUE"
	FALSE    TokenType = "FALSE"
	NULL     TokenType = "NULL"
	IF       TokenType = "IF"
	ELSE     TokenType = "ELSE"
	FOR      TokenType = "FOR"
	IN       TokenType = "IN"
	RETURN   TokenType = "RETURN"
	PROFILE  TokenType = "PROFILE"
	REGION   TokenType = "REGION"
)

// keywords maps keyword strings to their corresponding TokenType.
var keywords = map[string]TokenType{
	"fn":      FUNCTION,
	"true":    TRUE,
	"false":   FALSE,
	"null":    NULL,
	"if":      IF,
	"else":    ELSE,
	"for":     FOR,
	"in":      IN,
	"return":  RETURN,
	"profile": PROFILE,
	"region":  REGION,
}

// LookupIdent checks if the given identifier is a keyword.
// If it is, it returns the keyword's TokenType.
// Otherwise, it returns IDENT.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
