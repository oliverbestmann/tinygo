package main

import "iter"

func main() {
	testFuncRange(counter)
	testIterPull(counter)
	testStoredIter()
	println("go1.23 has lift-off!")
}

func testFuncRange(it iter.Seq[int]) {
	for i := range it {
		println(i)
	}
}

func testIterPull(it iter.Seq[int]) {
	next, stop := iter.Pull(it)
	defer stop()
	for {
		i, ok := next()
		if !ok {
			break
		}
		println(i)
	}
}

func counter(yield func(int) bool) {
	for i := 10; i >= 1; i-- {
		yield(i)
	}
}

type query struct{ words iter.Seq[string] }

//go:noinline
func newQuery(words ...string) *query {
	return &query{words: func(yield func(string) bool) {
		for _, w := range words {
			if !yield(w) {
				return
			}
		}
	}}
}

// The iterator is only known at runtime, so its call can only be devirtualized
// behind a pointer comparison.
func testStoredIter() {
	q := newQuery("a", "b", "stop", "c")
	for w := range q.words {
		if w == "stop" {
			break
		}
		println(w)
	}
}
