package impl

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/puppet-parser/parser"
)

func evalAccessExpression(e eval.Evaluator, expr *parser.AccessExpression) (result eval.Value) {
	keys := expr.Keys()
	op := expr.Operand()
	if qr, ok := op.(*parser.QualifiedReference); ok {
		if (qr.Name() == `TypeSet` || qr.Name() == `Object`) && len(keys) == 1 {
			// Defer evaluation of the type parameter until type is resolved
			if hash, ok := keys[0].(*parser.LiteralHash); ok {
				name := ``
				ne := hash.Get(`name`)
				if ne != nil {
					name = e.Eval(ne).String()
				}
				if qr.Name() == `Object` {
					return types.NewObjectType(name, nil, hash)
				}

				na := eval.RUNTIME_NAME_AUTHORITY
				ne = hash.Get(`name_authority`)
				if ne != nil {
					na = eval.URI(e.Eval(ne).String())
				}
				return types.NewTypeSetType(na, name, hash)
			}
		}

		args := make([]eval.Value, len(keys))
		e.DoStatic(func() {
			for idx, key := range keys {
				args[idx] = e.Eval(key)
			}
		})
		return eval_ParameterizedTypeExpression(e, qr, args, expr)
	}

	args := make([]eval.Value, len(keys))
	for idx, key := range keys {
		args[idx] = e.Eval(key)
	}

	lhs := e.Eval(op)

	switch lhs.(type) {
	case eval.List:
		return accessIndexedValue(expr, lhs.(eval.List), args)
	default:
		if tem, ok := lhs.PType().(eval.TypeWithCallableMembers); ok {
			if mbr, ok := tem.Member(`[]`); ok {
				return mbr.Call(e, lhs, nil, args)
			}
		}
	}
	panic(evalError(eval.EVAL_OPERATOR_NOT_APPLICABLE, op, issue.H{`operator`: `[]`, `left`: lhs.PType()}))
}

func accessIndexedValue(expr *parser.AccessExpression, lhs eval.List, args []eval.Value) (result eval.Value) {
	nArgs := len(args)

	intArg := func(index int) int {
		key := args[index]
		if arg, ok := eval.ToInt(key); ok {
			return int(arg)
		}
		panic(evalError(eval.EVAL_ILLEGAL_ARGUMENT_TYPE, expr.Keys()[index],
			issue.H{`expression`: lhs.PType(), `number`: 0, `expected`: `Integer`, `actual`: key}))
	}

	indexArg := func(argIndex int) int {
		index := intArg(argIndex)
		if index < 0 {
			index = lhs.Len() + index
		}
		if index > lhs.Len() {
			index = lhs.Len()
		}
		return index
	}

	countArg := func(argIndex int, start int) (count int) {
		count = intArg(argIndex)
		if start < 0 {
			if count > 0 {
				count += start
				if count < 0 {
					count = 0
				}
			}
			start = 0
		}
		if count < 0 {
			count = 1 + (lhs.Len() + count) - start
			if count < 0 {
				count = 0
			}
		} else if start+count > lhs.Len() {
			count = lhs.Len() - start
		}
		return
	}

	if hv, ok := lhs.(*types.HashValue); ok {
		if hv.Len() == 0 {
			return eval.UNDEF
		}
		if nArgs == 0 {
			panic(evalError(eval.EVAL_ILLEGAL_ARGUMENT_COUNT, expr, issue.H{`expression`: lhs.PType(), `expected`: `at least one`, `actual`: nArgs}))
		}
		if nArgs == 1 {
			if v, ok := hv.Get(args[0]); ok {
				return v
			}
			return eval.UNDEF
		}
		el := make([]eval.Value, 0, nArgs)
		for _, key := range args {
			if v, ok := hv.Get(key); ok {
				el = append(el, v)
			}
		}
		return types.WrapValues(el)
	}

	if nArgs == 0 || nArgs > 2 {
		panic(evalError(eval.EVAL_ILLEGAL_ARGUMENT_COUNT, expr, issue.H{`expression`: lhs.PType(), `expected`: `1 or 2`, `actual`: nArgs}))
	}
	if nArgs == 2 {
		start := indexArg(0)
		count := countArg(1, start)
		if start < 0 {
			start = 0
		}
		if start == lhs.Len() || count == 0 {
			if _, ok := lhs.(*types.StringValue); ok {
				return eval.EMPTY_STRING
			}
			return eval.EMPTY_ARRAY
		}
		return lhs.Slice(start, start+count)
	}
	pos := intArg(0)
	if pos < 0 {
		pos = lhs.Len() + pos
		if pos < 0 {
			return eval.UNDEF
		}
	}
	if pos >= lhs.Len() {
		return eval.UNDEF
	}
	return lhs.At(pos)
}

func eval_ParameterizedTypeExpression(e eval.Evaluator, qr *parser.QualifiedReference, args []eval.Value, expr *parser.AccessExpression) (tp eval.Type) {
	dcName := qr.DowncasedName()
	defer func() {
		if err := recover(); err != nil {
			convertCallError(err, expr, expr.Keys())
		}
	}()

	switch dcName {
	case `array`:
		tp = types.NewArrayType2(args...)
	case `boolean`:
		tp = types.NewBooleanType2(args...)
	case `callable`:
		tp = types.NewCallableType2(args...)
	case `collection`:
		tp = types.NewCollectionType2(args...)
	case `enum`:
		tp = types.NewEnumType2(args...)
	case `float`:
		tp = types.NewFloatType2(args...)
	case `hash`:
		tp = types.NewHashType2(args...)
	case `init`:
		tp = types.NewInitType2(args...)
	case `integer`:
		tp = types.NewIntegerType2(args...)
	case `iterable`:
		tp = types.NewIterableType2(args...)
	case `iterator`:
		tp = types.NewIteratorType2(args...)
	case `like`:
		tp = types.NewLikeType2(args...)
	case `notundef`:
		tp = types.NewNotUndefType2(args...)
	case `object`:
		tp = types.NewObjectType2(e, args...)
	case `optional`:
		tp = types.NewOptionalType2(args...)
	case `pattern`:
		tp = types.NewPatternType2(args...)
	case `regexp`:
		tp = types.NewRegexpType2(args...)
	case `runtime`:
		tp = types.NewRuntimeType2(args...)
	case `semver`:
		tp = types.NewSemVerType2(args...)
	case `sensitive`:
		tp = types.NewSensitiveType2(args...)
	case `string`:
		tp = types.NewStringType2(args...)
	case `struct`:
		tp = types.NewStructType2(args...)
	case `timespan`:
		tp = types.NewTimespanType2(args...)
	case `timestamp`:
		tp = types.NewTimestampType2(args...)
	case `tuple`:
		tp = types.NewTupleType2(args...)
	case `type`:
		tp = types.NewTypeType2(args...)
	case `typereference`:
		tp = types.NewTypeReferenceType2(args...)
	case `uri`:
		tp = types.NewUriType2(args...)
	case `variant`:
		tp = types.NewVariantType2(args...)
	case `any`:
	case `binary`:
	case `catalogentry`:
	case `data`:
	case `default`:
	case `numeric`:
	case `scalar`:
	case `semverrange`:
	case `typealias`:
	case `undef`:
	case `unit`:
		panic(evalError(eval.EVAL_NOT_PARAMETERIZED_TYPE, expr, issue.H{`type`: expr}))
	default:
		oe := e.Eval(qr)
		if oo, ok := oe.(eval.ObjectType); ok && oo.IsParameterized() {
			tp = types.NewObjectTypeExtension(e, oo, args)
		} else {
			tp = types.NewTypeReferenceType(expr.String())
		}
	}
	return
}
