package main

var index = 3

func main() {
	explicitPanic()
	defer func() {
		println("recovered runtime error:", recover() != nil)
	}()
	var s []int
	println(s[index])
}

func explicitPanic() {
	defer func() {
		println("recovered:", recover())
	}()
	panic("explicit")
}
