package lib

type Checker = func(any) bool

func GenericWithLocals[T any]() (any, any, Checker, Checker) {
	type UsesT struct{ V T }
	type NoT struct{ V int }
	var z T
	return UsesT{V: z}, NoT{V: 42},
		func(x any) bool { _, ok := x.(UsesT); return ok },
		func(x any) bool { _, ok := x.(NoT); return ok }
}

func SiblingClosures() (any, any, Checker, Checker) {
	var av, bv any
	var ac, bc Checker
	func() {
		type Foo struct{ V int }
		av = Foo{V: 1}
		ac = func(x any) bool { _, ok := x.(Foo); return ok }
	}()
	func() {
		type Foo struct{ V int }
		bv = Foo{V: 2}
		bc = func(x any) bool { _, ok := x.(Foo); return ok }
	}()
	return av, bv, ac, bc
}

type Box[T any] struct{}

func Wrap[T any](v T) any { return v }

// Local declares a local type in a method of a generic type. It is passed to
// another generic function, so the local type is part of a function name.
func (Box[T]) Local() (any, Checker) {
	type Inner struct{ V T }
	return Wrap[Inner](Inner{}), func(x any) bool { _, ok := x.(Inner); return ok }
}
