package token

import (
	"strconv"
)

type Token int

const (
	ILLEGAL Token = iota
	EOF
	Comment
	Ace
	Bar
	Bas
	Buc
	Cab
	Cen
	Col
	Com
	Doq
	Dot
	Fas
	Gal
	Gap
	Gar
	Hax
	Hep
	Kel
	Ker
	Ket
	Lus
	Mic
	Pal
	Pam
	Par
	Pat
	Sel
	Ser
	Sig
	Soq
	Tar
	Tic
	Tis
	Wut
	Zap
	HepHep
	TisTis
	Rune
	Face
	Num
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	Comment: "COMMENT",

	Ace: "ACE",
	Gap: "GAP",
	Bar: "|",
	Bas: `\`,
	Buc: "$",
	Cab: "_",
	Cen: "%",
	Col: ":",
	Com: ",",
	Doq: `"`,
	Dot: ".",
	Fas: "/",
	Gal: "<",
	Gar: ">",
	Hax: "#",
	Hep: "-",
	Kel: "{",
	Ker: "}",
	Ket: "^",
	Lus: "+",
	Mic: ";",
	Pal: "(",
	Pam: "&",
	Par: ")",
	Pat: "@",
	Sel: "[",
	Ser: "]",
	Sig: "~",
	Soq: "'",
	Tar: "*",
	Tic: "`",
	Tis: "=",
	Wut: "?",
	Zap: "!",

	HepHep: "--",
	TisTis: "==",
	Rune:   "RUNE",
	Face:   "FACE",
	Num:    "NUM",
}

func (t Token) String() string {
	s := ""
	if 0 <= t && t < Token(len(tokens)) {
		s = tokens[t]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(t)) + ")"
	}
	return s
}
