package main

func visit(n *node) {
	if n == nil {
		return
	}
	visit(n.lhs)
	visit(n.rhs)
	visit(n.cond)
	visit(n.then)
	visit(n.els)
	visit(n.init)
	visit(n.inc)
	for b := n.body; b != nil; b = b.next {
		visit(b)
	}
	for a := n.args; a != nil; a = a.next {
		visit(a)
	}
	switch n.kind {
	case ndMul:
		fallthrough
	case ndDiv:
		fallthrough
	case ndEq:
		fallthrough
	case ndNe:
		fallthrough
	case ndLt:
		fallthrough
	case ndLe:
		fallthrough
	case ndFunCall:
		fallthrough
	case ndNum:
		n.ty = &typ{kind: tyInt}
		return
	case ndLvar:
		n.ty = n.lv.ty
		return
	case ndAdd:
		if n.rhs.ty.kind == tyPtr {
			tmp := n.lhs
			n.lhs = n.rhs
			n.rhs = tmp
		}
		if n.rhs.ty.kind == tyPtr {
			errorTok(n.tok, "invalid pointer arithmetic operands")
		}
		n.ty = n.lhs.ty
		return
	case ndSub:
		if n.rhs.ty.kind == tyPtr {
			errorTok(n.tok, "invalid pointer arithmetic operands")
		}
		n.ty = n.lhs.ty
		return
	case ndAssign:
		n.ty = n.lhs.ty
		return
	case ndAddr:
		n.ty = &typ{kind: tyPtr, base: n.lhs.ty}
		return
	case ndDeref:
		if n.lhs.ty.kind != tyPtr {
			errorTok(n.tok, "invalid pointer dereference")
		}
		n.ty = n.lhs.ty.base
		return
	}

}

func addType(p *fun) {
	for fn := p; fn != nil; fn = fn.next {
		for n := fn.node; n != nil; n = n.next {
			visit(n)
		}
	}
}
