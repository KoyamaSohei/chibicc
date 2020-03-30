package main

import (
	"fmt"
	"os"
)

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

func codegen(p *prog) {
	o := 0
	for v := p.locals; v != nil; v = v.next {
		o += 8
		v.offset = o
	}
	p.stackSize = o
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")
	fmt.Printf("  push rbp\n")
	fmt.Printf("  mov rbp, rsp\n")
	fmt.Printf("  sub rsp, %d\n", p.stackSize)
	for s := p.node; s != nil; s = s.next {
		gen(s)
	}
	fmt.Printf(".Lreturn:\n")
	fmt.Printf("  mov rsp, rbp\n")
	fmt.Printf("  pop rbp\n")
	fmt.Printf("  ret\n")
}
