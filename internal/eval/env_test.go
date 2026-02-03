package eval

import (
	"os"
	"testing"
)

func TestEnvironment_SetAndGet(t *testing.T) {
	env := NewEnvironment(os.Stdout)

	env.Set("x", &Integer{Value: 42})

	val, ok := env.Get("x")
	if !ok {
		t.Fatal("expected to find 'x' in environment")
	}

	intVal, ok := val.(*Integer)
	if !ok {
		t.Fatalf("expected *Integer, got %T", val)
	}

	if intVal.Value != 42 {
		t.Errorf("expected value 42, got %d", intVal.Value)
	}
}

func TestEnvironment_GetUndefined(t *testing.T) {
	env := NewEnvironment(os.Stdout)

	val, ok := env.Get("undefined")
	if ok {
		t.Error("expected 'undefined' to not be found")
	}

	if val != nil {
		t.Errorf("expected nil for undefined variable, got %v", val)
	}
}

func TestEnvironment_Overwrite(t *testing.T) {
	env := NewEnvironment(os.Stdout)

	env.Set("x", &Integer{Value: 1})
	env.Set("x", &Integer{Value: 2})

	val, _ := env.Get("x")
	if val.(*Integer).Value != 2 {
		t.Errorf("expected value 2 after overwrite, got %d", val.(*Integer).Value)
	}
}

func TestEnvironment_ReadFromOuter(t *testing.T) {
	outer := NewEnvironment(os.Stdout)
	outer.Set("x", &Integer{Value: 10})

	inner := NewEnclosedEnvironment(outer)

	// Inner can read outer's variable
	val, ok := inner.Get("x")
	if !ok {
		t.Fatal("expected inner to find 'x' from outer")
	}
	if val.(*Integer).Value != 10 {
		t.Errorf("expected 10, got %d", val.(*Integer).Value)
	}
}

func TestEnvironment_ShadowOuter(t *testing.T) {
	outer := NewEnvironment(os.Stdout)
	outer.Set("x", &Integer{Value: 10})

	inner := NewEnclosedEnvironment(outer)
	inner.SetLocal("x", &Integer{Value: 99})

	// Inner sees shadowed value
	val, _ := inner.Get("x")
	if val.(*Integer).Value != 99 {
		t.Errorf("expected inner x=99, got %d", val.(*Integer).Value)
	}

	// Outer is unchanged
	val, _ = outer.Get("x")
	if val.(*Integer).Value != 10 {
		t.Errorf("expected outer x=10, got %d", val.(*Integer).Value)
	}
}

func TestEnvironment_InnerNotVisibleToOuter(t *testing.T) {
	outer := NewEnvironment(os.Stdout)
	inner := NewEnclosedEnvironment(outer)
	inner.Set("y", &Integer{Value: 5})

	_, ok := outer.Get("y")
	if ok {
		t.Error("outer should not see inner's variable")
	}
}

func TestEnvironment_MultipleNestingLevels(t *testing.T) {
	global := NewEnvironment(os.Stdout)
	global.Set("a", &Integer{Value: 1})

	level1 := NewEnclosedEnvironment(global)
	level1.Set("b", &Integer{Value: 2})

	level2 := NewEnclosedEnvironment(level1)
	level2.Set("c", &Integer{Value: 3})

	if val, ok := level2.Get("a"); !ok || val.(*Integer).Value != 1 {
		t.Error("level2 should access global 'a'")
	}
	if val, ok := level2.Get("b"); !ok || val.(*Integer).Value != 2 {
		t.Error("level2 should access level1 'b'")
	}
	if val, ok := level2.Get("c"); !ok || val.(*Integer).Value != 3 {
		t.Error("level2 should access own 'c'")
	}

	if _, ok := level1.Get("c"); ok {
		t.Error("level1 should not see level2 'c'")
	}
}
