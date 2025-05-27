package evaluator

import (
	"fmt"
	"interpreter/ast"
	"interpreter/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return EvalProgram(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.Boolean:
		if node.Value {
			return TRUE
		}
		return FALSE
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if IsError(right) {
			return right
		}
		return EvalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		right := Eval(node.Right, env)
		if IsError(left) {
			return left
		}
		if IsError(right) {
			return right
		}
		return EvalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return EvalBlockStatement(node, env)
	case *ast.IfExpression:
		return EvalIfExpression(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if IsError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if IsError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.Identifier:
		return EvalIdentifier(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if IsError(function) {
			return function
		}
		args := EvalExpressions(node.Arguments, env)
		if len(args) == 1 && IsError(args[0]) {
			return args[0]
		}
		return ApplyFunction(function, args)
	case *ast.ArrayLiteral:
		elements := EvalExpressions(node.Elements, env)
		if len(elements) == 1 && IsError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if IsError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if IsError(index) {
			return index
		}
		return EvalIndexExpression(left, index)
	}
	return nil
}

func EvalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJECT && index.Type() == object.INTEGER_OBJECT:
		return EvalArrayIndexExpression(left, index)
	default:
		return NewError("index operator not supported: %s", left.Type())
	}
}

func EvalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return NULL
	}
	return arrayObject.Elements[idx]
}

func ApplyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := ExtendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return UnwrapReturnValue(evaluated)
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return NewError("not a function: %s", fn.Type())
	}
}

func ExtendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

func UnwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func EvalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if IsError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func EvalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return NewError("identifier not found: " + node.Value)
}

func EvalIfExpression(ifExpression *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ifExpression.Condition, env)
	if IsError(condition) {
		return condition
	}

	if IsTruthy(condition) {
		return Eval(ifExpression.Consequence, env)
	} else if ifExpression.Alternative != nil {
		return Eval(ifExpression.Alternative, env)
	} else {
		return NULL
	}
}

func IsTruthy(object object.Object) bool {
	switch object {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func NewError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func EvalProgram(statements []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func IsError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJECT
	}
	return false
}

func EvalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			if result.Type() == object.RETURN_VALUE_OBJECT || result.Type() == object.ERROR_OBJECT {
				return result
			}
		}
	}

	return result
}

func EvalInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJECT && right.Type() == object.INTEGER_OBJECT:
		return EvalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJECT && right.Type() == object.STRING_OBJECT:
		return EvalStringInfixExpression(operator, left, right)
	case operator == "==":
		return NativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return NativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return NewError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return NewError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func EvalIntegerInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value
	switch operator {
	case "+":
		return &object.Integer{Value: leftValue + rightValue}
	case "-":
		return &object.Integer{Value: leftValue - rightValue}
	case "*":
		return &object.Integer{Value: leftValue * rightValue}
	case "/":
		return &object.Integer{Value: leftValue / rightValue}
	case "<":
		return NativeBoolToBooleanObject(leftValue < rightValue)
	case ">":
		return NativeBoolToBooleanObject(leftValue > rightValue)
	case "==":
		return NativeBoolToBooleanObject(leftValue == rightValue)
	case "!=":
		return NativeBoolToBooleanObject(leftValue != rightValue)
	default:
		return NewError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func EvalStringInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value
	if operator != "+" {
		return NewError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
	return &object.String{Value: leftValue + rightValue}
}

func NativeBoolToBooleanObject(input bool) object.Object {
	if input {
		return TRUE
	}
	return FALSE
}

func EvalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return EvalBangOperatorExpression(right)
	case "-":
		return EvalPrefixMinusOperatorExpression(right)

	default:
		return NewError("unknown operator: %s%s", operator, right.Type())
	}
}

func EvalBangOperatorExpression(right object.Object) object.Object {
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

func EvalPrefixMinusOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJECT {
		return NewError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}
