package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		e := fmt.Errorf("%s: invalid number of arguments", os.Args[0])
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	s := os.Args[1]
	inpt = s
	t = tokenize([]rune(s))
	p := program()
	for fn := p; fn != nil; fn = fn.next {
		o := 0
		for v := fn.locals; v != nil; v = v.next {
			o += 8
			v.offset = o
		}
		fn.stackSize = o
	}
	codegen(p)
	os.Exit(0)
}
