package eval

import "testing"

func TestIntegerObject(t *testing.T) {
	tests := []struct {
		value           int64
		expectedType    ObjectType
		expectedInspect string
	}{
		{42, INTEGER_OBJ, "42"},
		{0, INTEGER_OBJ, "0"},
		{-5, INTEGER_OBJ, "-5"},
		{9223372036854775807, INTEGER_OBJ, "9223372036854775807"},
	}

	for _, tt := range tests {
		obj := &Integer{Value: tt.value}

		if obj.Type() != tt.expectedType {
			t.Errorf("Integer.Type() = %q, want %q", obj.Type(), tt.expectedType)
		}

		if obj.Inspect() != tt.expectedInspect {
			t.Errorf("Integer.Inspect() = %q, want %q", obj.Inspect(), tt.expectedInspect)
		}
	}
}

func TestFloatObject(t *testing.T) {
	tests := []struct {
		value           float64
		expectedType    ObjectType
		expectedInspect string
	}{
		{3.14, FLOAT_OBJ, "3.14"},
		{0.0, FLOAT_OBJ, "0"},
		{-2.5, FLOAT_OBJ, "-2.5"},
		{100.001, FLOAT_OBJ, "100.001"},
	}

	for _, tt := range tests {
		obj := &Float{Value: tt.value}

		if obj.Type() != tt.expectedType {
			t.Errorf("Float.Type() = %q, want %q", obj.Type(), tt.expectedType)
		}

		if obj.Inspect() != tt.expectedInspect {
			t.Errorf("Float.Inspect() = %q, want %q", obj.Inspect(), tt.expectedInspect)
		}
	}
}

func TestStringObject(t *testing.T) {
	tests := []struct {
		value           string
		expectedType    ObjectType
		expectedInspect string
	}{
		{"hello", STRING_OBJ, "hello"},
		{"", STRING_OBJ, ""},
		{"hello world", STRING_OBJ, "hello world"},
		{"us-west-2", STRING_OBJ, "us-west-2"},
	}

	for _, tt := range tests {
		obj := &String{Value: tt.value}

		if obj.Type() != tt.expectedType {
			t.Errorf("String.Type() = %q, want %q", obj.Type(), tt.expectedType)
		}

		if obj.Inspect() != tt.expectedInspect {
			t.Errorf("String.Inspect() = %q, want %q", obj.Inspect(), tt.expectedInspect)
		}
	}
}

func TestBooleanObject(t *testing.T) {
	tests := []struct {
		value           bool
		expectedType    ObjectType
		expectedInspect string
	}{
		{true, BOOLEAN_OBJ, "true"},
		{false, BOOLEAN_OBJ, "false"},
	}

	for _, tt := range tests {
		obj := &Boolean{Value: tt.value}

		if obj.Type() != tt.expectedType {
			t.Errorf("Boolean.Type() = %q, want %q", obj.Type(), tt.expectedType)
		}

		if obj.Inspect() != tt.expectedInspect {
			t.Errorf("Boolean.Inspect() = %q, want %q", obj.Inspect(), tt.expectedInspect)
		}
	}
}

func TestBooleanSingletons(t *testing.T) {
	if TRUE.Type() != BOOLEAN_OBJ {
		t.Errorf("TRUE.Type() = %q, want %q", TRUE.Type(), BOOLEAN_OBJ)
	}
	if TRUE.Value != true {
		t.Errorf("TRUE.Value = %t, want true", TRUE.Value)
	}
	if TRUE.Inspect() != "true" {
		t.Errorf("TRUE.Inspect() = %q, want %q", TRUE.Inspect(), "true")
	}

	if FALSE.Type() != BOOLEAN_OBJ {
		t.Errorf("FALSE.Type() = %q, want %q", FALSE.Type(), BOOLEAN_OBJ)
	}
	if FALSE.Value != false {
		t.Errorf("FALSE.Value = %t, want false", FALSE.Value)
	}
	if FALSE.Inspect() != "false" {
		t.Errorf("FALSE.Inspect() = %q, want %q", FALSE.Inspect(), "false")
	}

	t1 := TRUE
	t2 := TRUE
	if t1 != t2 {
		t.Error("TRUE singleton should return same instance")
	}

	f1 := FALSE
	f2 := FALSE
	if f1 != f2 {
		t.Error("FALSE singleton should return same instance")
	}
}

func TestNullObject(t *testing.T) {
	obj := &Null{}

	if obj.Type() != NULL_OBJ {
		t.Errorf("Null.Type() = %q, want %q", obj.Type(), NULL_OBJ)
	}

	if obj.Inspect() != "null" {
		t.Errorf("Null.Inspect() = %q, want %q", obj.Inspect(), "null")
	}
}

func TestNullSingleton(t *testing.T) {
	if NULL.Type() != NULL_OBJ {
		t.Errorf("NULL.Type() = %q, want %q", NULL.Type(), NULL_OBJ)
	}

	if NULL.Inspect() != "null" {
		t.Errorf("NULL.Inspect() = %q, want %q", NULL.Inspect(), "null")
	}

	n1 := NULL
	n2 := NULL
	if n1 != n2 {
		t.Error("NULL singleton should return same instance")
	}
}

func TestErrorObject(t *testing.T) {
	tests := []struct {
		message         string
		line            int
		column          int
		expectedType    ObjectType
		expectedInspect string
	}{
		{
			"undefined variable: x",
			5,
			12,
			ERROR_OBJ,
			"error at line 5, column 12: undefined variable: x",
		},
		{
			"type mismatch: INTEGER + STRING",
			10,
			1,
			ERROR_OBJ,
			"error at line 10, column 1: type mismatch: INTEGER + STRING",
		},
		{
			"division by zero",
			1,
			1,
			ERROR_OBJ,
			"error at line 1, column 1: division by zero",
		},
	}

	for _, tt := range tests {
		obj := &Error{
			Message: tt.message,
			Line:    tt.line,
			Column:  tt.column,
		}

		if obj.Type() != tt.expectedType {
			t.Errorf("Error.Type() = %q, want %q", obj.Type(), tt.expectedType)
		}

		if obj.Inspect() != tt.expectedInspect {
			t.Errorf("Error.Inspect() = %q, want %q", obj.Inspect(), tt.expectedInspect)
		}
	}
}

func TestListElements(t *testing.T) {
	elements := []Object{
		&Integer{Value: 10},
		&Integer{Value: 20},
		&Integer{Value: 30},
	}
	list := &List{Elements: elements}

	if len(list.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(list.Elements))
	}

	for i, expected := range []int64{10, 20, 30} {
		integer, ok := list.Elements[i].(*Integer)
		if !ok {
			t.Fatalf("element %d is not *Integer, got %T", i, list.Elements[i])
		}
		if integer.Value != expected {
			t.Errorf("element %d = %d, want %d", i, integer.Value, expected)
		}
	}
}

func TestObjectInterface(t *testing.T) {
	objects := []Object{
		&Integer{Value: 42},
		&Float{Value: 3.14},
		&String{Value: "hello"},
		&Boolean{Value: true},
		&Null{},
		&Error{Message: "test", Line: 1, Column: 1},
		&List{Elements: []Object{&Integer{Value: 1}}},
		TRUE,
		FALSE,
		NULL,
	}

	for _, obj := range objects {
		_ = obj.Type()
		_ = obj.Inspect()
	}
}

func TestBuiltinObject(t *testing.T) {
	fn := func(env *Environment, args ...Object) Object { return NULL }
	obj := &Builtin{Name: "test", Fn: fn}

	if obj.Type() != BUILTIN_OBJ {
		t.Errorf("Builtin.Type() = %q, want %q", obj.Type(), BUILTIN_OBJ)
	}

	if obj.Inspect() != "builtin:test" {
		t.Errorf("Builtin.Inspect() = %q, want %q", obj.Inspect(), "builtin:test")
	}
}
