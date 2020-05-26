package scanner

import (
	"reflect"
	"testing"

	"lukechampine.com/go-urbit/hoon/token"
	. "lukechampine.com/go-urbit/hoon/token"
)

func TestScan(t *testing.T) {
	tests := []struct {
		hoon string
		exp  []Token
	}{
		{
			hoon: `=(a +(b))`,
			exp:  []Token{Tis, Pal, Face, Ace, Lus, Pal, Face, Par, Par},
		},
		{
			hoon: `=/  n  1
                  [. .]:n  :: dup`,
			exp: []Token{Rune, Gap, Face, Gap, Num, Gap, Sel, Dot, Ace, Dot, Ser, Col, Face, Gap, Comment},
		},
		{
			hoon: `|=  n=@
                   =/  acc=@  1
                   |-
                   ?:  =(n 0)  acc
                   %=  $
                     n  (dec n)
                     acc  (mul acc n)
                   ==`,
			exp: []Token{
				Rune, Gap, Face, Tis, Pat, Gap,
				Rune, Gap, Face, Tis, Pat, Gap, Num, Gap,
				Rune, Gap,
				Rune, Gap, Tis, Pal, Face, Ace, Num, Par, Gap, Face, Gap,
				Rune, Gap, Buc, Gap,
				Face, Gap, Pal, Face, Ace, Face, Par, Gap,
				Face, Gap, Pal, Face, Ace, Face, Ace, Face, Par, Gap,
				TisTis,
			},
		},
	}
	for _, test := range tests {
		var ts []Token
		for s := New([]byte(test.hoon)); ; {
			tok, _ := s.Scan()
			if tok == token.EOF {
				break
			}
			ts = append(ts, tok)
		}
		if !reflect.DeepEqual(ts, test.exp) {
			t.Fatalf("bad scan:\nexp: %v\ngot: %v", test.exp, ts)
		}
	}
}
