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
	n := expr()
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")
	gen(n)
	fmt.Printf("  pop rax\n")
	fmt.Printf("  ret\n")
	os.Exit(0)
}
