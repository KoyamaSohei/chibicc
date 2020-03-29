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
	n := program()
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")
	fmt.Printf("  push rbp\n")
	fmt.Printf("  mov rbp, rsp\n")
	fmt.Printf("  sub rsp, 208\n")
	for s := n; s != nil; s = s.next {
		gen(s)
	}
	fmt.Printf(".Lreturn:\n")
	fmt.Printf("  mov rsp, rbp\n")
	fmt.Printf("  pop rbp\n")
	fmt.Printf("  ret\n")
	os.Exit(0)
}
