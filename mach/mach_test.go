package mach

import (
	"testing"

	"lukechampine.com/urbit/hoon/ast"
	"lukechampine.com/urbit/hoon/parser"
	"lukechampine.com/urbit/hoon/scanner"
)

func parse(s string) ast.Node {
	return parser.New(scanner.New([]byte(s))).Parse()
}

func Test(t *testing.T) {
	var tests = []struct {
		desc string
		hoon string
		llvm string
	}{
		{
			desc: "atom literal",
			hoon: `1`,
			llvm: `
define i32 @main() {
0:
	ret i32 1
}
`[1:],
		},
		{
			desc: "resolve face in subject",
			hoon: `
=/  n  1
n
`,
			llvm: `
define i32 @main() {
0:
	ret i32 1
}
`[1:],
		},
		{
			desc: "add two faces",
			hoon: `
=/  a  2
=/  b  a
(add a b)
`,
			llvm: `
define i32 @main() {
0:
	%1 = add i32 2, 2
	ret i32 %1
}
`[1:],
		},
		{
			desc: "change subject, simple conditional",
			hoon: `
=/  a  2
=/  b  7
=.  a  8
?:  =(a +(b))
  a
b
`,
			llvm: `
define i32 @main() {
0:
	%1 = add i32 7, 1
	%2 = icmp eq i32 8, %1
	br i1 %2, label %3, label %4

3:
	ret i32 8

4:
	ret i32 7
}
`[1:],
		},
		{
			desc: "barhep recursion",
			hoon: `
=/  n    1
=/  acc  1
|-
?:  =(n 6)
  acc
$(acc (mul acc n), n +(n))
`,
			llvm: `
define i32 @main() {
0:
	%1 = call i32 @0(i32 1, i32 1)
	ret i32 %1
}

define private i32 @0(i32 %acc, i32 %n) {
0:
	%1 = icmp eq i32 %n, 6
	br i1 %1, label %2, label %3

2:
	ret i32 %acc

3:
	%4 = mul i32 %acc, %n
	%5 = add i32 %n, 1
	%6 = call i32 @0(i32 %4, i32 %5)
	ret i32 %6
}
`[1:],
		},
		{
			desc: "non-tail-recursive barhep",
			hoon: `
=/  n  5
|-
?:  =(n 0)
  1
(mul n $(n (dec n)))
`,
			llvm: `
define i32 @main() {
0:
	%1 = call i32 @0(i32 5)
	ret i32 %1
}

define private i32 @0(i32 %n) {
0:
	%1 = icmp eq i32 %n, 0
	br i1 %1, label %2, label %3

2:
	ret i32 1

3:
	%4 = sub i32 %n, 1
	%5 = call i32 @0(i32 %4)
	%6 = mul i32 %n, %5
	ret i32 %6
}
`[1:],
		},
		{
			desc: "recurse without modifying entire subject",
			hoon: `
=/  a  5
=/  b  0
|-
?:  =(a +(b))  b
$(b +(b))
`,
			llvm: `
define i32 @main() {
0:
	%1 = call i32 @0(i32 5, i32 0)
	ret i32 %1
}

define private i32 @0(i32 %a, i32 %b) {
0:
	%1 = add i32 %b, 1
	%2 = icmp eq i32 %a, %1
	br i1 %2, label %3, label %4

3:
	ret i32 %b

4:
	%5 = add i32 %b, 1
	%6 = call i32 @0(i32 %a, i32 %5)
	ret i32 %6
}
`[1:],
		},
		{
			desc: "assign gate to face",
			hoon: `
=/  f
  |=  a=@
  =/  g
    |=  b=@
    (dec b)
  (g a)
(f 5)
`,
			llvm: `
define i32 @main() {
0:
	%1 = call i32 @0(i32 5)
	ret i32 %1
}

define private i32 @0(i32 %a) {
0:
	%1 = call i32 @1(i32 %a, i32 %a)
	ret i32 %1
}

define private i32 @1(i32 %a, i32 %b) {
0:
	%1 = sub i32 %b, 1
	ret i32 %1
}
`[1:],
		},
	}
	for _, test := range tests {
		got := Transpile(parse(test.hoon))
		if got != test.llvm {
			t.Errorf("bad transpile:\nhoon: %s\nllvm: %q\ngot:  %s", test.hoon, test.llvm, got)
		}
	}
}
