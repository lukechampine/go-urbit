package mach

import (
	"testing"

	"lukechampine.com/go-urbit/hoon/ast"
	"lukechampine.com/go-urbit/hoon/parser"
	"lukechampine.com/go-urbit/hoon/scanner"
)

func parse(s string) ast.Node {
	return parser.New(scanner.New([]byte(s))).Parse()
}

func Test(t *testing.T) {
	var tests = []struct {
		hoon string
		llvm string
	}{
		{
			hoon: `1`,
			llvm: `
define i32 @main() {
0:
	ret i32 1
}
`[1:],
		},
		{
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

define private i32 @0(i32 %n, i32 %acc) {
0:
	%1 = icmp eq i32 %n, 6
	br i1 %1, label %2, label %3

2:
	ret i32 %acc

3:
	%4 = add i32 %n, 1
	%5 = mul i32 %acc, %n
	%6 = call i32 @0(i32 %4, i32 %5)
	ret i32 %6
}
`[1:],
		},
		{
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
	}
	for _, test := range tests {
		got := Transpile(parse(test.hoon))
		if got != test.llvm {
			t.Errorf("bad transpile:\nhoon: %s\nllvm: %q\ngot:  %s", test.hoon, test.llvm, got)
		}
	}
}
