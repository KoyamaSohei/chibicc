package main

import (
	"fmt"
	"os"
)

func alignTo(n int, align int) int {
	return (n + align - 1) & ^(align - 1)
}

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
	addType(p)
	for fn := p.fns; fn != nil; fn = fn.next {
		o := 0
		for vl := fn.locals; vl != nil; vl = vl.next {
			va := vl.v
			o += sizeOf(va.ty)
			vl.v.offset = o
		}
		fn.stackSize = alignTo(o, 8)
	}
	codegen(p)
	os.Exit(0)
}
