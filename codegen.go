package main

import (
	"fmt"
)

var (
	labelSeq = 0
	funcname []rune
	argreg   = [6]string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}
)

func genAddr(n *node) {
	switch n.kind {
	case ndVar:
		if v := n.v; v.isLocal {
			fmt.Printf("  lea rax, [rbp-%d]\n", v.offset)
			fmt.Printf("  push rax\n")
		} else {
			fmt.Printf("  push offset %s\n", string(v.name))
		}
		return
	case ndDeref:
		gen(n.lhs)
		return
	}
	errorTok(n.tok, "not an lvalue")
}

func genLval(n *node) {
	if n.ty.kind == tyArray {
		errorTok(n.tok, "not an lvalue")
	}
	genAddr(n)
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
	case ndNull:
		return
	case ndNum:
		fmt.Printf("  push %d\n", n.val)
		return
	case ndExprStmt:
		gen(n.lhs)
		fmt.Printf("  add rsp, 8\n")
		return
	case ndVar:
		genAddr(n)
		if n.ty.kind != tyArray {
			load()
		}
		return
	case ndAssign:
		genLval(n.lhs)
		gen(n.rhs)
		store()
		return
	case ndAddr:
		genAddr(n.lhs)
		return
	case ndDeref:
		gen(n.lhs)
		if n.ty.kind != tyArray {
			load()
		}
		return
	case ndIf:
		seq := labelSeq
		labelSeq++
		if n.els != nil {
			gen(n.cond)
			fmt.Printf("  pop rax\n")
			fmt.Printf("  cmp rax, 0\n")
			fmt.Printf("  je .Lelse%d\n", seq)
			gen(n.then)
			fmt.Printf("  jmp .Lend%d\n", seq)
			fmt.Printf(".Lelse%d:\n", seq)
			gen(n.els)
			fmt.Printf(".Lend%d:\n", seq)
		} else {
			gen(n.cond)
			fmt.Printf("  pop rax\n")
			fmt.Printf("  cmp rax, 0\n")
			fmt.Printf("  je .Lend%d\n", seq)
			gen(n.then)
			fmt.Printf(".Lend%d:\n", seq)
		}
		return
	case ndWhile:
		seq := labelSeq
		labelSeq++
		fmt.Printf(".Lbegin%d:\n", seq)
		gen(n.cond)
		fmt.Printf("  pop rax\n")
		fmt.Printf("  cmp rax, 0\n")
		fmt.Printf("  je .Lend%d\n", seq)
		gen(n.then)
		fmt.Printf("  jmp .Lbegin%d\n", seq)
		fmt.Printf(".Lend%d:\n", seq)
		return
	case ndFor:
		seq := labelSeq
		labelSeq++
		if n.init != nil {
			gen(n.init)
		}
		fmt.Printf(".Lbegin%d:\n", seq)
		if n.cond != nil {
			gen(n.cond)
			fmt.Printf("  pop rax\n")
			fmt.Printf("  cmp rax, 0\n")
			fmt.Printf("  je .Lend%d\n", seq)
		}
		gen(n.then)
		if n.inc != nil {
			gen(n.inc)
		}
		fmt.Printf("  jmp .Lbegin%d\n", seq)
		fmt.Printf(".Lend%d:\n", seq)
		return
	case ndBlock:
		for b := n.body; b != nil; b = b.next {
			gen(b)
		}
		return
	case ndFunCall:
		c := 0
		for arg := n.args; arg != nil; arg = arg.next {
			gen(arg)
			c++
		}
		for c > 0 {
			c--
			fmt.Printf("  pop %s\n", argreg[c])
		}
		seq := labelSeq
		labelSeq++
		fmt.Printf("  mov rax, rsp\n")
		fmt.Printf("  and rax, 15\n")
		fmt.Printf("  jnz .Lcall%d\n", seq)
		fmt.Printf("  mov rax, 0\n")
		fmt.Printf("  call %s\n", string(n.funcname))
		fmt.Printf("  jmp .Lend%d\n", seq)
		fmt.Printf(".Lcall%d:\n", seq)
		fmt.Printf("  sub rsp, 8\n")
		fmt.Printf("  mov rax, 0\n")
		fmt.Printf("  call %s\n", string(n.funcname))
		fmt.Printf("  add rsp, 8\n")
		fmt.Printf(".Lend%d:\n", seq)
		fmt.Printf("  push rax\n")
		return
	case ndRet:
		gen(n.lhs)
		fmt.Printf("  pop rax\n")
		fmt.Printf("  jmp .Lreturn.%s\n", string(funcname))
		return
	}
	gen(n.lhs)
	gen(n.rhs)
	fmt.Printf("  pop rdi\n")
	fmt.Printf("  pop rax\n")

	switch n.kind {
	case ndAdd:
		if n.ty.base != nil {
			fmt.Printf("  imul rdi, %d\n", sizeOf(n.ty.base))
		}
		fmt.Printf("  add rax, rdi\n")
	case ndSub:
		if n.ty.base != nil {
			fmt.Printf("  imul rdi, %d\n", sizeOf(n.ty.base))
		}
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

func emitData(p *prog) {
	fmt.Printf(".data\n")
	for vl := p.globals; vl != nil; vl = vl.next {
		v := vl.v
		fmt.Printf("%s:\n", string(v.name))
		fmt.Printf("  .zero %d\n", sizeOf(v.ty))
	}
}

func emitText(p *prog) {
	fmt.Printf(".text\n")
	for fn := p.fns; fn != nil; fn = fn.next {
		fmt.Printf(".global %s\n", string(fn.name))
		fmt.Printf("%s:\n", string(fn.name))
		funcname = fn.name
		fmt.Printf("  push rbp\n")
		fmt.Printf("  mov rbp, rsp\n")
		fmt.Printf("  sub rsp, %d\n", fn.stackSize)
		i := 0
		for vl := fn.params; vl != nil; vl = vl.next {
			v := vl.v
			fmt.Printf("  mov [rbp-%d], %s\n", v.offset, argreg[i])
			i++
		}
		for n := fn.node; n != nil; n = n.next {
			gen(n)
		}
		fmt.Printf(".Lreturn.%s:\n", string(funcname))
		fmt.Printf("  mov rsp, rbp\n")
		fmt.Printf("  pop rbp\n")
		fmt.Printf("  ret\n")
	}
}

func codegen(p *prog) {
	fmt.Printf(".intel_syntax noprefix\n")
	emitData(p)
	emitText(p)
}
