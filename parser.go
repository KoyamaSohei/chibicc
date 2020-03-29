package main

import "fmt"

type nodeKind int

const (
	ndAdd nodeKind = iota
	ndSub
	ndMul
	ndDiv
	ndNum
)

type node struct {
	kind nodeKind
	lhs  *node
	rhs  *node
	val  int
}

func newBinary(k nodeKind, lhs *node, rhs *node) *node {
	return &node{kind: k, lhs: lhs, rhs: rhs}
}

func newNumber(v int) *node {
	return &node{kind: ndNum, val: v}
}

func primary() *node {
	if consume('(') {
		n := expr()
		expect(')')
		return n
	}
	return newNumber(expectNumber())
}

func unary() *node {
	if consume('+') {
		return unary()
	}
	if consume('-') {
		return newBinary(ndSub, newNumber(0), unary())
	}
	return primary()
}

func mul() *node {
	n := unary()
	for {
		if consume('*') {
			n = newBinary(ndMul, n, unary())
		} else if consume('/') {
			n = newBinary(ndDiv, n, unary())
		} else {
			return n
		}
	}
}

func expr() *node {
	n := mul()
	for {
		if consume('+') {
			n = newBinary(ndAdd, n, mul())
		} else if consume('-') {
			n = newBinary(ndSub, n, mul())
		} else {
			return n
		}
	}
}

func gen(n *node) {
	if n.kind == ndNum {
		fmt.Printf("  push %d\n", n.val)
		return
	}
	gen(n.lhs)
	gen(n.rhs)
	fmt.Printf("  pop rdi\n")
	fmt.Printf("  pop rax\n")

	switch n.kind {
	case ndAdd:
		fmt.Printf("  add rax, rdi\n")
	case ndSub:
		fmt.Printf("  sub rax, rdi\n")
	case ndMul:
		fmt.Printf("  imul rax, rdi\n")
	case ndDiv:
		fmt.Printf("  cqo\n")
		fmt.Printf("  idiv rdi\n")
	}
	fmt.Printf("  push rax\n")
}
