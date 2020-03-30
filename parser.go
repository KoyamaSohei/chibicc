package main

import (
	"fmt"
	"os"
	"reflect"
)

type nodeKind int

const (
	ndAdd nodeKind = iota
	ndSub
	ndMul
	ndDiv
	ndEq
	ndNe
	ndLt
	ndLe
	ndAssign
	ndRet
	ndExprStmt
	ndLvar
	ndNum
)

type lvar struct {
	next   *lvar
	name   []rune
	offset int
}

type node struct {
	kind nodeKind
	next *node
	lhs  *node
	rhs  *node
	lv   *lvar
	val  int
}

type prog struct {
	node      *node
	locals    *lvar
	stackSize int
}

var locals *lvar

func findLvar(tok *token) *lvar {
	for v := locals; v != nil; v = v.next {
		if len(v.name) == tok.len && reflect.DeepEqual(tok.str[:tok.len], v.name) {
			return v
		}
	}
	return nil
}

func newUnary(k nodeKind, n *node) *node {
	return &node{kind: k, lhs: n}
}

func newBinary(k nodeKind, lhs *node, rhs *node) *node {
	return &node{kind: k, lhs: lhs, rhs: rhs}
}

func newNumber(v int) *node {
	return &node{kind: ndNum, val: v}
}

func newLvar(v *lvar) *node {
	return &node{kind: ndLvar, lv: v}
}

func pushLvar(name []rune) *lvar {
	v := &lvar{next: locals, name: name}
	locals = v
	return v
}

func primary() *node {
	if consume([]rune("(")) {
		n := expr()
		expect([]rune(")"))
		return n
	}

	if tok := consumeIdent(); tok != nil {
		v := findLvar(tok)
		if v == nil {
			v = pushLvar(tok.str[:tok.len])
		}
		return newLvar(v)
	}
	return newNumber(expectNumber())
}

func unary() *node {
	if consume([]rune("+")) {
		return unary()
	}
	if consume([]rune("-")) {
		return newBinary(ndSub, newNumber(0), unary())
	}
	return primary()
}

func mul() *node {
	n := unary()
	for {
		if consume([]rune("*")) {
			n = newBinary(ndMul, n, unary())
		} else if consume([]rune("/")) {
			n = newBinary(ndDiv, n, unary())
		} else {
			return n
		}
	}
}

func add() *node {
	n := mul()
	for {
		if consume([]rune("+")) {
			n = newBinary(ndAdd, n, mul())
		} else if consume([]rune("-")) {
			n = newBinary(ndSub, n, mul())
		} else {
			return n
		}
	}
}

func relational() *node {
	n := add()
	for {
		if consume([]rune("<")) {
			n = newBinary(ndLt, n, add())
		} else if consume([]rune("<=")) {
			n = newBinary(ndLe, n, add())
		} else if consume([]rune(">")) {
			n = newBinary(ndLt, add(), n)
		} else if consume([]rune(">=")) {
			n = newBinary(ndLe, add(), n)
		} else {
			return n
		}
	}
}

func equality() *node {
	n := relational()
	for {
		if consume([]rune("==")) {
			n = newBinary(ndEq, n, relational())
		} else if consume([]rune("!=")) {
			n = newBinary(ndNe, n, relational())
		} else {
			return n
		}
	}
}

func assign() *node {
	n := equality()
	if consume([]rune("=")) {
		n = newBinary(ndAssign, n, assign())
	}
	return n
}

func expr() *node {
	return assign()
}

func stmt() *node {
	if consume([]rune("return")) {
		n := newUnary(ndRet, expr())
		expect([]rune(";"))
		return n
	}
	n := newUnary(ndExprStmt, expr())
	expect([]rune(";"))
	return n
}

func program() *prog {
	var h node
	cur := &h
	for !atEOF() {
		cur.next = stmt()
		cur = cur.next
	}
	return &prog{node: h.next, locals: locals}
}

func genAddr(n *node) {
	if n.kind != ndLvar {
		fmt.Fprintln(os.Stderr, fmt.Errorf("not an lvalue"))
		os.Exit(1)
	}
	fmt.Printf("  lea rax, [rbp-%d]\n", n.lv.offset)
	fmt.Printf("  push rax\n")
}

func load() {
	fmt.Printf("  pop rax\n")
	fmt.Printf("  mov rax, [rax]\n")
	fmt.Printf("  push rax\n")
}

func store() {
	fmt.Printf("  pop rdi\n")
	fmt.Printf("  pop rax\n")
	fmt.Printf("  mov [rax], rdi\n")
	fmt.Printf("  push rdi\n")
}

func gen(n *node) {
	switch n.kind {
	case ndNum:
		fmt.Printf("  push %d\n", n.val)
		return
	case ndExprStmt:
		gen(n.lhs)
		fmt.Printf("  add rsp, 8\n")
		return
	case ndLvar:
		genAddr(n)
		load()
		return
	case ndAssign:
		genAddr(n.lhs)
		gen(n.rhs)
		store()
		return
	case ndRet:
		gen(n.lhs)
		fmt.Printf("  pop rax\n")
		fmt.Printf("  jmp .Lreturn\n")
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
	case ndEq:
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  sete al\n")
		fmt.Printf("  movzb rax, al\n")
	case ndNe:
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  setne al\n")
		fmt.Printf("  movzb rax, al\n")
	case ndLt:
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  setl al\n")
		fmt.Printf("  movzb rax, al\n")
	case ndLe:
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  setle al\n")
		fmt.Printf("  movzb rax, al\n")
	}
	fmt.Printf("  push rax\n")
}
