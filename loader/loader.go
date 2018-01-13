package loader

import (
	"fmt"
	"sync"

	. "github.com/puppetlabs/go-evaluator/evaluator"
)

type (
	loaderEntry struct {
		value interface{}
		origin string
	}

	basicLoader struct {
		namedEntries map[string]Entry
	}

	parentedLoader struct {
		basicLoader
		parent Loader
	}
)

var staticLoader = &basicLoader{make(map[string]Entry, 64)}
var resolvableConstructors = make([]ResolvableFunction, 0, 16)
var resolvableFunctions = make([]ResolvableFunction, 0, 16)
var resolvableFunctionsLock sync.Mutex

func init() {
	StaticLoader = func() Loader {
		return staticLoader
	}

	NewParentedLoader = func(parent Loader) DefiningLoader {
		return &parentedLoader{basicLoader{make(map[string]Entry, 64)}, parent}
	}

	RegisterGoFunction = func(function ResolvableFunction) {
		resolvableFunctionsLock.Lock()
		resolvableFunctions = append(resolvableFunctions, function)
		resolvableFunctionsLock.Unlock()
	}

  RegisterGoConstructor = func(function ResolvableFunction) {
		resolvableFunctionsLock.Lock()
		resolvableConstructors = append(resolvableConstructors, function)
		resolvableFunctionsLock.Unlock()
	}

	NewLoaderEntry = func (value interface{}, origin string) Entry {
		return &loaderEntry{value, origin}
	}

	Load = load
}

func popDeclaredGoFunctions() (funcs []ResolvableFunction, ctors []ResolvableFunction) {
	resolvableFunctionsLock.Lock()
	funcs = resolvableFunctions
	if len(funcs) > 0 {
		resolvableFunctions = make([]ResolvableFunction, 0, 16)
	}
	ctors = resolvableConstructors
	if len(ctors) > 0 {
		resolvableConstructors = make([]ResolvableFunction, 0, 16)
	}
	resolvableFunctionsLock.Unlock()
	return
}

func (e *loaderEntry) Origin() string {
	return e.origin
}

func (e *loaderEntry) Value() interface{} {
	return e.value
}

func (l *basicLoader) ResolveGoFunctions(c EvalContext) {
	funcs, ctors := popDeclaredGoFunctions()
	for _, rf := range funcs {
		l.SetEntry(NewTypedName(FUNCTION, rf.Name()), &loaderEntry{rf.Resolve(c), ``})
	}
	for _, ct := range ctors {
		l.SetEntry(NewTypedName(CONSTRUCTOR, ct.Name()), &loaderEntry{ct.Resolve(c) ,``})
	}
}

func load(l Loader, name TypedName) (interface{}, bool) {
	if name.NameAuthority() != l.NameAuthority() {
		return nil, false
	}
	entry := l.LoadEntry(name)
	if entry == nil {
		if dl, ok := l.(DefiningLoader); ok {
			dl.SetEntry(name, &loaderEntry{nil, ``})
		}
		return nil, false
	}
	if entry.Value() == nil {
		return nil, false
	}
	return entry.Value(), true
}

func (l *basicLoader) LoadEntry(name TypedName) Entry {
	return l.GetEntry(name)
}

func (l *basicLoader) GetEntry(name TypedName) Entry {
	return l.namedEntries[name.MapKey()]
}

func (l *basicLoader) SetEntry(name TypedName, entry Entry) Entry {
	if _, ok := l.namedEntries[name.MapKey()]; ok {
		panic(fmt.Sprintf(`Attempt to redefine %s`, name.String()))
	}
	l.namedEntries[name.MapKey()] = entry
	return entry
}

func (l *basicLoader) NameAuthority() URI {
	return RUNTIME_NAME_AUTHORITY
}

func (l *parentedLoader) LoadEntry(name TypedName) Entry {
	entry := l.parent.LoadEntry(name)
	if entry == nil || entry.Value() == nil {
		entry = l.basicLoader.LoadEntry(name)
	}
	return entry
}

func (l *parentedLoader) NameAuthority() URI {
	return l.parent.NameAuthority()
}
