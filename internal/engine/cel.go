package engine

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
)

func newCELEnv() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Variable("event", cel.DynType),
		cel.Variable("request", cel.DynType),
		cel.Variable("meta", cel.DynType),
		cel.Variable("protocol", cel.DynType),
		cel.Variable("namespace", cel.DynType),
	)
}

func compileCELCondition(expression string) (cel.Program, error) {
	env, err := newCELEnv()
	if err != nil {
		return nil, err
	}

	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}
	if ast.OutputType() != cel.BoolType && ast.OutputType() != cel.DynType {
		return nil, fmt.Errorf("condition expr must return bool, got %s", ast.OutputType())
	}
	return env.Program(ast)
}

func compileCELBody(expression string) (cel.Program, error) {
	env, err := newCELEnv()
	if err != nil {
		return nil, err
	}

	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}
	return env.Program(ast)
}

func evalCELBool(program cel.Program, input map[string]any) (bool, error) {
	if program == nil {
		return false, fmt.Errorf("cel program is nil")
	}
	out, _, err := program.Eval(input)
	if err != nil {
		return false, err
	}
	switch typed := out.(type) {
	case types.Bool:
		return bool(typed), nil
	default:
		value := out.Value()
		boolean, ok := value.(bool)
		if !ok {
			return false, fmt.Errorf("cel expression did not return bool")
		}
		return boolean, nil
	}
}

func evalCELValue(program cel.Program, input map[string]any) (any, error) {
	if program == nil {
		return nil, fmt.Errorf("cel program is nil")
	}
	out, _, err := program.Eval(input)
	if err != nil {
		return nil, err
	}
	return nativeCELValue(out)
}

func nativeCELValue(value ref.Val) (any, error) {
	if value == nil {
		return nil, nil
	}
	switch typed := value.(type) {
	case types.Bool:
		return bool(typed), nil
	case types.Int:
		return int64(typed), nil
	case types.Uint:
		return uint64(typed), nil
	case types.Double:
		return float64(typed), nil
	case types.String:
		return string(typed), nil
	case types.Bytes:
		return []byte(typed), nil
	case traits.Lister:
		size := typed.Size()
		result := make([]any, 0, int(size.(types.Int)))
		it := typed.Iterator()
		for it.HasNext() == types.True {
			item, err := nativeCELValue(it.Next())
			if err != nil {
				return nil, err
			}
			result = append(result, item)
		}
		return result, nil
	case traits.Mapper:
		result := make(map[string]any)
		it := typed.Iterator()
		for it.HasNext() == types.True {
			keyVal := it.Next()
			key, err := nativeCELValue(keyVal)
			if err != nil {
				return nil, err
			}
			mappedValue := typed.Get(keyVal)
			valueNative, err := nativeCELValue(mappedValue)
			if err != nil {
				return nil, err
			}
			result[fmt.Sprint(key)] = valueNative
		}
		return result, nil
	}
	return value.Value(), nil
}
