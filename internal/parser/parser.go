// Package parser implements a recursive descent parser for the AWSL language.
// It transforms a stream of tokens from the lexer into an abstract syntax tree.
package parser

import (
	"fmt"

	"github.com/boattime/awsl/internal/ast"
	"github.com/boattime/awsl/internal/lexer"
	"github.com/boattime/awsl/internal/token"
)

// MaxErrors is the maximum number of errors the parser will collect
// before giving up.
const MaxErrors = 20

// Error represents a parsing error with position information.
type Error struct {
	Message string
	Line    int
	Column  int
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("line %d, column %d: %s", e.Line, e.Column, e.Message)
}

// Parser performs syntactic analysis on AWSL source code.
// It consumes tokens from the lexer and produces an AST.
type Parser struct {
	lexer *lexer.Lexer

	curToken  token.Token // Current token being examined
	peekToken token.Token // Next token (one token lookahead)

	errors []*Error
}

// New creates a new Parser for the given lexer.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:  l,
		errors: []*Error{},
	}

	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()

	return p
}

// Errors returns the list of parsing errors encountered.
func (p *Parser) Errors() []*Error {
	return p.errors
}

// HasErrors returns true if any parsing errors were encountered.
func (p *Parser) HasErrors() bool {
	return len(p.errors) > 0
}

// nextToken advances to the next token in the input.
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

// curTokenIs reports whether the current token is of the given type.
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs reports whether the next token is of the given type.
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek checks if the next token is of the expected type.
// If so, it advances to that token and returns true.
// Otherwise, it records an error and returns false.
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// addError adds a parsing error with the given message and position.
func (p *Parser) addError(line, column int, format string, args ...any) {
	if len(p.errors) >= MaxErrors {
		return
	}

	err := &Error{
		Message: fmt.Sprintf(format, args...),
		Line:    line,
		Column:  column,
	}
	p.errors = append(p.errors, err)
}

// curError records an error at the current token position.
func (p *Parser) curError(format string, args ...any) {
	p.addError(p.curToken.Line, p.curToken.Column, format, args...)
}

// peekError records an error for an unexpected peek token.
func (p *Parser) peekError(expected token.TokenType) {
	p.addError(
		p.peekToken.Line,
		p.peekToken.Column,
		"expected %s, got %s",
		expected,
		p.peekToken.Type,
	)
}

// synchronize advances the parser to a synchronization point after an error.
// This allows the parser to continue and potentially find more errors.
// Synchronization points are statement boundaries: semicolons and keywords
// that start new statements.
func (p *Parser) synchronize() {
	for !p.curTokenIs(token.EOF) {
		// If we just passed a semicolon, we're at a statement boundary
		if p.curToken.Type == token.SEMICOLON {
			p.nextToken()
			return
		}

		// If the next token starts a new statement, stop here
		switch p.peekToken.Type {
		case token.FUNCTION,
			token.IF,
			token.FOR,
			token.RETURN,
			token.PROFILE,
			token.REGION:
			p.nextToken()
			return
		}

		p.nextToken()
	}
}

// ParseProgram parses the entire input and returns the AST.
// If parsing errors occur, they can be retrieved via Errors().
// The returned program may be partially complete if errors occurred.
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{
		Statements: []ast.Statement{},
	}

	for !p.curTokenIs(token.EOF) {
		if len(p.errors) >= MaxErrors {
			break
		}

		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}

	return program
}

// parseStatement parses a single statement based on the current token.
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.PROFILE, token.REGION:
		return p.parseContextStatement()
	case token.IF:
		return p.parseIfStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.FUNCTION:
		return p.parseFunctionDeclaration()
	case token.IDENT:
		// Could be assignment (x = ...) or expression statement (foo())
		if p.peekTokenIs(token.ASSIGN) {
			return p.parseAssignmentStatement()
		}
		return p.parseExpressionStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseContextStatement parses profile or region statements.
// Grammar: context_statement = ( "profile" | "region" ) string ";" ;
func (p *Parser) parseContextStatement() *ast.ContextStatement {
	stmt := &ast.ContextStatement{Token: p.curToken}

	// Expect string value
	if !p.expectPeek(token.STRING) {
		p.synchronize()
		return nil
	}
	stmt.Value = p.curToken.Literal

	// Expect semicolon
	if !p.expectPeek(token.SEMICOLON) {
		p.synchronize()
		return nil
	}

	p.nextToken() // Move past semicolon
	return stmt
}

// parseAssignmentStatement parses variable assignments.
// Grammar: assignment = identifier "=" expr ";" ;
func (p *Parser) parseAssignmentStatement() *ast.AssignmentStatement {
	stmt := &ast.AssignmentStatement{
		Token: p.curToken,
		Name: &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		},
	}

	// Move past identifier to '='
	if !p.expectPeek(token.ASSIGN) {
		p.synchronize()
		return nil
	}

	p.nextToken() // Move past '='

	stmt.Value = p.parseExpression()
	if stmt.Value == nil {
		p.synchronize()
		return nil
	}

	// Expect semicolon
	if !p.expectPeek(token.SEMICOLON) {
		p.synchronize()
		return nil
	}

	p.nextToken() // Move past semicolon
	return stmt
}

// parseIfStatement parses conditional statements.
// Grammar: if_statement = "if" "(" expr ")" block [ "else" block ] ;
func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}

	// Expect opening paren
	if !p.expectPeek(token.LPAREN) {
		p.synchronize()
		return nil
	}

	p.nextToken() // Move past '('

	stmt.Condition = p.parseExpression()
	if stmt.Condition == nil {
		p.synchronize()
		return nil
	}

	// Expect closing paren
	if !p.expectPeek(token.RPAREN) {
		p.synchronize()
		return nil
	}

	// Expect opening brace for consequence
	if !p.expectPeek(token.LBRACE) {
		p.synchronize()
		return nil
	}

	stmt.Consequence = p.parseBlockStatement()
	if stmt.Consequence == nil {
		return nil
	}

	// Check for optional else clause
	if p.curTokenIs(token.ELSE) {
		if !p.expectPeek(token.LBRACE) {
			p.synchronize()
			return nil
		}

		stmt.Alternative = p.parseBlockStatement()
		if stmt.Alternative == nil {
			return nil
		}
	}

	return stmt
}

// parseForStatement parses for-in loops.
// Grammar: for_statement = "for" "(" identifier "in" expr ")" block ;
func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken}

	// Expect opening paren
	if !p.expectPeek(token.LPAREN) {
		p.synchronize()
		return nil
	}

	// Expect iterator identifier
	if !p.expectPeek(token.IDENT) {
		p.synchronize()
		return nil
	}

	stmt.Iterator = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	// Expect 'in' keyword
	if !p.expectPeek(token.IN) {
		p.synchronize()
		return nil
	}

	p.nextToken() // Move past 'in'

	stmt.Iterable = p.parseExpression()
	if stmt.Iterable == nil {
		p.synchronize()
		return nil
	}

	// Expect closing paren
	if !p.expectPeek(token.RPAREN) {
		p.synchronize()
		return nil
	}

	// Expect opening brace for body
	if !p.expectPeek(token.LBRACE) {
		p.synchronize()
		return nil
	}

	stmt.Body = p.parseBlockStatement()
	if stmt.Body == nil {
		return nil
	}

	return stmt
}

// parseReturnStatement parses return statements.
// Grammar: return_statement = "return" [ expr ] ";" ;
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken() // Move past 'return'

	// Check for bare return (no expression)
	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken() // Move past semicolon
		return stmt
	}

	stmt.Value = p.parseExpression()
	if stmt.Value == nil {
		p.synchronize()
		return nil
	}

	// Expect semicolon
	if !p.expectPeek(token.SEMICOLON) {
		p.synchronize()
		return nil
	}

	p.nextToken() // Move past semicolon
	return stmt
}

// parseFunctionDeclaration parses function definitions.
// Grammar: function_decl = "fn" identifier "(" [ param_list ] ")" block ;
func (p *Parser) parseFunctionDeclaration() *ast.FunctionDeclaration {
	stmt := &ast.FunctionDeclaration{Token: p.curToken}

	// Expect function name
	if !p.expectPeek(token.IDENT) {
		p.synchronize()
		return nil
	}

	stmt.Name = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	// Expect opening paren
	if !p.expectPeek(token.LPAREN) {
		p.synchronize()
		return nil
	}

	stmt.Parameters = p.parseParameterList()

	// Expect closing paren (parseParameterList leaves us before it)
	if !p.expectPeek(token.RPAREN) {
		p.synchronize()
		return nil
	}

	// Expect opening brace for body
	if !p.expectPeek(token.LBRACE) {
		p.synchronize()
		return nil
	}

	stmt.Body = p.parseBlockStatement()
	if stmt.Body == nil {
		return nil
	}

	return stmt
}

// parseParameterList parses function parameter names.
// Grammar: param_list = identifier { "," identifier } ;
func (p *Parser) parseParameterList() []*ast.Identifier {
	params := []*ast.Identifier{}

	// Check for empty parameter list
	if p.peekTokenIs(token.RPAREN) {
		return params
	}

	// Parse first parameter
	if !p.expectPeek(token.IDENT) {
		return params
	}

	params = append(params, &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	})

	// Parse remaining parameters
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // Move to comma
		if !p.expectPeek(token.IDENT) {
			return params
		}
		params = append(params, &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		})
	}

	return params
}

// parseBlockStatement parses a block of statements.
// Grammar: block = "{" { statement } "}" ;
// Assumes curToken is '{' when called.
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{
		Token:      p.curToken,
		Statements: []ast.Statement{},
	}

	p.nextToken() // Move past '{'

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		if len(p.errors) >= MaxErrors {
			break
		}

		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
	}

	if !p.curTokenIs(token.RBRACE) {
		p.curError("expected }, got %s", p.curToken.Type)
		return nil
	}

	p.nextToken() // Move past '}'
	return block
}

// parseExpressionStatement parses an expression as a statement.
// Grammar: expr_statement = expr ";" ;
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression()
	if stmt.Expression == nil {
		p.synchronize()
		return nil
	}

	// Expect semicolon
	if !p.expectPeek(token.SEMICOLON) {
		p.synchronize()
		return nil
	}

	p.nextToken() // Move past semicolon
	return stmt
}

// parseExpression parses an expression.
// This is the entry point for expression parsing and starts at the lowest
// precedence level (equality).
// Grammar: expr = logic_or ;
func (p *Parser) parseExpression() ast.Expression {
	return p.parseOr()
}

// parseOr parses or expressions.
// Grammar: logic_or = logic_and { "||" logic_and } ;
func (p *Parser) parseOr() ast.Expression {
	left := p.parseAnd()
	if left == nil {
		return nil
	}

	for p.peekTokenIs(token.OR) {
		p.nextToken() // Move to or
		or := p.curToken

		p.nextToken() // Move past or
		right := p.parseAnd()
		if right == nil {
			return nil
		}

		left = &ast.InfixExpression{
			Token:    or,
			Left:     left,
			Operator: or.Literal,
			Right:    right,
		}
	}

	return left
}

// parseAnd parses and expressions.
// Grammar: logic_and = equality { "&&" equality } ;
func (p *Parser) parseAnd() ast.Expression {
	left := p.parseEquality()
	if left == nil {
		return nil
	}

	for p.peekTokenIs(token.AND) {
		p.nextToken() // Move to and
		and := p.curToken

		p.nextToken() // Move past and
		right := p.parseEquality()
		if right == nil {
			return nil
		}

		left = &ast.InfixExpression{
			Token:    and,
			Left:     left,
			Operator: and.Literal,
			Right:    right,
		}
	}

	return left
}

// parseEquality parses equality expressions.
// Grammar: equality = comparison { ( "==" | "!=" ) comparison } ;
func (p *Parser) parseEquality() ast.Expression {
	left := p.parseComparison()
	if left == nil {
		return nil
	}

	for p.peekTokenIs(token.EQ) || p.peekTokenIs(token.NOT_EQ) {
		p.nextToken() // Move to operator
		operator := p.curToken

		p.nextToken() // Move past operator
		right := p.parseComparison()
		if right == nil {
			return nil
		}

		left = &ast.InfixExpression{
			Token:    operator,
			Left:     left,
			Operator: operator.Literal,
			Right:    right,
		}
	}

	return left
}

// parseComparison parses comparison expressions.
// Grammar: comparison = term { ( "<" | ">" | "<=" | ">=" ) term } ;
func (p *Parser) parseComparison() ast.Expression {
	left := p.parseTerm()
	if left == nil {
		return nil
	}

	for p.peekTokenIs(token.LT) || p.peekTokenIs(token.GT) ||
		p.peekTokenIs(token.LTE) || p.peekTokenIs(token.GTE) {
		p.nextToken() // Move to operator
		operator := p.curToken

		p.nextToken() // Move past operator
		right := p.parseTerm()
		if right == nil {
			return nil
		}

		left = &ast.InfixExpression{
			Token:    operator,
			Left:     left,
			Operator: operator.Literal,
			Right:    right,
		}
	}

	return left
}

// parseTerm parses addition and subtraction expressions.
// Grammar: term = factor { ( "+" | "-" ) factor } ;
func (p *Parser) parseTerm() ast.Expression {
	left := p.parseFactor()
	if left == nil {
		return nil
	}

	for p.peekTokenIs(token.PLUS) || p.peekTokenIs(token.MINUS) {
		p.nextToken() // Move to operator
		operator := p.curToken

		p.nextToken() // Move past operator
		right := p.parseFactor()
		if right == nil {
			return nil
		}

		left = &ast.InfixExpression{
			Token:    operator,
			Left:     left,
			Operator: operator.Literal,
			Right:    right,
		}
	}

	return left
}

// parseFactor parses multiplication and division expressions.
// Grammar: factor = unary { ( "*" | "/" ) unary } ;
func (p *Parser) parseFactor() ast.Expression {
	left := p.parseUnary()
	if left == nil {
		return nil
	}

	for p.peekTokenIs(token.ASTERISK) || p.peekTokenIs(token.SLASH) {
		p.nextToken() // Move to operator
		operator := p.curToken

		p.nextToken() // Move past operator
		right := p.parseUnary()
		if right == nil {
			return nil
		}

		left = &ast.InfixExpression{
			Token:    operator,
			Left:     left,
			Operator: operator.Literal,
			Right:    right,
		}
	}

	return left
}

// parseUnary parses unary expressions (prefix operators).
// Grammar: unary = ( "!" | "-" ) unary | postfix ;
func (p *Parser) parseUnary() ast.Expression {
	if p.curTokenIs(token.BANG) || p.curTokenIs(token.MINUS) {
		operator := p.curToken

		p.nextToken() // Move past operator
		right := p.parseUnary()
		if right == nil {
			return nil
		}

		return &ast.PrefixExpression{
			Token:    operator,
			Operator: operator.Literal,
			Right:    right,
		}
	}

	return p.parsePostfix()
}

// parsePostfix parses postfix expressions (calls, index, member access, pipe).
// Grammar: postfix = primary { call | index | member | pipe } ;
func (p *Parser) parsePostfix() ast.Expression {
	left := p.parsePrimary()
	if left == nil {
		return nil
	}

	for {
		switch {
		case p.peekTokenIs(token.LPAREN):
			p.nextToken() // Move to '('
			left = p.parseCallExpression(left)
			if left == nil {
				return nil
			}

		case p.peekTokenIs(token.LBRACKET):
			p.nextToken() // Move to '['
			left = p.parseIndexExpression(left)
			if left == nil {
				return nil
			}

		case p.peekTokenIs(token.DOT):
			p.nextToken() // Move to '.'
			left = p.parseMemberExpression(left)
			if left == nil {
				return nil
			}

		case p.peekTokenIs(token.PIPE):
			p.nextToken() // Move to '|'
			left = p.parsePipeExpression(left)
			if left == nil {
				return nil
			}

		default:
			return left
		}
	}
}

// parseCallExpression parses a function call.
// Grammar: call = "(" [ arg_list ] ")" ;
// Assumes curToken is '(' when called.
func (p *Parser) parseCallExpression(function ast.Expression) *ast.CallExpression {
	expr := &ast.CallExpression{
		Token:    p.curToken,
		Function: function,
	}

	expr.Arguments = p.parseArgumentList()

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return expr
}

// parseArgumentList parses function call arguments.
// Grammar: arg_list = arg { "," arg } ;
//
//	arg = [ identifier ":" ] expr ;
func (p *Parser) parseArgumentList() []ast.Argument {
	args := []ast.Argument{}

	// Check for empty argument list
	if p.peekTokenIs(token.RPAREN) {
		return args
	}

	p.nextToken() // Move to first argument

	arg := p.parseArgument()
	if arg == nil {
		return args
	}
	args = append(args, *arg)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // Move to comma
		p.nextToken() // Move past comma

		arg := p.parseArgument()
		if arg == nil {
			return args
		}
		args = append(args, *arg)
	}

	return args
}

// parseArgument parses a single argument (positional or named).
// Grammar: arg = [ identifier ":" ] expr ;
func (p *Parser) parseArgument() *ast.Argument {
	// Check for named argument: identifier followed by colon
	if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.COLON) {
		name := &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}

		p.nextToken() // Move to ':'
		p.nextToken() // Move past ':'

		value := p.parseExpression()
		if value == nil {
			return nil
		}

		return &ast.Argument{
			Name:  name,
			Value: value,
		}
	}

	// Positional argument
	value := p.parseExpression()
	if value == nil {
		return nil
	}

	return &ast.Argument{
		Name:  nil,
		Value: value,
	}
}

// parseIndexExpression parses array/list index access.
// Grammar: index = "[" expr "]" ;
// Assumes curToken is '[' when called.
func (p *Parser) parseIndexExpression(left ast.Expression) *ast.IndexExpression {
	expr := &ast.IndexExpression{
		Token: p.curToken,
		Left:  left,
	}

	p.nextToken() // Move past '['

	expr.Index = p.parseExpression()
	if expr.Index == nil {
		return nil
	}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return expr
}

// parseMemberExpression parses member/property access.
// Grammar: member = "." identifier ;
// Assumes curToken is '.' when called.
func (p *Parser) parseMemberExpression(object ast.Expression) *ast.MemberExpression {
	expr := &ast.MemberExpression{
		Token:  p.curToken,
		Object: object,
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	expr.Member = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	return expr
}

// parsePipeExpression parses the pipe operator for formatting.
// Grammar: pipe = "|" "format" ( "csv" | "table" ) ;
// Assumes curToken is '|' when called.
func (p *Parser) parsePipeExpression(left ast.Expression) *ast.PipeExpression {
	expr := &ast.PipeExpression{
		Token: p.curToken,
		Left:  left,
	}

	// Expect 'format' identifier
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	if p.curToken.Literal != "format" {
		p.curError("expected 'format' after pipe, got %q", p.curToken.Literal)
		return nil
	}

	// Expect format type: 'csv' or 'table'
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	format := p.curToken.Literal
	if format != "csv" && format != "table" {
		p.curError("expected 'csv' or 'table', got %q", format)
		return nil
	}

	expr.Format = format

	return expr
}

// parsePrimary parses primary expressions (literals, identifiers, grouped).
// Grammar: primary = identifier | number | string | "true" | "false" | "null"
//
//	| "(" expr ")" | list_literal | object_literal ;
func (p *Parser) parsePrimary() ast.Expression {
	switch p.curToken.Type {
	case token.IDENT:
		return &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}

	case token.INT:
		return p.parseIntegerLiteral()

	case token.FLOAT:
		return p.parseFloatLiteral()

	case token.STRING:
		return &ast.StringLiteral{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}

	case token.TRUE:
		return &ast.BooleanLiteral{
			Token: p.curToken,
			Value: true,
		}

	case token.FALSE:
		return &ast.BooleanLiteral{
			Token: p.curToken,
			Value: false,
		}

	case token.NULL:
		return &ast.NullLiteral{
			Token: p.curToken,
		}

	case token.LPAREN:
		return p.parseGroupedExpression()

	case token.LBRACKET:
		return p.parseListLiteral()

	case token.LBRACE:
		return p.parseObjectLiteral()

	default:
		p.curError("unexpected token %s", p.curToken.Type)
		return nil
	}
}

// parseIntegerLiteral parses an integer literal.
func (p *Parser) parseIntegerLiteral() *ast.IntegerLiteral {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	var value int64
	_, err := fmt.Sscanf(p.curToken.Literal, "%d", &value)
	if err != nil {
		p.curError("could not parse %q as integer", p.curToken.Literal)
		return nil
	}

	lit.Value = value
	return lit
}

// parseFloatLiteral parses a floating-point literal.
func (p *Parser) parseFloatLiteral() *ast.FloatLiteral {
	lit := &ast.FloatLiteral{Token: p.curToken}

	var value float64
	_, err := fmt.Sscanf(p.curToken.Literal, "%f", &value)
	if err != nil {
		p.curError("could not parse %q as float", p.curToken.Literal)
		return nil
	}

	lit.Value = value
	return lit
}

// parseGroupedExpression parses a parenthesized expression.
// Assumes curToken is '(' when called.
func (p *Parser) parseGroupedExpression() *ast.GroupedExpression {
	expr := &ast.GroupedExpression{Token: p.curToken}

	p.nextToken() // Move past '('

	expr.Expression = p.parseExpression()
	if expr.Expression == nil {
		return nil
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return expr
}

// parseListLiteral parses a list/array literal.
// Grammar: list_literal = "[" [ expr { "," expr } ] "]" ;
// Assumes curToken is '[' when called.
func (p *Parser) parseListLiteral() *ast.ListLiteral {
	lit := &ast.ListLiteral{
		Token:    p.curToken,
		Elements: []ast.Expression{},
	}

	// Check for empty list
	if p.peekTokenIs(token.RBRACKET) {
		p.nextToken() // Move to ']'
		return lit
	}

	p.nextToken() // Move to first element

	elem := p.parseExpression()
	if elem == nil {
		return nil
	}
	lit.Elements = append(lit.Elements, elem)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // Move to comma
		p.nextToken() // Move past comma

		elem := p.parseExpression()
		if elem == nil {
			return nil
		}
		lit.Elements = append(lit.Elements, elem)
	}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return lit
}

// parseObjectLiteral parses an object literal.
// Grammar: object_literal = "{" [ pair { "," pair } ] "}" ;
//
//	pair = identifier ":" expr ;
//
// Assumes curToken is '{' when called.
func (p *Parser) parseObjectLiteral() *ast.ObjectLiteral {
	lit := &ast.ObjectLiteral{
		Token: p.curToken,
		Pairs: []ast.ObjectPair{},
	}

	// Check for empty object
	if p.peekTokenIs(token.RBRACE) {
		p.nextToken() // Move to '}'
		return lit
	}

	// Parse first pair
	pair := p.parseObjectPair()
	if pair == nil {
		return nil
	}
	lit.Pairs = append(lit.Pairs, *pair)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // Move past comma

		pair := p.parseObjectPair()
		if pair == nil {
			return nil
		}
		lit.Pairs = append(lit.Pairs, *pair)
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return lit
}

// parseObjectPair parses a key-value pair in an object literal.
// Grammar: pair = identifier ":" expr ;
func (p *Parser) parseObjectPair() *ast.ObjectPair {
	// Expect identifier key
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	pair := &ast.ObjectPair{
		Key: &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		},
	}

	// Expect colon
	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken() // Move past ':'

	pair.Value = p.parseExpression()
	if pair.Value == nil {
		return nil
	}

	return pair
}
