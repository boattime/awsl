// Package ast defines the abstract syntax tree nodes for the AWSL language.
// The AST is produced by the parser and consumed by the interpreter.
package ast

import (
	"strings"

	"github.com/boattime/awsl/internal/token"
)

// Position represents a location in source code.
type Position struct {
	Line   int // 1-based line number
	Column int // 1-based column number
}

// Node represents any node in the abstract syntax tree.
// All AST nodes must implement this interface.
type Node interface {
	// Pos returns the position of the first character of the node.
	Pos() Position

	// String returns a string representation of the node
	// for debugging and testing purposes.
	String() string
}

// Statement represents a statement node in the AST.
// Statements do not produce values directly.
type Statement interface {
	Node
	// statementNode is a marker method to distinguish statements from expressions.
	statementNode()
}

// Expression represents an expression node in the AST.
// Expressions produce values when evaluated.
type Expression interface {
	Node
	// expressionNode is a marker method to distinguish expressions from statements.
	expressionNode()
}

// Program represents the root node of every AWSL program.
// A program consists of a sequence of statements.
type Program struct {
	Statements []Statement
}

// Pos returns the position of the first statement,
// or line 1, column 1 if the program is empty.
func (p *Program) Pos() Position {
	if len(p.Statements) > 0 {
		return p.Statements[0].Pos()
	}
	return Position{Line: 1, Column: 1}
}

// String returns the program as a string by concatenating
// all statement strings.
func (p *Program) String() string {
	var out strings.Builder
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// ExpressionStatement wraps an expression as a statement.
// This allows expressions to be used where statements are expected,
// such as function calls: print("hello");
type ExpressionStatement struct {
	Token      token.Token // The first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}

// Pos returns the position of the expression.
func (es *ExpressionStatement) Pos() Position {
	return Position{Line: es.Token.Line, Column: es.Token.Column}
}

// String returns the expression as a string.
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// AssignmentStatement represents a variable assignment: identifier = expression;
type AssignmentStatement struct {
	Token token.Token // The identifier token
	Name  *Identifier
	Value Expression
}

func (as *AssignmentStatement) statementNode() {}

// Pos returns the position of the identifier.
func (as *AssignmentStatement) Pos() Position {
	return Position{Line: as.Token.Line, Column: as.Token.Column}
}

// String returns the assignment as a string.
func (as *AssignmentStatement) String() string {
	var out strings.Builder
	out.WriteString(as.Name.String())
	out.WriteString(" = ")
	if as.Value != nil {
		out.WriteString(as.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// ContextStatement represents profile or region context setters.
// Examples: profile "production"; region "us-west-2";
type ContextStatement struct {
	Token token.Token // PROFILE or REGION token
	Value string      // The string value (without quotes)
}

func (cs *ContextStatement) statementNode() {}

// Pos returns the position of the context keyword.
func (cs *ContextStatement) Pos() Position {
	return Position{Line: cs.Token.Line, Column: cs.Token.Column}
}

// String returns the context statement as a string.
func (cs *ContextStatement) String() string {
	var out strings.Builder
	out.WriteString(cs.Token.Literal)
	out.WriteString(" \"")
	out.WriteString(cs.Value)
	out.WriteString("\";")
	return out.String()
}

// BlockStatement represents a block of statements enclosed in braces.
// Example: { statement1; statement2; }
type BlockStatement struct {
	Token      token.Token // The '{' token
	Statements []Statement
}

func (bs *BlockStatement) statementNode() {}

// Pos returns the position of the opening brace.
func (bs *BlockStatement) Pos() Position {
	return Position{Line: bs.Token.Line, Column: bs.Token.Column}
}

// String returns the block as a string.
func (bs *BlockStatement) String() string {
	var out strings.Builder
	out.WriteString("{ ")
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	out.WriteString(" }")
	return out.String()
}

// IfStatement represents a conditional statement.
// Example: if (condition) { ... } else { ... }
type IfStatement struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement // May be nil if no else clause
}

func (is *IfStatement) statementNode() {}

// Pos returns the position of the if keyword.
func (is *IfStatement) Pos() Position {
	return Position{Line: is.Token.Line, Column: is.Token.Column}
}

// String returns the if statement as a string.
func (is *IfStatement) String() string {
	var out strings.Builder
	out.WriteString("if (")
	out.WriteString(is.Condition.String())
	out.WriteString(") ")
	out.WriteString(is.Consequence.String())
	if is.Alternative != nil {
		out.WriteString(" else ")
		out.WriteString(is.Alternative.String())
	}
	return out.String()
}

// ForStatement represents a for-in loop.
// Example: for (item in collection) { ... }
type ForStatement struct {
	Token    token.Token // The 'for' token
	Iterator *Identifier // The loop variable
	Iterable Expression  // The collection being iterated
	Body     *BlockStatement
}

func (fs *ForStatement) statementNode() {}

// Pos returns the position of the for keyword.
func (fs *ForStatement) Pos() Position {
	return Position{Line: fs.Token.Line, Column: fs.Token.Column}
}

// String returns the for statement as a string.
func (fs *ForStatement) String() string {
	var out strings.Builder
	out.WriteString("for (")
	out.WriteString(fs.Iterator.String())
	out.WriteString(" in ")
	out.WriteString(fs.Iterable.String())
	out.WriteString(") ")
	out.WriteString(fs.Body.String())
	return out.String()
}

// ReturnStatement represents a return statement.
// Example: return value; or return;
type ReturnStatement struct {
	Token token.Token // The 'return' token
	Value Expression  // May be nil for bare return
}

func (rs *ReturnStatement) statementNode() {}

// Pos returns the position of the return keyword.
func (rs *ReturnStatement) Pos() Position {
	return Position{Line: rs.Token.Line, Column: rs.Token.Column}
}

// String returns the return statement as a string.
func (rs *ReturnStatement) String() string {
	var out strings.Builder
	out.WriteString("return")
	if rs.Value != nil {
		out.WriteString(" ")
		out.WriteString(rs.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// FunctionDeclaration represents a function definition.
// Example: fn name(param1, param2) { ... }
type FunctionDeclaration struct {
	Token      token.Token // The 'fn' token
	Name       *Identifier
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fd *FunctionDeclaration) statementNode() {}

// Pos returns the position of the fn keyword.
func (fd *FunctionDeclaration) Pos() Position {
	return Position{Line: fd.Token.Line, Column: fd.Token.Column}
}

// String returns the function declaration as a string.
func (fd *FunctionDeclaration) String() string {
	var out strings.Builder
	out.WriteString("fn ")
	out.WriteString(fd.Name.String())
	out.WriteString("(")
	params := make([]string, len(fd.Parameters))
	for i, p := range fd.Parameters {
		params[i] = p.String()
	}
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fd.Body.String())
	return out.String()
}

// Identifier represents a variable or function name.
type Identifier struct {
	Token token.Token // The IDENT token
	Value string
}

func (i *Identifier) expressionNode() {}

// Pos returns the position of the identifier.
func (i *Identifier) Pos() Position {
	return Position{Line: i.Token.Line, Column: i.Token.Column}
}

// String returns the identifier value.
func (i *Identifier) String() string {
	return i.Value
}

// IntegerLiteral represents an integer value.
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}

// Pos returns the position of the integer.
func (il *IntegerLiteral) Pos() Position {
	return Position{Line: il.Token.Line, Column: il.Token.Column}
}

// String returns the integer as a string.
func (il *IntegerLiteral) String() string {
	return il.Token.Literal
}

// FloatLiteral represents a floating-point value.
type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode() {}

// Pos returns the position of the float.
func (fl *FloatLiteral) Pos() Position {
	return Position{Line: fl.Token.Line, Column: fl.Token.Column}
}

// String returns the float as a string.
func (fl *FloatLiteral) String() string {
	return fl.Token.Literal
}

// StringLiteral represents a string value.
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode() {}

// Pos returns the position of the string.
func (sl *StringLiteral) Pos() Position {
	return Position{Line: sl.Token.Line, Column: sl.Token.Column}
}

// String returns the string with quotes.
func (sl *StringLiteral) String() string {
	return "\"" + sl.Value + "\""
}

// BooleanLiteral represents a boolean value (true or false).
type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode() {}

// Pos returns the position of the boolean.
func (bl *BooleanLiteral) Pos() Position {
	return Position{Line: bl.Token.Line, Column: bl.Token.Column}
}

// String returns "true" or "false".
func (bl *BooleanLiteral) String() string {
	return bl.Token.Literal
}

// NullLiteral represents the null value.
type NullLiteral struct {
	Token token.Token
}

func (nl *NullLiteral) expressionNode() {}

// Pos returns the position of the null keyword.
func (nl *NullLiteral) Pos() Position {
	return Position{Line: nl.Token.Line, Column: nl.Token.Column}
}

// String returns "null".
func (nl *NullLiteral) String() string {
	return "null"
}

// PrefixExpression represents a prefix operator expression.
// Examples: !condition, -number
type PrefixExpression struct {
	Token    token.Token // The prefix token (! or -)
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode() {}

// Pos returns the position of the operator.
func (pe *PrefixExpression) Pos() Position {
	return Position{Line: pe.Token.Line, Column: pe.Token.Column}
}

// String returns the prefix expression as a string.
func (pe *PrefixExpression) String() string {
	var out strings.Builder
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

// InfixExpression represents a binary operator expression.
// Examples: a + b, x == y
type InfixExpression struct {
	Token    token.Token // The operator token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode() {}

// Pos returns the position of the left operand.
func (ie *InfixExpression) Pos() Position {
	return ie.Left.Pos()
}

// String returns the infix expression as a string.
func (ie *InfixExpression) String() string {
	var out strings.Builder
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" ")
	out.WriteString(ie.Operator)
	out.WriteString(" ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// CallExpression represents a function or method call.
// Examples: print("hello"), lambda.list(runtime: "python3.12")
type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or MemberExpression
	Arguments []Argument
}

func (ce *CallExpression) expressionNode() {}

// Pos returns the position of the function being called.
func (ce *CallExpression) Pos() Position {
	return ce.Function.Pos()
}

// String returns the call expression as a string.
func (ce *CallExpression) String() string {
	var out strings.Builder
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	args := make([]string, len(ce.Arguments))
	for i, a := range ce.Arguments {
		args[i] = a.String()
	}
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// Argument represents a function call argument.
// It may be positional (Name is nil) or named (Name is set).
type Argument struct {
	Name  *Identifier // nil for positional arguments
	Value Expression
}

// String returns the argument as a string.
func (a *Argument) String() string {
	if a.Name != nil {
		return a.Name.String() + ": " + a.Value.String()
	}
	return a.Value.String()
}

// IndexExpression represents array/list index access.
// Example: items[0], data[i + 1]
type IndexExpression struct {
	Token token.Token // The '[' token
	Left  Expression  // The expression being indexed
	Index Expression  // The index expression
}

func (ie *IndexExpression) expressionNode() {}

// Pos returns the position of the expression being indexed.
func (ie *IndexExpression) Pos() Position {
	return ie.Left.Pos()
}

// String returns the index expression as a string.
func (ie *IndexExpression) String() string {
	var out strings.Builder
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")
	return out.String()
}

// MemberExpression represents member/property access.
// Example: user.name, lambda.list
type MemberExpression struct {
	Token  token.Token // The '.' token
	Object Expression  // The object being accessed
	Member *Identifier // The member name
}

func (me *MemberExpression) expressionNode() {}

// Pos returns the position of the object being accessed.
func (me *MemberExpression) Pos() Position {
	return me.Object.Pos()
}

// String returns the member expression as a string.
func (me *MemberExpression) String() string {
	var out strings.Builder
	out.WriteString("(")
	out.WriteString(me.Object.String())
	out.WriteString(".")
	out.WriteString(me.Member.String())
	out.WriteString(")")
	return out.String()
}

// PipeExpression represents the pipe operator for formatting.
// Example: items | format csv, data | format table
type PipeExpression struct {
	Token  token.Token // The '|' token
	Left   Expression  // The expression being piped
	Format string      // "csv" or "table"
}

func (pe *PipeExpression) expressionNode() {}

// Pos returns the position of the expression being piped.
func (pe *PipeExpression) Pos() Position {
	return pe.Left.Pos()
}

// String returns the pipe expression as a string.
func (pe *PipeExpression) String() string {
	var out strings.Builder
	out.WriteString("(")
	out.WriteString(pe.Left.String())
	out.WriteString(" | format ")
	out.WriteString(pe.Format)
	out.WriteString(")")
	return out.String()
}

// ListLiteral represents a list/array literal.
// Example: [1, 2, 3], ["a", "b"]
type ListLiteral struct {
	Token    token.Token // The '[' token
	Elements []Expression
}

func (ll *ListLiteral) expressionNode() {}

// Pos returns the position of the opening bracket.
func (ll *ListLiteral) Pos() Position {
	return Position{Line: ll.Token.Line, Column: ll.Token.Column}
}

// String returns the list literal as a string.
func (ll *ListLiteral) String() string {
	var out strings.Builder
	out.WriteString("[")
	elements := make([]string, len(ll.Elements))
	for i, e := range ll.Elements {
		elements[i] = e.String()
	}
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

// ObjectLiteral represents an object literal.
// Example: {name: "test", count: 5}
type ObjectLiteral struct {
	Token token.Token  // The '{' token
	Pairs []ObjectPair // Ordered list of key-value pairs
}

func (ol *ObjectLiteral) expressionNode() {}

// Pos returns the position of the opening brace.
func (ol *ObjectLiteral) Pos() Position {
	return Position{Line: ol.Token.Line, Column: ol.Token.Column}
}

// String returns the object literal as a string.
func (ol *ObjectLiteral) String() string {
	var out strings.Builder
	out.WriteString("{")
	pairs := make([]string, len(ol.Pairs))
	for i, p := range ol.Pairs {
		pairs[i] = p.String()
	}
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

// ObjectPair represents a key-value pair in an object literal.
type ObjectPair struct {
	Key   *Identifier
	Value Expression
}

// String returns the pair as a string.
func (op *ObjectPair) String() string {
	return op.Key.String() + ": " + op.Value.String()
}

// GroupedExpression represents a parenthesized expression.
// Example: (a + b) * c
type GroupedExpression struct {
	Token      token.Token // The '(' token
	Expression Expression
}

func (ge *GroupedExpression) expressionNode() {}

// Pos returns the position of the opening parenthesis.
func (ge *GroupedExpression) Pos() Position {
	return Position{Line: ge.Token.Line, Column: ge.Token.Column}
}

// String returns the grouped expression as a string.
func (ge *GroupedExpression) String() string {
	var out strings.Builder
	out.WriteString("(")
	out.WriteString(ge.Expression.String())
	out.WriteString(")")
	return out.String()
}
