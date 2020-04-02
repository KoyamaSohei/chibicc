package main

func charType() *typ {
	return &typ{kind: tyChar}
}

func intType() *typ {
	return &typ{kind: tyInt}
}

func pointerTo(b *typ) *typ {
	return &typ{kind: tyPtr, base: b}
}

func arrayOf(b *typ, s int) *typ {
	return &typ{kind: tyArray, base: b, arraySize: s}
}

func sizeOf(ty *typ) int {
	switch ty.kind {
	case tyChar:
		return 1
	case tyInt:
		fallthrough
	case tyPtr:
		return 8
	}
	return sizeOf(ty.base) * ty.arraySize
}

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
		n.ty = intType()
		return
	case ndVar:
		n.ty = n.v.ty
		return
	case ndAdd:
		if n.rhs.ty.base != nil {
			tmp := n.lhs
			n.lhs = n.rhs
			n.rhs = tmp
		}
		if n.rhs.ty.base != nil {
			errorTok(n.tok, "invalid pointer arithmetic operands")
		}
		n.ty = n.lhs.ty
		return
	case ndSub:
		if n.rhs.ty.base != nil {
			errorTok(n.tok, "invalid pointer arithmetic operands")
		}
		n.ty = n.lhs.ty
		return
	case ndAssign:
		n.ty = n.lhs.ty
		return
	case ndAddr:
		if n.lhs.ty.kind == tyArray {
			n.ty = pointerTo(n.lhs.ty.base)
		} else {
			n.ty = pointerTo(n.lhs.ty)
		}
		return
	case ndDeref:
		if n.lhs.ty.base == nil {
			errorTok(n.tok, "invalid pointer dereference")
		}
		n.ty = n.lhs.ty.base
		return
	case ndSizeOf:
		n.kind = ndNum
		n.ty = intType()
		n.val = sizeOf(n.lhs.ty)
		n.lhs = nil
		return
	}

}

func addType(p *prog) {
	for fn := p.fns; fn != nil; fn = fn.next {
		for n := fn.node; n != nil; n = n.next {
			visit(n)
		}
	}
}
