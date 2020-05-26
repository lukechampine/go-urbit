package ast

import (
	"fmt"
	"io"

	"lukechampine.com/go-urbit/hoon/token"
)

type Node interface {
	isNode()
}

func (Buc) isNode()        {}
func (Pat) isNode()        {}
func (Dot) isNode()        {}
func (Face) isNode()       {}
func (FacedValue) isNode() {}
func (Num) isNode()        {}
func (Rune) isNode()       {}
func (Cell) isNode()       {}

type Buc struct {
	Tok token.Token
}

type Pat struct {
	Tok token.Token
}

type Dot struct {
	Tok token.Token
}

type Face struct {
	Tok  token.Token
	Name string
}

type FacedValue struct {
	Tok   token.Token
	Face  Face
	Value Node
}

type Num struct {
	Tok token.Token
	Int string
}

type Rune struct {
	Tok  token.Token
	Lit  string
	Args []Node
}

type Cell struct {
	Tok  token.Token
	Head Node
	Tail Node
}

func Print(w io.Writer, n Node) (err error) {
	writeString := func(str string) {
		if err == nil {
			_, err = io.WriteString(w, str)
		}
	}
	writeNode := func(p Node) {
		if err == nil {
			err = Print(w, p)
		}
	}
	switch n := n.(type) {
	case Buc:
		writeString("$")
	case Pat:
		writeString("@")
	case Dot:
		writeString(".")
	case Face:
		writeString(n.Name)
	case FacedValue:
		writeNode(n.Face)
		writeString("=")
		writeNode(n.Value)
	case Num:
		writeString(n.Int)
	case Rune:
		switch n.Lit {
		case "%=":
			writeString(n.Lit)
			writeString("(")
			writeNode(n.Args[0])
			for i := 1; i < len(n.Args); i += 2 {
				if i > 1 {
					writeString(",")
				}
				writeString(" ")
				writeNode(n.Args[i])
				writeString(" ")
				writeNode(n.Args[i+1])
			}
			writeString(")")
		case "%-":
			writeString("(")
			for i, arg := range n.Args {
				if i > 0 {
					writeString(" ")
				}
				writeNode(arg)
			}
			writeString(")")
		default:
			writeString(n.Lit)
			writeString("(")
			for i, arg := range n.Args {
				if i > 0 {
					writeString(" ")
				}
				writeNode(arg)
			}
			writeString(")")
		}
	case Cell:
		writeString("[")
		writeNode(n.Head)
		writeString(" ")
		writeNode(n.Tail)
		writeString("]")
	default:
		panic(fmt.Sprintf("unknown node type %T", n))
	}
	return
}
