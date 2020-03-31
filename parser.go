package main

import (
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
	ndIf
	ndWhile
	ndFor
	ndBlock
	ndFunCall
	ndExprStmt
	ndLvar
	ndNum
)

type lvar struct {
	name   []rune
	offset int
}

type varlist struct {
	next *varlist
	lvar *lvar
}

type node struct {
	kind     nodeKind
	next     *node
	lhs      *node
	rhs      *node
	cond     *node
	then     *node
	els      *node
	init     *node
	inc      *node
	body     *node
	funcname []rune
	args     *node
	lv       *lvar
	val      int
}

type fun struct {
	next      *fun
	name      []rune
	params    *varlist
	node      *node
	locals    *varlist
	stackSize int
}

var locals *varlist

func findLvar(tok *token) *lvar {
	for vl := locals; vl != nil; vl = vl.next {
		v := vl.lvar
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
	v := &lvar{name: name}
	vl := &varlist{lvar: v, next: locals}
	locals = vl
	return v
}

func primary() *node {
	if consume([]rune("(")) {
		n := expr()
		expect([]rune(")"))
		return n
	}

	if tok := consumeIdent(); tok != nil {
		if consume([]rune("(")) {
			return &node{kind: ndFunCall, funcname: tok.str[:tok.len], args: funcArgs()}
		}
		v := findLvar(tok)
		if v == nil {
			v = pushLvar(tok.str[:tok.len])
		}
		return newLvar(v)
	}
	return newNumber(expectNumber())
}

func funcArgs() *node {
	if consume([]rune(")")) {
		return nil
	}
	h := assign()
	cur := h
	for consume([]rune(",")) {
		cur.next = assign()
		cur = cur.next
	}
	expect([]rune(")"))
	return h
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
	if consume([]rune("if")) {
		n := &node{kind: ndIf}
		expect([]rune("("))
		n.cond = expr()
		expect([]rune(")"))
		n.then = stmt()
		if consume([]rune("else")) {
			n.els = stmt()
		}
		return n
	}
	if consume([]rune("while")) {
		n := &node{kind: ndWhile}
		expect([]rune("("))
		n.cond = expr()
		expect([]rune(")"))
		n.then = stmt()
		return n
	}
	if consume([]rune("for")) {
		n := &node{kind: ndFor}
		expect([]rune("("))
		if !consume([]rune(";")) {
			n.init = readExprStmt()
			expect([]rune(";"))
		}
		if !consume([]rune(";")) {
			n.cond = expr()
			expect([]rune(";"))
		}
		if !consume([]rune(")")) {
			n.inc = readExprStmt()
			expect([]rune(")"))
		}
		n.then = stmt()
		return n
	}
	if consume([]rune("{")) {
		var h node
		cur := &h
		for !consume([]rune("}")) {
			cur.next = stmt()
			cur = cur.next
		}
		n := &node{kind: ndBlock}
		n.body = h.next
		return n
	}
	n := readExprStmt()
	expect([]rune(";"))
	return n
}

func readExprStmt() *node {
	return newUnary(ndExprStmt, expr())
}

func function() *fun {
	locals = nil
	fn := &fun{name: expectIdent()}
	expect([]rune("("))
	fn.params = readFuncParams()
	expect([]rune("{"))
	var h node
	cur := &h
	for !consume([]rune("}")) {
		cur.next = stmt()
		cur = cur.next
	}
	fn.node = h.next
	fn.locals = locals
	return fn
}

func readFuncParams() *varlist {
	if consume([]rune(")")) {
		return nil
	}
	h := &varlist{lvar: pushLvar(expectIdent())}
	cur := h
	for !consume([]rune(")")) {
		expect([]rune(","))
		cur.next = &varlist{lvar: pushLvar(expectIdent())}
		cur = cur.next
	}
	return h
}

func program() *fun {
	var h fun
	cur := &h
	for !atEOF() {
		cur.next = function()
		cur = cur.next
	}
	return h.next
}
