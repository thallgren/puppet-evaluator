package types

import (
	"io"
	"strings"

	"github.com/lyraproj/puppet-evaluator/errors"
	"github.com/lyraproj/puppet-evaluator/eval"
	"reflect"
)

var Boolean_FALSE = &BooleanValue{0}
var Boolean_TRUE = &BooleanValue{1}

type (
	BooleanType struct {
		value int // -1 == unset, 0 == false, 1 == true
	}

	// BooleanValue keeps only the value because the type is known and not parameterized
	BooleanValue BooleanType
)

var booleanType_DEFAULT = &BooleanType{-1}

var Boolean_Type eval.ObjectType

func init() {
	Boolean_Type = newObjectType(`Pcore::BooleanType`, `Pcore::ScalarDataType {
  attributes => {
    value => { type => Optional[Boolean], value => undef }
  }
}`, func(ctx eval.Context, args []eval.Value) eval.Value {
		return NewBooleanType2(args...)
	})

	newGoConstructor(`Boolean`,
		func(d eval.Dispatch) {
			d.Param(`Variant[Integer, Float, Boolean, Enum['false','true','yes','no','y','n',true]]`)
			d.Function(func(c eval.Context, args []eval.Value) eval.Value {
				arg := args[0]
				switch arg.(type) {
				case *IntegerValue:
					if arg.(*IntegerValue).Int() == 0 {
						return Boolean_FALSE
					}
					return Boolean_TRUE
				case *FloatValue:
					if arg.(*FloatValue).Float() == 0.0 {
						return Boolean_FALSE
					}
					return Boolean_TRUE
				case *BooleanValue:
					return arg
				default:
					switch strings.ToLower(arg.String()) {
					case `false`, `no`, `n`:
						return Boolean_FALSE
					default:
						return Boolean_TRUE
					}
					return arg.(eval.IterableValue).Iterator().AsArray()
				}
			})
		},
	)
}

func DefaultBooleanType() *BooleanType {
	return booleanType_DEFAULT
}

func NewBooleanType(value bool) *BooleanType {
	n := 0
	if value {
		n = 1
	}
	return &BooleanType{n}
}

func NewBooleanType2(args ...eval.Value) *BooleanType {
	switch len(args) {
	case 0:
		return DefaultBooleanType()
	case 1:
		if bv, ok := args[0].(*BooleanValue); ok {
			return NewBooleanType(bv.Bool())
		}
		panic(NewIllegalArgumentType2(`Boolean[]`, 0, `Boolean`, args[0]))
	default:
		panic(errors.NewIllegalArgumentCount(`Boolean[]`, `0 or 1`, len(args)))
	}
}

func (t *BooleanType) Accept(v eval.Visitor, g eval.Guard) {
	v(t)
}

func (t *BooleanType) Default() eval.Type {
	return booleanType_DEFAULT
}

func (t *BooleanType) Generic() eval.Type {
	return booleanType_DEFAULT
}

func (t *BooleanType) Equals(o interface{}, g eval.Guard) bool {
	if bo, ok := o.(*BooleanType); ok {
		return t.value == bo.value
	}
	return false
}

func (t *BooleanType) Get(key string) (eval.Value, bool) {
	switch key {
	case `value`:
		switch t.value {
		case 0:
			return Boolean_FALSE, true
		case 1:
			return Boolean_TRUE, true
		default:
			return eval.UNDEF, true
		}
	default:
		return nil, false
	}
}

func (t *BooleanType) MetaType() eval.ObjectType {
	return Boolean_Type
}

func (t *BooleanType) Name() string {
	return `Boolean`
}

func (t *BooleanType) String() string {
	switch t.value {
	case 0:
		return `Boolean[false]`
	case 1:
		return `Boolean[true]`
	default:
		return `Boolean`
	}
}

func (t *BooleanType) IsAssignable(o eval.Type, g eval.Guard) bool {
	if bo, ok := o.(*BooleanType); ok {
		return t.value == -1 || t.value == bo.value
	}
	return false
}

func (t *BooleanType) IsInstance(o eval.Value, g eval.Guard) bool {
	if bo, ok := o.(*BooleanValue); ok {
		return t.value == -1 || t.value == bo.value
	}
	return false
}

func (t *BooleanType) Parameters() []eval.Value {
	if t.value == -1 {
		return eval.EMPTY_VALUES
	}
	return []eval.Value{&BooleanValue{t.value}}
}

func (t *BooleanType) ReflectType(c eval.Context) (reflect.Type, bool) {
	return reflect.TypeOf(true), true
}

func (t *BooleanType) CanSerializeAsString() bool {
	return true
}

func (t *BooleanType) SerializationString() string {
	return t.String()
}

func (t *BooleanType) ToString(b io.Writer, s eval.FormatContext, g eval.RDetect) {
	TypeToString(t, b, s, g)
}

func (t *BooleanType) PType() eval.Type {
	return &TypeType{t}
}

func WrapBoolean(val bool) *BooleanValue {
	if val {
		return Boolean_TRUE
	}
	return Boolean_FALSE
}

func (bv *BooleanValue) Bool() bool {
	return bv.value == 1
}

func (bv *BooleanValue) Equals(o interface{}, g eval.Guard) bool {
	if ov, ok := o.(*BooleanValue); ok {
		return bv.value == ov.value
	}
	return false
}

func (bv *BooleanValue) Float() float64 {
	return float64(bv.value)
}

func (bv *BooleanValue) Int() int64 {
	return int64(bv.value)
}

func (bv *BooleanValue) Reflect(c eval.Context) reflect.Value {
	return reflect.ValueOf(bv.value == 1)
}

var theTrue = true
var theFalse = false
var theTruePtr = &theTrue
var theFalsePtr = &theFalse

var reflectTrue = reflect.ValueOf(theTrue)
var reflectFalse = reflect.ValueOf(theFalse)
var reflectTruePtr = reflect.ValueOf(theTruePtr)
var reflectFalsePtr = reflect.ValueOf(theFalsePtr)

func (bv *BooleanValue) ReflectTo(c eval.Context, value reflect.Value) {
	if value.Kind() == reflect.Interface {
		if bv.value == 1 {
			value.Set(reflectTrue)
		} else {
			value.Set(reflectFalse)
		}
	} else if value.Kind() == reflect.Ptr {
		if bv.value == 1 {
			value.Set(reflectTruePtr)
		} else {
			value.Set(reflectFalsePtr)
		}
	} else {
		value.SetBool(bv.value == 1)
	}
}

func (t *BooleanValue) CanSerializeAsString() bool {
	return true
}

func (t *BooleanValue) SerializationString() string {
	return t.String()
}

func (bv *BooleanValue) String() string {
	if bv.value == 1 {
		return `true`
	}
	return `false`
}

func (bv *BooleanValue) ToString(b io.Writer, s eval.FormatContext, g eval.RDetect) {
	f := eval.GetFormat(s.FormatMap(), bv.PType())
	switch f.FormatChar() {
	case 't':
		f.ApplyStringFlags(b, bv.stringVal(f.IsAlt(), `true`, `false`), false)
	case 'T':
		f.ApplyStringFlags(b, bv.stringVal(f.IsAlt(), `True`, `False`), false)
	case 'y':
		f.ApplyStringFlags(b, bv.stringVal(f.IsAlt(), `yes`, `no`), false)
	case 'Y':
		f.ApplyStringFlags(b, bv.stringVal(f.IsAlt(), `Yes`, `No`), false)
	case 'd', 'x', 'X', 'o', 'b', 'B':
		WrapInteger(bv.Int()).ToString(b, eval.NewFormatContext(DefaultIntegerType(), f, s.Indentation()), g)
	case 'e', 'E', 'f', 'g', 'G', 'a', 'A':
		WrapFloat(bv.Float()).ToString(b, eval.NewFormatContext(DefaultFloatType(), f, s.Indentation()), g)
	case 's', 'p':
		f.ApplyStringFlags(b, bv.stringVal(false, `true`, `false`), false)
	default:
		panic(s.UnsupportedFormat(bv.PType(), `tTyYdxXobBeEfgGaAsp`, f))
	}
}

func (bv *BooleanValue) stringVal(alt bool, yes string, no string) string {
	str := no
	if bv.value == 1 {
		str = yes
	}
	if alt {
		str = str[:1]
	}
	return str
}

func (bv *BooleanValue) ToKey() eval.HashKey {
	if bv.value == 1 {
		return eval.HashKey([]byte{1, HK_BOOLEAN, 1})
	}
	return eval.HashKey([]byte{1, HK_BOOLEAN, 0})
}

func (bv *BooleanValue) PType() eval.Type {
	return DefaultBooleanType()
}
