package object

import (
	"bytes"
	"fmt"
	"interpreter/ast"
	"strings"
)

const (
	BUILTIN_OBJECT      = "BUILTIN"
	STRING_OBJECT       = "STRING"
	FUNCTION_OBJECT     = "FUNCTION"
	INTEGER_OBJECT      = "INTEGER"
	BOOLEAN_OBJECT      = "BOOLEAN"
	NULL_OBJECT         = "NULL"
	RETURN_VALUE_OBJECT = "RETURN_VALUE"
	ERROR_OBJECT        = "ERROR"
	ARRAY_OBJECT        = "ARRAY"
)

type Array struct {
	Elements []Object
}

func (arrayObject *Array) Type() string { return ARRAY_OBJECT }
func (arrayObject *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range arrayObject.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type Error struct {
	Message string
}

func (error *Error) Type() string { return ERROR_OBJECT }
func (error *Error) Inspect() string {
	return "ERROR: " + error.Message
}

type Object interface {
	Type() string
	Inspect() string
}

type Environment struct {
	store map[string]Object
	outer *Environment
}

func NewEnvironment() *Environment {
	store := make(map[string]Object)
	return &Environment{store: store, outer: nil}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (environment *Environment) Get(name string) (Object, bool) {
	obj, ok := environment.store[name]
	if !ok && environment.outer != nil {
		obj, ok = environment.outer.Get(name)
	}
	return obj, ok
}

func (environment *Environment) Set(name string, val Object) Object {
	environment.store[name] = val
	return val
}

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (function *Function) Type() string { return FUNCTION_OBJECT }
func (function *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}

	for _, p := range function.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(function.Body.String())
	out.WriteString("\n}")

	return out.String()
}

type ReturnValue struct {
	Value Object
}

func (returnValue *ReturnValue) Type() string    { return RETURN_VALUE_OBJECT }
func (returnValue *ReturnValue) Inspect() string { return returnValue.Value.Inspect() }

type Integer struct {
	Value int64
}

func (integer *Integer) Inspect() string { return fmt.Sprintf("%d", integer.Value) }
func (integer *Integer) Type() string    { return INTEGER_OBJECT }

type String struct {
	Value string
}

func (string *String) Inspect() string { return string.Value }
func (string *String) Type() string    { return STRING_OBJECT }

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (builtin *Builtin) Type() string    { return BUILTIN_OBJECT }
func (builtin *Builtin) Inspect() string { return "builtin function" }

type Boolean struct {
	Value bool
}

func (boolean *Boolean) Inspect() string { return fmt.Sprintf("%t", boolean.Value) }
func (boolean *Boolean) Type() string    { return BOOLEAN_OBJECT }

type Null struct {
}

func (null *Null) Inspect() string { return "null" }
func (null *Null) Type() string    { return NULL_OBJECT }
