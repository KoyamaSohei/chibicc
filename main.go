package main

import (
	"fmt"
	"os"
)

type tokenKind int

const (
	tkReserved tokenKind = iota
	tkNum
	tkEOF
)

type token struct {
	kind tokenKind
	next *token
	val  int
	str  []rune
}

var t = &token{}

func comsume(op rune) bool {
	if c := t.str[0]; t.kind != tkReserved || c != op {
		return false
	}
	t = t.next
	return true
}

func expect(op rune) {
	if c := t.str[0]; t.kind != tkReserved || c != op {
		err := fmt.Errorf("'%c' is not '%c'", c, op)
		fmt.Println(err)
		os.Exit(1)
	}
	t = t.next
}

func expectNumber() int {
	if t.kind != tkNum {
		err := fmt.Errorf("this is not number")
		fmt.Println(err)
		os.Exit(1)
	}
	v := t.val
	t = t.next
	return v
}

func atEOF() bool {
	return t.kind == tkEOF
}

func newToken(k tokenKind, cur *token, str []rune) *token {
	p := &token{kind: k, str: str}
	cur.next = p
	return p
}

func isDigit(c rune) bool {
	switch c {
	case '0':
		fallthrough
	case '1':
		fallthrough
	case '2':
		fallthrough
	case '3':
		fallthrough
	case '4':
		fallthrough
	case '5':
		fallthrough
	case '6':
		fallthrough
	case '7':
		fallthrough
	case '8':
		fallthrough
	case '9':
		return true
	default:
		return false
	}
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
		if c == '+' || c == '-' {
			cur = newToken(tkReserved, cur, p)
			p = p[1:]
			continue
		}
		if isDigit(c) {
			cur = newToken(tkNum, cur, p)
			v, err := strtoi(&p)
			if err != nil {
				err := fmt.Errorf("parse error")
				fmt.Println(err)
				os.Exit(1)
			}
			cur.val = v
			continue
		}
		e := fmt.Errorf("cannot tokenize")
		fmt.Println(e)
		os.Exit(1)
	}
	newToken(tkEOF, cur, p)
	return h.next
}

func main() {
	if len(os.Args) != 2 {
		e := fmt.Errorf("%s: invalid number of arguments", os.Args[0])
		fmt.Println(e)
		os.Exit(1)
	}
	s := os.Args[1]
	t = tokenize([]rune(s))
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")
	fmt.Printf("  mov rax, %d\n", expectNumber())
	for !atEOF() {
		if comsume('+') {
			fmt.Printf("  add rax, %d\n", expectNumber())
			continue
		}
		expect('-')
		fmt.Printf("  sub rax, %d\n", expectNumber())
	}
	fmt.Printf("  ret\n")
	os.Exit(0)
}
