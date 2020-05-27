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

func (t *transpiler) expr(s subject, fn *ir.Func, b *ir.Block, n ast.Node) value.Value {
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
				t.expr(s, fn, b, n.Args[0]),
				t.expr(s, fn, b, n.Args[1]),
			)
		case ".+":
			return b.NewAdd(
				t.expr(s, fn, b, n.Args[0]),
				constant.NewInt(atomType, 1),
			)
		case "=/":
			face := n.Args[0].(ast.Face)
			f := t.expr(s, fn, b, n.Args[1])
			return t.expr(s.with(face.Name, f), fn, b, n.Args[2])
		case "=.":
			face := n.Args[0].(ast.Face)
			s = s.with(face.Name, t.expr(s, fn, b, n.Args[1]))
			return t.expr(s, fn, b, n.Args[2])
		case "?:":
			pred := t.expr(s, fn, b, n.Args[0])
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
			fn.Linkage = enum.LinkagePrivate
			fnb := fn.NewBlock("")
			maybeRet(fnb, t.expr(fs, fn, fnb, n.Args[0]))
			return b.NewCall(fn, args...)
		case "%-":
			gate := n.Args[0].(ast.Face)
			arg := func(i int) value.Value {
				return t.expr(s, fn, b, n.Args[i+1])
			}
			switch gate.Name {
			case "dec":
				return b.NewSub(arg(0), constant.NewInt(atomType, 1))
			case "add":
				return b.NewAdd(arg(0), arg(1))
			case "sub":
				return b.NewSub(arg(0), arg(1))
			case "mul":
				return b.NewMul(arg(0), arg(1))
			default:
				panic(fmt.Sprintf("unhandled gate %s", gate.Name))
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
							args[i] = t.expr(s, fn, b, n.Args[j+1])
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
