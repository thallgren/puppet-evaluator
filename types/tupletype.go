package types

import (
	"io"
	"math"

	"github.com/lyraproj/puppet-evaluator/eval"
)

type TupleType struct {
	size              *IntegerType
	givenOrActualSize *IntegerType
	types             []eval.Type
}

var Tuple_Type eval.ObjectType

func init() {
	Tuple_Type = newObjectType(`Pcore::TupleType`,
		`Pcore::AnyType {
	attributes => {
		types => Array[Type],
		size_type => {
      type => Optional[Type[Integer]],
      value => undef
    }
  }
}`, func(ctx eval.Context, args []eval.Value) eval.Value {
			tupleArgs := args[0].(*ArrayValue).AppendTo([]eval.Value{})
			if len(args) > 1 {
				tupleArgs = append(tupleArgs, args[1].(*IntegerType).Parameters()...)
			}
			return NewTupleType2(tupleArgs...)
		})

	// Go constructor for Tuple instances is registered by ArrayType
}

func DefaultTupleType() *TupleType {
	return tupleType_DEFAULT
}

func EmptyTupleType() *TupleType {
	return tupleType_EMPTY
}

func NewTupleType(types []eval.Type, size *IntegerType) *TupleType {
	var givenOrActualSize *IntegerType
	sz := int64(len(types))
	if size == nil {
		givenOrActualSize = NewIntegerType(sz, sz)
	} else {
		if sz == 0 {
			if *size == *IntegerType_POSITIVE {
				return DefaultTupleType()
			}
			if *size == *IntegerType_ZERO {
				return EmptyTupleType()
			}
		}
		givenOrActualSize = size
	}
	return &TupleType{size, givenOrActualSize, types}
}

func NewTupleType2(args ...eval.Value) *TupleType {
	return tupleFromArgs(false, WrapValues(args))
}

func NewTupleType3(args eval.List) *TupleType {
	return tupleFromArgs(false, args)
}

func tupleFromArgs(callable bool, args eval.List) *TupleType {
	argc := args.Len()
	if argc == 0 {
		return tupleType_DEFAULT
	}

	var rng, givenOrActualRng *IntegerType
	var ok bool
	var min int64

	last := args.At(argc - 1)
	max := int64(-1)
	if _, ok = last.(*DefaultValue); ok {
		max = math.MaxInt64
	} else if n, ok := toInt(last); ok {
		max = n
	}
	if max >= 0 {
		if argc == 1 {
			rng = NewIntegerType(min, math.MaxInt64)
			argc = 0
		} else {
			if min, ok = toInt(args.At(argc - 2)); ok {
				rng = NewIntegerType(min, max)
				argc -= 2
			} else {
				argc--
				rng = NewIntegerType(max, int64(argc))
			}
		}
		givenOrActualRng = rng
	} else {
		rng = nil
		givenOrActualRng = NewIntegerType(int64(argc), int64(argc))
	}

	if argc == 0 {
		if *rng == *IntegerType_ZERO {
			return tupleType_EMPTY
		}
		if callable {
			return &TupleType{rng, rng, []eval.Type{DefaultUnitType()}}
		}
		if *rng == *IntegerType_POSITIVE {
			return tupleType_DEFAULT
		}
		return &TupleType{rng, rng, []eval.Type{}}
	}

	var tupleTypes []eval.Type
	ok = false
	failIdx := -1
	if argc == 1 {
		// One arg can be either array of types or a type
		tupleTypes, failIdx = toTypes(args.Slice(0, 1))
		ok = failIdx < 0
	}

	if !ok {
		tupleTypes, failIdx = toTypes(args.Slice(0, argc))
		if failIdx >= 0 {
			name := `Tuple[]`
			if callable {
				name = `Callable[]`
			}
			panic(NewIllegalArgumentType2(name, failIdx, `Type`, args.At(failIdx)))
		}
	}
	return &TupleType{rng, givenOrActualRng, tupleTypes}
}

func (t *TupleType) Accept(v eval.Visitor, g eval.Guard) {
	v(t)
	t.size.Accept(v, g)
	for _, c := range t.types {
		c.Accept(v, g)
	}
}

func (t *TupleType) At(i int) eval.Value {
	if i >= 0 {
		if i < len(t.types) {
			return t.types[i]
		}
		if int64(i) < t.givenOrActualSize.max {
			return t.types[len(t.types)-1]
		}
	}
	return _UNDEF
}

func (t *TupleType) CommonElementType() eval.Type {
	top := len(t.types)
	if top == 0 {
		return anyType_DEFAULT
	}
	cet := t.types[0]
	for idx := 1; idx < top; idx++ {
		cet = commonType(cet, t.types[idx])
	}
	return cet
}

func (t *TupleType) Default() eval.Type {
	return tupleType_DEFAULT
}

func (t *TupleType) Equals(o interface{}, g eval.Guard) bool {
	if ot, ok := o.(*TupleType); ok && len(t.types) == len(ot.types) && eval.GuardedEquals(t.size, ot.size, g) {
		for idx, col := range t.types {
			if !col.Equals(ot.types[idx], g) {
				return false
			}
		}
		return true
	}
	return false
}

func (t *TupleType) Generic() eval.Type {
	return NewTupleType(alterTypes(t.types, generalize), t.size)
}

func (t *TupleType) Get(key string) (value eval.Value, ok bool) {
	switch key {
	case `types`:
		tps := make([]eval.Value, len(t.types))
		for i, t := range t.types {
			tps[i] = t
		}
		return WrapValues(tps), true
	case `size_type`:
		if t.size == nil {
			return _UNDEF, true
		}
		return t.size, true
	}
	return nil, false
}

func (t *TupleType) IsAssignable(o eval.Type, g eval.Guard) bool {
	switch o.(type) {
	case *ArrayType:
		at := o.(*ArrayType)
		if !GuardedIsInstance(t.givenOrActualSize, WrapInteger(at.size.Min()), g) {
			return false
		}
		top := len(t.types)
		if top == 0 {
			return true
		}
		elemType := at.typ
		for idx := 0; idx < top; idx++ {
			if !GuardedIsAssignable(t.types[idx], elemType, g) {
				return false
			}
		}
		return true

	case *TupleType:
		tt := o.(*TupleType)
		if !(t.size == nil || GuardedIsInstance(t.size, WrapInteger(tt.givenOrActualSize.Min()), g)) {
			return false
		}

		if len(t.types) > 0 {
			top := len(tt.types)
			if top == 0 {
				return t.givenOrActualSize.min == 0
			}

			last := len(t.types) - 1
			for idx := 0; idx < top; idx++ {
				myIdx := idx
				if myIdx > last {
					myIdx = last
				}
				if !GuardedIsAssignable(t.types[myIdx], tt.types[idx], g) {
					return false
				}
			}
		}
		return true

	default:
		return false
	}
}

func (t *TupleType) IsInstance(v eval.Value, g eval.Guard) bool {
	if iv, ok := v.(*ArrayValue); ok {
		return t.IsInstance2(iv, g)
	}
	return false
}

func (t *TupleType) IsInstance2(vs eval.List, g eval.Guard) bool {
	osz := vs.Len()
	if !t.givenOrActualSize.IsInstance3(osz) {
		return false
	}

	last := len(t.types) - 1
	if last < 0 {
		return true
	}

	tdx := 0
	for idx := 0; idx < osz; idx++ {
		if !GuardedIsInstance(t.types[tdx], vs.At(idx), g) {
			return false
		}
		if tdx < last {
			tdx++
		}
	}
	return true
}

func (t *TupleType) IsInstance3(vs []eval.Value, g eval.Guard) bool {
	osz := len(vs)
	if !t.givenOrActualSize.IsInstance3(osz) {
		return false
	}

	last := len(t.types) - 1
	if last < 0 {
		return true
	}

	tdx := 0
	for idx := 0; idx < osz; idx++ {
		if !GuardedIsInstance(t.types[tdx], vs[idx], g) {
			return false
		}
		if tdx < last {
			tdx++
		}
	}
	return true
}

func (t *TupleType) MetaType() eval.ObjectType {
	return Tuple_Type
}

func (t *TupleType) Name() string {
	return `Tuple`
}

func (t *TupleType) Resolve(c eval.Context) eval.Type {
	rts := make([]eval.Type, len(t.types))
	for i, ts := range t.types {
		rts[i] = resolve(c, ts)
	}
	t.types = rts
	return t
}

func (t *TupleType) CanSerializeAsString() bool {
	for _, v := range t.types {
		if !canSerializeAsString(v) {
			return false
		}
	}
	return true
}

func (t *TupleType) SerializationString() string {
	return t.String()
}

func (t *TupleType) Size() *IntegerType {
	return t.givenOrActualSize
}

func (t *TupleType) String() string {
	return eval.ToString2(t, NONE)
}

func (t *TupleType) Parameters() []eval.Value {
	top := len(t.types)
	params := make([]eval.Value, 0, top+2)
	for _, c := range t.types {
		params = append(params, c)
	}
	if !(t.size == nil || top == 0 && *t.size == *IntegerType_POSITIVE) {
		params = append(params, t.size.SizeParameters()...)
	}
	return params
}

func (t *TupleType) ToString(b io.Writer, s eval.FormatContext, g eval.RDetect) {
	TypeToString(t, b, s, g)
}

func (t *TupleType) PType() eval.Type {
	return &TypeType{t}
}

func (t *TupleType) Types() []eval.Type {
	return t.types
}

var tupleType_DEFAULT = &TupleType{IntegerType_POSITIVE, IntegerType_POSITIVE, []eval.Type{}}
var tupleType_EMPTY = &TupleType{IntegerType_ZERO, IntegerType_ZERO, []eval.Type{}}
