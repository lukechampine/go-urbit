package parser

import (
	"fmt"

	"lukechampine.com/urbit/hoon/ast"
	"lukechampine.com/urbit/hoon/scanner"
	"lukechampine.com/urbit/hoon/token"
)

// TODO: determine appropriate values for these
var precedences = map[token.Token]int{
	token.Col: 2,
	token.Tis: 2,
}

type runeEntry struct {
	args    int
	jogging bool
	post    int // some runes have fixed args *after* the jog
}

// TODO: probably numerous errors here
var runeTab = map[string]runeEntry{
	".^": {args: 2},
	".+": {args: 1},
	".*": {args: 2},
	".=": {args: 2},
	".?": {args: 1},
	"!>": {args: 1},
	"!<": {args: 2},
	"!:": {args: 1},
	"!.": {args: 1},
	"!=": {args: 1},
	"!?": {args: 2},
	"=+": {args: 2},
	"=-": {args: 2},
	"=|": {args: 2},
	"=/": {args: 3},
	"=;": {args: 2},
	"=.": {args: 3},
	"=:": {jogging: true, post: 2},
	"=?": {args: 4},
	"=*": {args: 3},
	"=>": {args: 2},
	"=<": {args: 2},
	"=~": {args: 2},
	"=,": {args: 2},
	"=^": {args: 4},
	"?>": {args: 2},
	"?<": {args: 2},
	"?|": {args: 1, jogging: true},
	"?&": {args: 1, jogging: true},
	"?!": {args: 1},
	"?=": {args: 2},
	"?:": {args: 3},
	"?.": {args: 3},
	"?@": {args: 3},
	"?^": {args: 3},
	"?~": {args: 3},
	"?-": {args: 1, jogging: true},
	"?+": {args: 1, jogging: true},
	"|_": {args: 1, jogging: true},
	"|%": {args: 0, jogging: true}, // cores can be empty
	"|:": {args: 2},
	"|.": {args: 1},
	"|-": {args: 1},
	"|?": {args: 1},
	"|^": {args: 2, jogging: true},
	"|~": {args: 2},
	"|=": {args: 2},
	"|*": {args: 2},
	"|@": {args: 2},
	"++": {args: 2},
	"+$": {args: 2},
	"+*": {args: 2},
	"+|": {args: 1},
	":-": {args: 2},
	":_": {args: 2},
	":+": {args: 3},
	":^": {args: 4},
	":*": {args: 1, jogging: true},
	":~": {args: 1, jogging: true},
	"::": {args: 1},
	"%~": {args: 3, jogging: true},
	"%-": {args: 2},
	"%.": {args: 2},
	"%+": {args: 3},
	"%^": {args: 4},
	"%:": {args: 2, jogging: true},
	"%=": {args: 3, jogging: true},
	"%_": {args: 2, jogging: true},
	"%*": {args: 3, jogging: true},
	"^|": {args: 1},
	"^&": {args: 1},
	"^?": {args: 1},
	"^:": {args: 1},
	"^.": {args: 2},
	"^-": {args: 2},
	"^+": {args: 2},
	"^~": {args: 1},
	"^*": {args: 1},
	"^=": {args: 2},
	"$_": {args: 1},
	"$%": {args: 1, jogging: true},
	"$:": {args: 1, jogging: true},
	"$?": {args: 1, jogging: true},
	"$<": {args: 2},
	"$>": {args: 2},
	"$-": {args: 2},
	"$@": {args: 2},
	"$^": {args: 2},
	"$~": {args: 2},
	"$=": {args: 2},
	";:": {args: 2, jogging: true},
	";+": {args: 1},
	";/": {args: 1},
	";*": {args: 1},
	";=": {args: 1, jogging: true},
	";;": {args: 2},
	";~": {args: 2},
	"~>": {args: 2},
	"~|": {args: 2},
	"~_": {args: 2},
	"~$": {args: 2},
	"~%": {args: 3, jogging: true},
	"~<": {args: 2},
	"~+": {args: 1},
	"~/": {args: 2},
	"~&": {args: 2},
	"~?": {args: 3},
	"~!": {args: 2},
}

func hasArms(r string) bool {
	switch r {
	case "|%", "|^": // TODO
		return true
	default:
		return false
	}
}

type Parser struct {
	s   *scanner.Scanner
	tok token.Token
	lit string
}

func (p *Parser) next() {
	p.tok, p.lit = p.s.Scan()
}

func (p *Parser) consumeWhitespace() {
	for p.tok == token.Ace || p.tok == token.Gap {
		p.next()
	}
}

func (p *Parser) expect(t token.Token) {
	if p.tok != t {
		panic(fmt.Sprintf("parse: expected %q, got %q", t, p.tok))
	}
	p.next()
}

func (p *Parser) consumeWide() []ast.Node {
	var nodes []ast.Node
	for {
		nodes = append(nodes, p.parseExpr())
		if p.tok == token.Par {
			p.next()
			return nodes
		}
		p.expect(token.Ace)
	}
}

func (p *Parser) consumeWideComma() []ast.Node {
	var nodes []ast.Node
	for {
		nodes = append(nodes, p.parseExpr())
		p.expect(token.Ace)
		nodes = append(nodes, p.parseExpr())
		if p.tok == token.Par {
			p.next()
			return nodes
		}
		p.expect(token.Com)
		p.expect(token.Ace)
	}
}

func (p *Parser) consumeTall(stop token.Token) []ast.Node {
	var nodes []ast.Node
	for {
		p.expect(token.Gap)
		if p.tok == stop {
			p.next()
			return nodes
		}
		nodes = append(nodes, p.parseExpr())
	}
}

func (p *Parser) Parse() ast.Node {
	p.consumeWhitespace()
	n := p.parseExpr()
	p.consumeWhitespace()
	return n
}

func (p *Parser) parseExpr() ast.Node {
	return p.parseBinaryExpr(1) // lowest precedence
}

func (p *Parser) parseBinaryExpr(prec int) ast.Node {
	n := p.parseUnaryExpr()
	for {
		op := p.tok
		qPrec, ok := precedences[op]
		if !ok || qPrec < prec {
			break
		}
		p.next()
		next := p.parseBinaryExpr(qPrec + 1)
		switch op {
		case token.Col:
			n = ast.Rune{
				Tok:  op,
				Lit:  "=<",
				Args: []ast.Node{n, next},
			}
		case token.Tis:
			n = ast.Tis{
				Tok:   op,
				Left:  n,
				Right: next,
			}
		default:
			panic("unhandled binop")
		}
	}
	return n
}

func (p *Parser) parseUnaryExpr() ast.Node {
	t, lit := p.tok, p.lit
	p.next()
	switch t {
	case token.Face:
		if p.tok == token.Pal {
			p.next()
			return ast.Rune{
				Tok: t,
				Lit: "%=",
				Args: append([]ast.Node{ast.Face{
					Tok:  t,
					Name: lit,
				}}, p.consumeWideComma()...),
			}
		}
		return ast.Face{Tok: t, Name: lit}
	case token.Num:
		return ast.Num{Tok: t, Int: lit}
	case token.Pat:
		return ast.Pat{Tok: t}
	case token.Dot:
		return ast.Dot{Tok: t}
	case token.Buc:
		if p.tok == token.Pal {
			p.next()
			return ast.Rune{
				Tok: t,
				Lit: "%=",
				Args: append([]ast.Node{ast.Buc{
					Tok: t,
				}}, p.consumeWideComma()...),
			}
		}
		return ast.Buc{Tok: t}
	case token.Rune:
		return p.parseRune(t, lit)
	case token.Lus:
		p.expect(token.Pal)
		q := p.parseExpr()
		p.expect(token.Par)
		return ast.Rune{
			Tok:  t,
			Lit:  ".+",
			Args: []ast.Node{q},
		}
	case token.Tis:
		p.expect(token.Pal)
		q := p.parseExpr()
		p.expect(token.Ace)
		r := p.parseExpr()
		p.expect(token.Par)
		return ast.Rune{
			Tok:  t,
			Lit:  ".=",
			Args: []ast.Node{q, r},
		}
	case token.Pal:
		return ast.Rune{
			Tok:  t,
			Lit:  "%-",
			Args: p.consumeWide(),
		}
	case token.Sel:
		n := p.parseCons()
		p.expect(token.Ser)
		return n
	default:
		panic(fmt.Sprintf("unhandled token %v (%v)", lit, t))
	}
}

func (p *Parser) parseRune(tok token.Token, lit string) ast.Node {
	e, ok := runeTab[lit]
	if !ok {
		panic("unhandled rune")
	}
	n := ast.Rune{
		Tok: tok,
		Lit: lit,
	}
	if p.tok == token.Pal {
		p.next()
		n.Args = p.consumeWide()
	} else {
		n.Args = make([]ast.Node, e.args)
		for i := range n.Args {
			p.expect(token.Gap)
			n.Args[i] = p.parseExpr()
		}
	}
	if hasArms(lit) {
		n.Args = append(n.Args, p.consumeTall(token.HepHep)...)
	} else if e.jogging {
		n.Args = append(n.Args, p.consumeTall(token.TisTis)...)
	}
	return n
}

func (p *Parser) parseCons() ast.Node {
	e := p.parseExpr()
	if p.tok == token.Ser {
		return e
	}
	p.expect(token.Ace)
	return ast.Cell{
		Tok:  p.tok,
		Head: e,
		Tail: p.parseCons(),
	}
}

func New(s *scanner.Scanner) *Parser {
	p := &Parser{
		s: s,
	}
	p.next()
	return p
}
