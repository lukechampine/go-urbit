package scanner

import (
	"strings"

	"lukechampine.com/go-urbit/hoon/token"
)

type Scanner struct {
	src []byte
	off int
	ch  rune
}

func (s *Scanner) next() {
	s.off++
	if s.off >= len(s.src) {
		s.off = len(s.src)
		s.ch = -1
		return
	}
	s.ch = rune(s.src[s.off])
}

// TODO: this is currently unused; is it necessary?
func (s *Scanner) peek() rune {
	if s.off+1 < len(s.src) {
		return rune(s.src[s.off+1])
	}
	return -1
}

func (s *Scanner) scanWhitespace() (token.Token, string) {
	var sb strings.Builder
	for s.ch == ' ' || s.ch == '\n' {
		sb.WriteRune(s.ch)
		s.next()
	}
	lit := sb.String()
	if lit == " " {
		return token.Ace, lit
	}
	return token.Gap, lit
}

func (s *Scanner) scanComment() (token.Token, string) {
	var sb strings.Builder
	sb.WriteString("::")
	for s.ch != '\n' && s.ch != -1 {
		sb.WriteRune(s.ch)
		s.next()
	}
	lit := sb.String()
	return token.Comment, lit
}

func isKebab(c rune) bool {
	return c == '-' || ('a' <= c && c <= 'z')
}

func isNumber(c rune) bool {
	return c == '.' || ('0' <= c && c <= '9')
}

func (s *Scanner) scanFace() (token.Token, string) {
	var sb strings.Builder
	for isKebab(s.ch) {
		sb.WriteRune(s.ch)
		s.next()
	}
	return token.Face, sb.String()
}

func (s *Scanner) scanNumber() (token.Token, string) {
	// TODO: this only parses very simple numbers
	var sb strings.Builder
	for isNumber(s.ch) {
		sb.WriteRune(s.ch)
		s.next()
	}
	return token.Num, sb.String()
}

var runeTab = func() map[int32]string {
	m := make(map[int32]string)
	for _, r := range strings.Fields(`
	.^ .+ .* .= .?
	!> !< !: !. != !? !!
	=+ =- =| =/ =; =. =: =? =* => =< =~ =, =^
	?> ?< ?| ?& ?! ?= ?: ?. ?@ ?^ ?~ ?- ?+
	|_ |% |: |. |- |? |^ |~ |= |* |@
	++ +$ +* +|
	:- :_ :+ :^ :* :~
	%~ %- %. %+ %^ %: %= %_ %*
	^| ^& ^? ^: ^. ^- ^+ ^~ ^* ^=
	$_ $% $: $? $< $> $- $@ $^ $~ $=
	;: ;+ ;/ ;* ;= ;; ;~
	~> ~| ~_ ~$ ~% ~< ~+ ~/ ~& ~? ~! ~=
	`) {
		key := (int32(r[0]) << 16) | int32(r[1])
		m[key] = r
	}
	return m
}()

func isComment(c, n rune) bool {
	return c == ':' && n == ':'
}

func isRune(c, n rune) (token.Token, string) {
	if lit := runeTab[(c<<16)|n]; lit != "" {
		return token.Rune, lit
	}
	return 0, ""
}

func isTerminator(c, n rune) (token.Token, string) {
	if c == n {
		if c == '-' {
			return token.HepHep, "--"
		} else if c == '=' {
			return token.TisTis, "=="
		}
	}
	return 0, ""
}

var singleCharTokenTab = map[rune]token.Token{
	'|': token.Bar, '\\': token.Bas, '$': token.Buc,
	'_': token.Cab, '%': token.Cen, ':': token.Col,
	',': token.Com, '"': token.Doq, '.': token.Dot,
	'/': token.Fas, '<': token.Gal, '>': token.Gar,
	'#': token.Hax, '-': token.Hep, '{': token.Kel,
	'}': token.Ker, '^': token.Ket, '+': token.Lus,
	';': token.Mic, '(': token.Pal, '&': token.Pam,
	')': token.Par, '@': token.Pat, '[': token.Sel,
	']': token.Ser, '~': token.Sig, '\'': token.Soq,
	'*': token.Tar, '`': token.Tic, '=': token.Tis,
	'?': token.Wut, '!': token.Zap,
}

func isSingleCharToken(c rune) (token.Token, string) {
	if tok, ok := singleCharTokenTab[c]; ok {
		return tok, tok.String()
	}
	return 0, ""
}

func (s *Scanner) Scan() (token.Token, string) {
	switch c := s.ch; {
	case c == ' ' || c == '\n':
		return s.scanWhitespace()
	case 'a' <= c && c <= 'z':
		return s.scanFace()
	case '0' <= c && c <= '9':
		return s.scanNumber()
	case c == -1:
		return token.EOF, ""
	default:
		s.next() // always make progress
		if isComment(c, s.ch) {
			s.next()
			return s.scanComment()
		}
		if tok, lit := isRune(c, s.ch); lit != "" {
			s.next()
			return tok, lit
		}
		if tok, lit := isTerminator(c, s.ch); lit != "" {
			s.next()
			return tok, lit
		}
		if tok, lit := isSingleCharToken(c); lit != "" {
			return tok, lit
		}
		return token.ILLEGAL, string(c)
	}
}

func New(src []byte) *Scanner {
	return &Scanner{
		src: src,
		ch:  rune(src[0]),
	}
}
