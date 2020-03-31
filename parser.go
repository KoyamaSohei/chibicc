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
	ndAddr
	ndDeref
	ndRet
	ndIf
	ndWhile
	ndFor
	ndSizeOf
	ndBlock
	ndFunCall
	ndExprStmt
	ndLvar
	ndNum
	ndNull
)

type lvar struct {
	name   []rune
	ty     *typ
	offset int
}

type varlist struct {
	next *varlist
	lvar *lvar
}

type node struct {
	kind     nodeKind
	next     *node
	ty       *typ
	tok      *token
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

type typeKind int

const (
	tyInt typeKind = iota
	tyPtr
	tyArray
)

type typ struct {
	kind      typeKind
	base      *typ
	arraySize int
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

func newUnary(k nodeKind, n *node, tok *token) *node {
	return &node{kind: k, lhs: n, tok: tok}
}

func newBinary(k nodeKind, lhs *node, rhs *node, tok *token) *node {
	return &node{kind: k, lhs: lhs, rhs: rhs, tok: tok}
}

func newNumber(v int, tok *token) *node {
	return &node{kind: ndNum, val: v, tok: tok}
}

func newLvar(v *lvar, tok *token) *node {
	return &node{kind: ndLvar, lv: v, tok: tok}
}

func pushLvar(name []rune, ty *typ) *lvar {
	v := &lvar{name: name, ty: ty}
	vl := &varlist{lvar: v, next: locals}
	locals = vl
	return v
}

func primary() *node {
	if consume([]rune("(")) != nil {
		n := expr()
		expect([]rune(")"))
		return n
	}
	if tok := consume([]rune("sizeof")); tok != nil {
		return newUnary(ndSizeOf, unary(), tok)
	}

	if tok := consumeIdent(); tok != nil {
		if consume([]rune("(")) != nil {
			return &node{kind: ndFunCall, funcname: tok.str[:tok.len], args: funcArgs(), tok: tok}
		}
		v := findLvar(tok)
		if v == nil {
			errorTok(tok, "undefined variable")
		}
		return newLvar(v, tok)
	}
	tok := t
	if tok.kind != tkNum {
		errorTok(tok, "expected expression")
	}
	return newNumber(expectNumber(), tok)
}

func funcArgs() *node {
	if consume([]rune(")")) != nil {
		return nil
	}
	h := assign()
	cur := h
	for consume([]rune(",")) != nil {
		cur.next = assign()
		cur = cur.next
	}
	expect([]rune(")"))
	return h
}

func unary() *node {
	if consume([]rune("+")) != nil {
		return unary()
	}
	if tok := consume([]rune("-")); tok != nil {
		return newBinary(ndSub, newNumber(0, tok), unary(), tok)
	}
	if tok := consume([]rune("&")); tok != nil {
		return newUnary(ndAddr, unary(), tok)
	}
	if tok := consume([]rune("*")); tok != nil {
		return newUnary(ndDeref, unary(), tok)
	}
	return primary()
}

func mul() *node {
	n := unary()
	for {
		if tok := consume([]rune("*")); tok != nil {
			n = newBinary(ndMul, n, unary(), tok)
		} else if tok := consume([]rune("/")); tok != nil {
			n = newBinary(ndDiv, n, unary(), tok)
		} else {
			return n
		}
	}
}

func add() *node {
	n := mul()
	for {
		if tok := consume([]rune("+")); tok != nil {
			n = newBinary(ndAdd, n, mul(), tok)
		} else if tok := consume([]rune("-")); tok != nil {
			n = newBinary(ndSub, n, mul(), tok)
		} else {
			return n
		}
	}
}

func relational() *node {
	n := add()
	for {
		if tok := consume([]rune("<")); tok != nil {
			n = newBinary(ndLt, n, add(), tok)
		} else if tok := consume([]rune("<=")); tok != nil {
			n = newBinary(ndLe, n, add(), tok)
		} else if tok := consume([]rune(">")); tok != nil {
			n = newBinary(ndLt, add(), n, tok)
		} else if tok := consume([]rune(">=")); tok != nil {
			n = newBinary(ndLe, add(), n, tok)
		} else {
			return n
		}
	}
}

func equality() *node {
	n := relational()
	for {
		if tok := consume([]rune("==")); tok != nil {
			n = newBinary(ndEq, n, relational(), tok)
		} else if tok := consume([]rune("!=")); tok != nil {
			n = newBinary(ndNe, n, relational(), tok)
		} else {
			return n
		}
	}
}

func assign() *node {
	n := equality()
	if tok := consume([]rune("=")); tok != nil {
		n = newBinary(ndAssign, n, assign(), tok)
	}
	return n
}

func expr() *node {
	return assign()
}

func stmt() *node {
	if tok := consume([]rune("return")); tok != nil {
		n := newUnary(ndRet, expr(), tok)
		expect([]rune(";"))
		return n
	}
	if tok := consume([]rune("if")); tok != nil {
		n := &node{kind: ndIf, tok: tok}
		expect([]rune("("))
		n.cond = expr()
		expect([]rune(")"))
		n.then = stmt()
		if consume([]rune("else")) != nil {
			n.els = stmt()
		}
		return n
	}
	if tok := consume([]rune("while")); tok != nil {
		n := &node{kind: ndWhile, tok: tok}
		expect([]rune("("))
		n.cond = expr()
		expect([]rune(")"))
		n.then = stmt()
		return n
	}
	if tok := consume([]rune("for")); tok != nil {
		n := &node{kind: ndFor, tok: tok}
		expect([]rune("("))
		if consume([]rune(";")) == nil {
			n.init = readExprStmt()
			expect([]rune(";"))
		}
		if consume([]rune(";")) == nil {
			n.cond = expr()
			expect([]rune(";"))
		}
		if consume([]rune(")")) == nil {
			n.inc = readExprStmt()
			expect([]rune(")"))
		}
		n.then = stmt()
		return n
	}
	if tok := consume([]rune("{")); tok != nil {
		var h node
		cur := &h
		for consume([]rune("}")) == nil {
			cur.next = stmt()
			cur = cur.next
		}
		n := &node{kind: ndBlock, tok: tok}
		n.body = h.next
		return n
	}
	if peek([]rune("int")) {
		return declaration()
	}
	n := readExprStmt()
	expect([]rune(";"))
	return n
}

func readExprStmt() *node {
	tt := t
	return newUnary(ndExprStmt, expr(), tt)
}

func declaration() *node {
	tok := t
	ty := baseType()
	name := expectIdent()
	ty = readTypeSuffix(ty)
	v := pushLvar(name, ty)
	if consume([]rune(";")) != nil {
		return &node{kind: ndNull, tok: tok}
	}
	expect([]rune("="))
	lhs := &node{kind: ndLvar, tok: tok, lv: v}
	rhs := expr()
	expect([]rune(";"))
	n := newBinary(ndAssign, lhs, rhs, tok)
	return newUnary(ndExprStmt, n, tok)
}

func function() *fun {
	locals = nil
	baseType()
	fn := &fun{name: expectIdent()}
	expect([]rune("("))
	fn.params = readFuncParams()
	expect([]rune("{"))
	var h node
	cur := &h
	for consume([]rune("}")) == nil {
		cur.next = stmt()
		cur = cur.next
	}
	fn.node = h.next
	fn.locals = locals
	return fn
}

func baseType() *typ {
	expect([]rune("int"))
	ty := intType()
	for consume([]rune("*")) != nil {
		ty = pointerTo(ty)
	}
	return ty
}

func readTypeSuffix(b *typ) *typ {
	if consume([]rune("[")) == nil {
		return b
	}
	sz := expectNumber()
	expect([]rune("]"))
	b = readTypeSuffix(b)
	return arrayOf(b, sz)
}

func readFuncParam() *varlist {
	ty := baseType()
	name := expectIdent()
	ty = readTypeSuffix(ty)
	return &varlist{lvar: pushLvar(name, ty)}
}

func readFuncParams() *varlist {
	if consume([]rune(")")) != nil {
		return nil
	}
	h := readFuncParam()
	cur := h
	for consume([]rune(")")) == nil {
		expect([]rune(","))
		cur.next = readFuncParam()
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
