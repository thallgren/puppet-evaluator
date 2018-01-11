package types

import (
	"io"

	. "github.com/puppetlabs/go-evaluator/evaluator"
	. "github.com/puppetlabs/go-evaluator/semver"
	. "github.com/puppetlabs/go-parser/parser"
)

type (
	TypeSetReference struct {
		name          string
		nameAuthority URI
		versionRange  *VersionRange
		typeSet       *TypeSetType
		annotations   *HashValue
	}

	TypeSetType struct {
		dcToCcMap          map[string]string
		name               string
		nameAuthority      URI
		pcoreURI           URI
		pcoreVersion       *Version
		version            *Version
		loader             Loader
		annotations        *HashValue
		initHashExpression Expression
	}
)

var typeSetType_DEFAULT = &TypeSetType{
	name:          `DefaultTypeSet`,
	nameAuthority: RUNTIME_NAME_AUTHORITY,
	pcoreURI:      PCORE_URI,
	pcoreVersion:  PCORE_VERSION,
	version:       NewVersion4(0, 0, 0, ``, ``),
}

func DefaultTypeSetType() *TypeSetType {
	return typeSetType_DEFAULT
}

func (t *TypeSetType) Accept(v Visitor, g Guard) {
	v(t)
	// TODO: Visit typeset members
}

func (t *TypeSetType) Default() PType {
	return typeSetType_DEFAULT
}

func (a *TypeSetType) Equals(other interface{}, guard Guard) bool {
	panic("implement me")
}

func (a *TypeSetType) String() string {
	panic("implement me")
}

func (a *TypeSetType) ToString(bld io.Writer, format FormatContext, g RDetect) {
	panic("implement me")
}

func (a *TypeSetType) Type() PType {
	panic("implement me")
}

func (a *TypeSetType) IsInstance(o PValue, g Guard) bool {
	panic("implement me")
}

func (a *TypeSetType) IsAssignable(t PType, g Guard) bool {
	panic("implement me")
}

func (a *TypeSetType) Annotations() *HashValue {
	return a.annotations
}

func (a *TypeSetType) Name() string {
	return a.name
}