package ast

import (
	"fmt"
	"io"

	"lukechampine.com/urbit/hoon/token"
)

type Node interface {
	isNode()
}

func (Buc) isNode()  {}
func (Pat) isNode()  {}
func (Dot) isNode()  {}
func (Face) isNode() {}
func (Slot) isNode() {}
func (Tis) isNode()  {}
func (Num) isNode()  {}
func (Rune) isNode() {}
func (Cell) isNode() {}

type Buc struct {
	Tok token.Token
}

type Pat struct {
	Tok token.Token
}

// TODO: replace with wing?
type Dot struct {
	Tok token.Token
}

type Face struct {
	Tok  token.Token
	Name string
}

type Slot struct {
	Tok     token.Token
	Address string
}

type Tis struct {
	Tok   token.Token
	Left  Node
	Right Node
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

func (r Rune) Reduce() Rune {
	switch r.Lit {
	// these are the only "real" runes; every other rune reduces to one of these
	case "=~", "%=", "?:":
		return r

	// these runes are just flipped versions of other runes
	case "=<", "%.":
		return Rune{
			Lit: map[string]string{
				"=<": "=>",
				"=-": "=+",
				"%.": "%-",
				"?.": "?:",
			}[r.Lit],
			Args: []Node{
				r.Args[1],
				r.Args[0],
			},
		}.Reduce()

	// tis
	case "=>":
		return Rune{
			Lit:  "=~",
			Args: r.Args,
		}.Reduce()
	case "=+":
		return Rune{
			Lit: "=>",
			Args: []Node{
				Cell{
					Head: r.Args[0],
					Tail: Dot{},
				},
				r.Args[1],
			},
		}.Reduce()
	case "=.", "=:":
		return Rune{
			Lit: "=>",
			Args: []Node{
				Rune{
					Lit: "%_",
					Args: append([]Node{
						Dot{},
					}, r.Args[:len(r.Args)-1]...),
				},
				r.Args[len(r.Args)-1],
			},
		}.Reduce()
	case "=|":
		return Rune{
			Lit: "=+",
			Args: []Node{
				Rune{
					Lit:  "^*",
					Args: []Node{r.Args[0]},
				},
				r.Args[1],
			},
		}.Reduce()
	case "=/":
		return Rune{
			Lit: "=+",
			Args: []Node{
				Tis{
					Left:  r.Args[0],
					Right: r.Args[1],
				},
				r.Args[2],
			},
		}.Reduce()
	case "=;":
		return Rune{
			Lit: "=/",
			Args: []Node{
				r.Args[0],
				r.Args[2],
				r.Args[1],
			},
		}.Reduce()
	case "=?":
		return Rune{
			Lit: "=.",
			Args: []Node{
				r.Args[0],
				Rune{
					Lit: "?:",
					Args: []Node{
						r.Args[1],
						r.Args[2],
						r.Args[0],
					},
				},
				r.Args[3],
			},
		}.Reduce()

	// bar
	case "|.":
		return Rune{
			Lit: "|%",
			Args: []Node{
				Buc{},
				r.Args[0],
			},
		}.Reduce()
	case "|_":
		return Rune{
			Lit: "=|",
			Args: []Node{
				r.Args[0],
				Rune{
					Lit:  "|%",
					Args: r.Args[1:],
				},
			},
		}.Reduce()
	case "|=":
		return Rune{
			Lit: "=|",
			Args: []Node{
				r.Args[0],
				Rune{
					Lit:  "|.",
					Args: []Node{r.Args[1]},
				},
			},
		}.Reduce()
	case "|^", "|-":
		return Rune{
			Lit: "=>",
			Args: []Node{
				Rune{
					Lit: "|%",
					Args: append([]Node{
						Buc{},
						r.Args[0],
					}, r.Args[1:]...),
				},
				Buc{},
			},
		}.Reduce()

	// cen
	case "%_":
		return Rune{
			Lit: "^+",
			Args: []Node{
				r.Args[0],
				Rune{
					Lit:  "%=",
					Args: r.Args[1:],
				},
			},
		}.Reduce()
	case "%~":
		return Rune{
			Lit: "=+",
			Args: []Node{
				r.Args[1],
				Rune{
					Lit: "=>",
					Args: []Node{
						Rune{
							Lit: "%=",
							Args: []Node{
								Slot{Address: "2"},
								Slot{Address: "6"},
								r.Args[2],
							},
						},
						r.Args[0],
					},
				},
			},
		}.Reduce()
	case "%-":
		return Rune{
			Lit:  "%~",
			Args: []Node{Buc{}, r.Args[0], r.Args[1]},
		}.Reduce()

	default:
		panic("unhandled reduction " + r.Lit)
	}
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
	case Tis:
		writeNode(n.Left)
		writeString("=")
		writeNode(n.Right)
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
