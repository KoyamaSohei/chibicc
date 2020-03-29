package main

import (
	"fmt"
	"os"
	"reflect"
)

type tokenKind int

const (
	tkReserved tokenKind = iota
	tkIdent
	tkNum
	tkEOF
)

type token struct {
	kind tokenKind
	next *token
	val  int
	str  []rune
	len  int
}

var (
	t    = &token{}
	inpt = ""
)

func errorAt(loc []rune, f string, r ...[]rune) {
	p := len(loc) - len(inpt)
	e := fmt.Errorf(inpt)
	fmt.Fprintln(os.Stderr, e)
	e = fmt.Errorf("%*s", p, "")
	fmt.Fprint(os.Stderr, e)
	fmt.Fprintln(os.Stderr, "^ ")
	if len(r) == 0 {
		e = fmt.Errorf(f)
	} else {
		e = fmt.Errorf(f, string(r[0]))
	}
	fmt.Fprintln(os.Stderr, e)
	os.Exit(1)
}

func consume(op []rune) bool {
	if t.kind != tkReserved {
		return false
	}
	if len(op) != t.len {
		return false
	}
	if c := t.str[0:t.len]; !reflect.DeepEqual(c, op) {
		return false
	}
	t = t.next
	return true
}

func consumeIdent() *token {
	if t.kind != tkIdent {
		return nil
	}
	tt := t
	t = t.next
	return tt
}

func expect(op []rune) {
	if t.kind != tkReserved {
		errorAt(t.str, "expected '%s'", op)
	}
	if c := t.str[0:t.len]; !reflect.DeepEqual(c, op) {
		errorAt(t.str, "expected '%s'", op)
	}
	t = t.next
}

func expectNumber() int {
	if t.kind != tkNum {
		errorAt(t.str, "expected a number")
	}
	v := t.val
	t = t.next
	return v
}

func atEOF() bool {
	return t.kind == tkEOF
}

func newToken(k tokenKind, cur *token, str []rune, len int) *token {
	p := &token{kind: k, str: str, len: len}
	cur.next = p
	return p
}

func startWith(str []rune, op []rune) bool {
	if len(str) < len(op) {
		return false
	}
	return reflect.DeepEqual(str[0:len(op)], op)
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || c == '_'
}

func isAlNum(c rune) bool {
	return isAlpha(c) || isDigit(c)
}

func strtoi(p *[]rune) (int, error) {
	s := *p
	c := s[0]
	s = s[1:]
	neg := false
	for c == ' ' {
		if len(s) == 0 {
			return -1, fmt.Errorf("parse error at %c", c)
		}
		c = s[0]
		s = s[1:]
	}
	if c == '-' {
		neg = true
		if len(s) == 0 {
			return -1, fmt.Errorf("parse error at %c", c)
		}
		c = s[0]
		s = s[1:]
	}
	if !isDigit(c) {
		return -1, fmt.Errorf("parse error at %c", c)
	}
	acc := 0
	for {
		k := int(c - '0')
		acc *= 10
		acc += k
		if len(s) == 0 || !isDigit(s[0]) {
			break
		}
		c = s[0]
		s = s[1:]
	}
	if neg {
		acc *= -1
	}
	*p = s
	return acc, nil
}

func isReserved(c rune) bool {
	switch c {
	case '+':
		fallthrough
	case '-':
		fallthrough
	case '*':
		fallthrough
	case '/':
		fallthrough
	case '(':
		fallthrough
	case ')':
		fallthrough
	case '<':
		fallthrough
	case '>':
		fallthrough
	case ';':
		fallthrough
	case '=':
		return true
	default:
		return false
	}
}

func tokenize(p []rune) *token {
	var h token
	h.next = nil
	cur := &h
	for len(p) > 0 {
		c := p[0]
		if c == ' ' {
			p = p[1:]
			continue
		}
		if startWith(p, []rune("return")) && !isAlNum(p[6]) {
			cur = newToken(tkReserved, cur, p, 6)
			p = p[6:]
			continue
		}
		if startWith(p, []rune("==")) || startWith(p, []rune("!=")) || startWith(p, []rune("<=")) || startWith(p, []rune(">=")) {
			cur = newToken(tkReserved, cur, p, 2)
			p = p[2:]
			continue
		}
		if isReserved(c) {
			cur = newToken(tkReserved, cur, p, 1)
			p = p[1:]
			continue
		}
		if isAlpha(c) {
			cur = newToken(tkIdent, cur, p, 1)
			p = p[1:]
			continue
		}
		if isDigit(c) {
			cur = newToken(tkNum, cur, p, 0)
			l := len(p)
			v, err := strtoi(&p)
			if err != nil {
				errorAt(p, "parse error at %s", []rune{c})
			}
			cur.val = v
			cur.len = l - len(p)
			continue
		}
		errorAt(p, "cannot tokenize %s", []rune{c})
	}
	newToken(tkEOF, cur, p, 0)
	return h.next
}
