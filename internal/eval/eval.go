// Package evaluator implements the tree-walking interpreter for AWSL.
// It takes an AST produced by the parser and evaluates it to produce
// runtime values.
package eval

import (
	"fmt"
	"github.com/boattime/awsl/internal/ast"
	"github.com/boattime/awsl/internal/token"
)

// Eval evaluates an AST node and returns the resulting runtime value.
// If evaluation fails, it returns an Error object.
func Eval(node ast.Node, env *Environment) Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.AssignmentStatement:
		return evalAssignment(node, env)

	// Literals
	case *ast.IntegerLiteral:
		return &Integer{Value: node.Value}
	case *ast.FloatLiteral:
		return &Float{Value: node.Value}
	case *ast.StringLiteral:
		return &String{Value: node.Value}
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.NullLiteral:
		return NULL

	// Expressions
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.CallExpression:
		return evalCallExpression(node, env)
	case *ast.PrefixExpression:
		return evalPrefixExpression(node, env)
	case *ast.InfixExpression:
		return evalInfixExpression(node, env)
	case *ast.GroupedExpression:
		return Eval(node.Expression, env)
	}

	pos := node.Pos()
	return newError(pos.Line, pos.Column, "unknown node type: %T", node)
}

// evalProgram evaluates all statements in a program and returns
// the result of the last statement.
func evalProgram(program *ast.Program, env *Environment) Object {
	var result Object = NULL

	for _, stmt := range program.Statements {
		result = Eval(stmt, env)

		// Stop evaluation if we hit an error
		if isError(result) {
			return result
		}
	}

	return result
}

// evalAssignment evaluates an assignment statement and stores
// the result in the environment.
func evalAssignment(node *ast.AssignmentStatement, env *Environment) Object {
	val := Eval(node.Value, env)
	if isError(val) {
		return val
	}

	env.Set(node.Name.Value, val)
	return NULL
}

// evalIdentifier looks up a variable in the environment.
func evalIdentifier(node *ast.Identifier, env *Environment) Object {
	val, ok := env.Get(node.Value)
	if !ok {
		pos := node.Pos()
		return newError(pos.Line, pos.Column, "undefined variable: %s", node.Value)
	}
	return val
}

// evalCallExpression evaluates a function call.
func evalCallExpression(node *ast.CallExpression, env *Environment) Object {
	function := Eval(node.Function, env)
	if isError(function) {
		return function
	}

	args, err := evalArguments(node.Arguments, env)
	if err != nil {
		return err
	}

	return applyFunction(function, args, node.Pos())
}

// evalArguments evaluates a list of arguments left to right.
func evalArguments(arguments []ast.Argument, env *Environment) ([]Object, *Error) {
	result := make([]Object, len(arguments))

	for i, arg := range arguments {
		evaluated := Eval(arg.Value, env)
		if isError(evaluated) {
			return nil, evaluated.(*Error)
		}
		result[i] = evaluated
	}

	return result, nil
}

// applyFunction calls a function with the given arguments.
func applyFunction(fn Object, args []Object, pos ast.Position) Object {
	switch function := fn.(type) {
	case *Builtin:
		return function.Fn(args...)
	default:
		return newError(pos.Line, pos.Column, "not a function: %s", fn.Type())
	}
}

// evalPrefixExpression evaluates prefix operators (! and -).
func evalPrefixExpression(node *ast.PrefixExpression, env *Environment) Object {
	right := Eval(node.Right, env)
	if isError(right) {
		return right
	}

	pos := node.Pos()

	switch node.Token.Type {
	case token.BANG:
		return evalBangOperator(right)
	case token.MINUS:
		return evalMinusPrefixOperator(right, pos)
	default:
		return newError(pos.Line, pos.Column, "unknown operator: %s%s", node.Token.Literal, right.Type())
	}
}

// evalBangOperator evaluates the ! operator.
// !true → false, !false → true, !null → true, anything else → false
func evalBangOperator(right Object) Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

// evalMinusPrefixOperator evaluates the unary minus operator.
func evalMinusPrefixOperator(right Object, pos ast.Position) Object {
	switch right := right.(type) {
	case *Integer:
		return &Integer{Value: -right.Value}
	case *Float:
		return &Float{Value: -right.Value}
	default:
		return newError(pos.Line, pos.Column, "unknown operator: -%s", right.Type())
	}
}

// evalInfixExpression evaluates binary operators.
func evalInfixExpression(node *ast.InfixExpression, env *Environment) Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	right := Eval(node.Right, env)
	if isError(right) {
		return right
	}

	pos := node.Pos()
	op := node.Token.Type

	switch {
	case left.Type() == INTEGER_OBJ && right.Type() == INTEGER_OBJ:
		return evalIntegerInfixExpression(op, left, right, pos)
	case left.Type() == FLOAT_OBJ && right.Type() == FLOAT_OBJ:
		return evalFloatInfixExpression(op, left, right, pos)
	case left.Type() == STRING_OBJ && right.Type() == STRING_OBJ:
		return evalStringInfixExpression(op, left, right, pos)
	case op == token.EQ:
		return nativeBoolToBooleanObject(left == right)
	case op == token.NOT_EQ:
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError(pos.Line, pos.Column, "type mismatch: %s %s %s", left.Type(), node.Token.Literal, right.Type())
	default:
		return newError(pos.Line, pos.Column, "unknown operator: %s %s %s", left.Type(), node.Token.Literal, right.Type())
	}
}

// evalIntegerInfixExpression evaluates binary operators on integers.
func evalIntegerInfixExpression(op token.TokenType, left, right Object, pos ast.Position) Object {
	leftVal := left.(*Integer).Value
	rightVal := right.(*Integer).Value

	switch op {
	case token.PLUS:
		return &Integer{Value: leftVal + rightVal}
	case token.MINUS:
		return &Integer{Value: leftVal - rightVal}
	case token.ASTERISK:
		return &Integer{Value: leftVal * rightVal}
	case token.SLASH:
		if rightVal == 0 {
			return newError(pos.Line, pos.Column, "division by zero")
		}
		return &Integer{Value: leftVal / rightVal}
	case token.LT:
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case token.GT:
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case token.LTE:
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case token.GTE:
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case token.EQ:
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case token.NOT_EQ:
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError(pos.Line, pos.Column, "unknown operator: INTEGER %s INTEGER", op)
	}
}

// evalFloatInfixExpression evaluates binary operators on floats.
func evalFloatInfixExpression(op token.TokenType, left, right Object, pos ast.Position) Object {
	leftVal := left.(*Float).Value
	rightVal := right.(*Float).Value

	switch op {
	case token.PLUS:
		return &Float{Value: leftVal + rightVal}
	case token.MINUS:
		return &Float{Value: leftVal - rightVal}
	case token.ASTERISK:
		return &Float{Value: leftVal * rightVal}
	case token.SLASH:
		if rightVal == 0 {
			return newError(pos.Line, pos.Column, "division by zero")
		}
		return &Float{Value: leftVal / rightVal}
	case token.LT:
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case token.GT:
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case token.LTE:
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case token.GTE:
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case token.EQ:
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case token.NOT_EQ:
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError(pos.Line, pos.Column, "unknown operator: FLOAT %s FLOAT", op)
	}
}

// evalStringInfixExpression evaluates binary operators on strings.
func evalStringInfixExpression(op token.TokenType, left, right Object, pos ast.Position) Object {
	leftVal := left.(*String).Value
	rightVal := right.(*String).Value

	switch op {
	case token.PLUS:
		return &String{Value: leftVal + rightVal}
	case token.EQ:
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case token.NOT_EQ:
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError(pos.Line, pos.Column, "unknown operator: STRING %s STRING", op)
	}
}

// nativeBoolToBooleanObject converts a Go bool to the appropriate singleton.
func nativeBoolToBooleanObject(value bool) *Boolean {
	if value {
		return TRUE
	}
	return FALSE
}

// newError creates a new Error object with position information.
func newError(line, column int, format string, args ...any) *Error {
	return &Error{
		Message: fmt.Sprintf(format, args...),
		Line:    line,
		Column:  column,
	}
}

// isError checks if an object is an Error.
func isError(obj Object) bool {
	return obj != nil && obj.Type() == ERROR_OBJ
}
