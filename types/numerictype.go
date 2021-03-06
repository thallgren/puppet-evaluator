package types

import (
	"io"

	"fmt"
	"github.com/lyraproj/puppet-evaluator/errors"
	"github.com/lyraproj/puppet-evaluator/eval"
	"reflect"
	"strconv"
)

type NumericType struct{}

var numericType_DEFAULT = &NumericType{}

var Numeric_Type eval.ObjectType

func init() {
	Numeric_Type = newObjectType(`Pcore::NumericType`, `Pcore::ScalarDataType {}`, func(ctx eval.Context, args []eval.Value) eval.Value {
		return DefaultNumericType()
	})

	newGoConstructor2(`Numeric`,
		func(t eval.LocalTypes) {
			t.Type(`Convertible`, `Variant[Numeric, Boolean, Pattern[/`+FLOAT_PATTERN+`/], Timespan, Timestamp]`)
			t.Type(`NamedArgs`, `Struct[from => Convertible, Optional[abs] => Boolean]`)
		},

		func(d eval.Dispatch) {
			d.Param(`NamedArgs`)
			d.Function(func(c eval.Context, args []eval.Value) eval.Value {
				return numberFromNamedArgs(args, true)
			})
		},

		func(d eval.Dispatch) {
			d.Param(`Convertible`)
			d.OptionalParam(`Boolean`)
			d.Function(func(c eval.Context, args []eval.Value) eval.Value {
				return numberFromPositionalArgs(args, true)
			})
		},
	)
}

func numberFromPositionalArgs(args []eval.Value, tryInt bool) eval.NumericValue {
	n := fromConvertible(args[0], tryInt)
	if len(args) > 1 && args[1].(*BooleanValue).Bool() {
		n = n.Abs()
	}
	return n
}

func numberFromNamedArgs(args []eval.Value, tryInt bool) eval.NumericValue {
	h := args[0].(*HashValue)
	n := fromConvertible(h.Get5(`from`, eval.UNDEF), tryInt)
	a := h.Get5(`abs`, nil)
	if a != nil && a.(*BooleanValue).Bool() {
		n = n.Abs()
	}
	return n
}

func DefaultNumericType() *NumericType {
	return numericType_DEFAULT
}

func (t *NumericType) Accept(v eval.Visitor, g eval.Guard) {
	v(t)
}

func (t *NumericType) Equals(o interface{}, g eval.Guard) bool {
	_, ok := o.(*NumericType)
	return ok
}

func (t *NumericType) IsAssignable(o eval.Type, g eval.Guard) bool {
	switch o.(type) {
	case *IntegerType, *FloatType:
		return true
	default:
		return false
	}
}

func (t *NumericType) IsInstance(o eval.Value, g eval.Guard) bool {
	switch o.PType().(type) {
	case *FloatType, *IntegerType:
		return true
	default:
		return false
	}
}

func (t *NumericType) MetaType() eval.ObjectType {
	return Numeric_Type
}

func (t *NumericType) Name() string {
	return `Numeric`
}

func (t *NumericType) ReflectType(c eval.Context) (reflect.Type, bool) {
	return reflect.TypeOf(float64(0.0)), true
}

func (t *NumericType)  CanSerializeAsString() bool {
  return true
}

func (t *NumericType)  SerializationString() string {
	return t.String()
}


func (t *NumericType) String() string {
	return eval.ToString2(t, NONE)
}

func (t *NumericType) ToString(b io.Writer, s eval.FormatContext, g eval.RDetect) {
	TypeToString(t, b, s, g)
}

func (t *NumericType) PType() eval.Type {
	return &TypeType{t}
}

func fromConvertible(c eval.Value, allowInt bool) eval.NumericValue {
	switch c.(type) {
	case *IntegerValue:
		iv := c.(*IntegerValue)
		if allowInt {
			return iv
		}
		return WrapFloat(iv.Float())
	case *TimestampValue:
		return WrapFloat(c.(*TimestampValue).Float())
	case *TimespanValue:
		return WrapFloat(c.(*TimespanValue).Float())
	case *BooleanValue:
		if allowInt {
			return WrapInteger(c.(*BooleanValue).Int())
		}
		return WrapFloat(c.(*BooleanValue).Float())
	case eval.NumericValue:
		return c.(eval.NumericValue)
	case *StringValue:
		s := c.String()
		if allowInt {
			if i, err := strconv.ParseInt(s, 0, 64); err == nil {
				return WrapInteger(i)
			}
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return WrapFloat(f)
		}
		if allowInt {
			if len(s) > 2 && s[0] == '0' && (s[1] == 'b' || s[1] == 'B') {
				if i, err := strconv.ParseInt(s[2:], 2, 64); err == nil {
					return WrapInteger(i)
				}
			}
		}
	}
	panic(errors.NewArgumentsError(`Numeric`, fmt.Sprintf(`Value of type %s cannot be converted to an Number`, c.PType().String())))
}
