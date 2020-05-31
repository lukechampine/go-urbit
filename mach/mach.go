package mach

import (
	"fmt"
	"sort"
	"strconv"

	"lukechampine.com/urbit/hoon/ast"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

var atomType = types.I32

func Transpile(n ast.Node) string {
	t := newTranspiler()
	main := t.module.NewFunc("main", atomType)
	entry := main.NewBlock("")
	maybeRet(entry, t.expr(subject{}, main, entry, n))
	return t.Finish()
}

type subject struct {
	m map[string]value.Value
}

func (s subject) get(key string) value.Value {
	return s.m[key]
}

func (s subject) with(key string, val value.Value) subject {
	m2 := make(map[string]value.Value, len(s.m))
	for k, v := range s.m {
		m2[k] = v
	}
	m2[key] = val
	return subject{m2}
}

func (s subject) faces() []string {
	fs := make([]string, 0, len(s.m))
	for k := range s.m {
		fs = append(fs, k)
	}
	sort.Strings(fs)
	return fs
}

type transpiler struct {
	module *ir.Module
	id     int64
}

func newTranspiler() *transpiler {
	m := ir.NewModule()
	return &transpiler{
		module: m,
	}
}

func maybeRet(b *ir.Block, v value.Value) {
	if v != nil {
		b.NewRet(v)
	}
}

func (t *transpiler) Finish() string {
	return t.module.String()
}

func (t *transpiler) uid() int64 {
	id := t.id
	t.id++
	return id
}

func sampleParams(n ast.Node) []*ir.Param {
	switch n := n.(type) {
	case ast.Tis:
		sampleName := n.Left.(ast.Face).Name
		var sampleType types.Type
		switch n.Right.(type) {
		case ast.Pat:
			sampleType = atomType
		default:
			panic("unhandled sample type")
		}
		return []*ir.Param{ir.NewParam(sampleName, sampleType)}
	case ast.Cell:
		return append(sampleParams(n.Head), sampleParams(n.Tail)...)
	default:
		panic("unhandled sample mold")
	}
}

func (t *transpiler) expr(s subject, fn *ir.Func, b *ir.Block, n ast.Node) value.Value {
	eval := func(n ast.Node) value.Value {
		return t.expr(s, fn, b, n)
	}

	switch n := n.(type) {
	case ast.Num:
		i, _ := strconv.Atoi(n.Int)
		return constant.NewInt(atomType, int64(i))
	case ast.Face:
		return s.get(n.Name)
	case ast.Buc:
		return fn
	case ast.Rune:
		switch n.Lit {
		case ".=":
			return b.NewICmp(
				enum.IPredEQ,
				eval(n.Args[0]),
				eval(n.Args[1]),
			)
		case ".+":
			return b.NewAdd(
				eval(n.Args[0]),
				constant.NewInt(atomType, 1),
			)
		case "=/":
			face := n.Args[0].(ast.Face)
			f := eval(n.Args[1])
			return t.expr(s.with(face.Name, f), fn, b, n.Args[2])
		case "=.":
			face := n.Args[0].(ast.Face)
			s = s.with(face.Name, eval(n.Args[1]))
			return eval(n.Args[2])
		case "?:":
			pred := eval(n.Args[0])
			tb := fn.NewBlock("")
			fb := fn.NewBlock("")
			b.NewCondBr(pred, tb, fb)
			maybeRet(tb, t.expr(s, fn, tb, n.Args[1]))
			maybeRet(fb, t.expr(s, fn, fb, n.Args[2]))
			return nil
		case "|-":
			// everything in the subject becomes an argument.
			var params []*ir.Param
			var args []value.Value
			var fs subject
			for _, face := range s.faces() {
				v := s.get(face)
				p := ir.NewParam(face, v.Type())
				params = append(params, p)
				args = append(args, v)
				fs = fs.with(face, p)
			}
			fn := t.module.NewFunc("", atomType, params...)
			fn.SetID(t.uid())
			fn.Linkage = enum.LinkagePrivate
			fnb := fn.NewBlock("")
			maybeRet(fnb, t.expr(fs, fn, fnb, n.Args[0]))
			return b.NewCall(fn, args...)
		case "|=":
			// everything in the subject becomes an argument, as well as the gate's arguments
			var params []*ir.Param
			var fs subject
			for _, face := range s.faces() {
				v := s.get(face)
				p := ir.NewParam(face, v.Type())
				params = append(params, p)
				fs = fs.with(face, p)
			}
			sample := sampleParams(n.Args[0])
			params = append(params, sample...)
			for _, p := range sample {
				fs = fs.with(p.Name(), p)
			}

			fn := t.module.NewFunc("", atomType, params...)
			fn.SetID(t.uid())
			fn.Linkage = enum.LinkagePrivate
			fnb := fn.NewBlock("")
			maybeRet(fnb, t.expr(fs, fn, fnb, n.Args[1]))
			return fn
		case "%-":
			switch gate := n.Args[0].(type) {
			case ast.Dot:
				f := fn
				args := make([]value.Value, len(f.Params)-len(n.Args[1:]), len(f.Params))
				for i := range args {
					args[i] = s.get(f.Params[i].Name())
				}
				for _, a := range n.Args[1:] {
					args = append(args, eval(a))
				}
				return b.NewCall(f, args...)
			case ast.Rune:
				f := eval(n.Args[0]).(*ir.Func)
				args := make([]value.Value, len(f.Params)-len(n.Args[1:]), len(f.Params))
				for i := range args {
					args[i] = s.get(f.Params[i].Name())
				}
				for _, a := range n.Args[1:] {
					args = append(args, eval(a))
				}
				return b.NewCall(f, args...)
			case ast.Face:
				switch gate.Name {
				case "dec":
					return b.NewSub(eval(n.Args[1]), constant.NewInt(atomType, 1))
				case "add":
					return b.NewAdd(eval(n.Args[1]), eval(n.Args[2]))
				case "sub":
					return b.NewSub(eval(n.Args[1]), eval(n.Args[2]))
				case "mul":
					return b.NewMul(eval(n.Args[1]), eval(n.Args[2]))
				default:
					v := s.get(gate.Name)
					if v == nil {
						panic(fmt.Sprintf("unknown gate %s", gate.Name))
					}
					f := v.(*ir.Func)
					if f == nil {
						panic(fmt.Sprintf("not a gate: %s", gate.Name))
					}
					args := make([]value.Value, len(f.Params)-len(n.Args[1:]), len(f.Params))
					for i := range args {
						args[i] = s.get(f.Params[i].Name())
					}
					for _, a := range n.Args[1:] {
						args = append(args, eval(a))
					}
					return b.NewCall(f, args...)
				}
			default:
				panic(fmt.Sprintf("bad gate type %T", gate))
			}
		case "%=":
			if _, ok := n.Args[0].(ast.Buc); ok {
				args := make([]value.Value, len(fn.Params))

				for i := range args {
					// by default, use existing value in subject
					args[i] = s.get(fn.Params[i].Name())

					// if new value provided in centis, use that instead
					for j := 1; j < len(n.Args); j += 2 {
						face := n.Args[j].(ast.Face)
						if face.Name == fn.Params[i].Name() {
							args[i] = eval(n.Args[j+1])
							break
						}
					}
				}
				return b.NewCall(fn, args...)
			}
			panic("unhandled centis")
		default:
			panic("unhandled rune " + n.Lit)
		}
	default:
		panic(fmt.Sprintf("unhandled node type %T", n))
	}
}
