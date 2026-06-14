package highlight

import (
	"strings"
	"testing"
)

// classOf returns the class of the first token whose text equals text, or a
// sentinel reporting that no such token was found.
func classOf(tokens []Token, text string) TokenClass {
	for _, t := range tokens {
		if t.Text == text {
			return t.Class
		}
	}
	return TokenClass("<not found: " + text + ">")
}

func TestTokenizeGoRoundTrip(t *testing.T) {
	srcs := []string{
		"",
		"\n\n",
		"\t  \n",
		"package main\n",
		"func main() {}\n",
		"x := 1 + 2*3\n",
		"// a comment\nvar s = \"a\\tstring\"\n",
		"/* block */ const c = 3.14i\nrune := 'x'\n",
		"package p\n\nimport \"fmt\"\n\nfunc f(a, b int) int {\n\treturn a + b // sum\n}\n",
		"this is not valid go @#$ but must round-trip\n",
		"package main\r\nfunc f() {}\r\n", // Windows CRLF line endings
		"x := `line1\r\nline2`\n",         // CR inside a raw string literal
		"x := \"unterminated",             // unterminated string literal
		"/* unterminated block comment",   // unterminated block comment
		"日本語 := 1 // 絵文字\U0001F600\n",     // multibyte identifiers and comment
	}
	for _, src := range srcs {
		var b strings.Builder
		for _, tok := range TokenizeGo(src) {
			b.WriteString(tok.Text)
		}
		if got := b.String(); got != src {
			t.Errorf("round-trip mismatch\n src: %q\n got: %q", src, got)
		}
	}
}

func TestClassify(t *testing.T) {
	src := "package main\n" +
		"// doc\n" +
		"func add(a int) error {\n" +
		"\ts := \"hi\"\n" +
		"\tn := 42\n" +
		"\t_ = make([]byte, n)\n" +
		"\tok := true\n" +
		"\treturn nil\n" +
		"}\n"
	tokens := TokenizeGo(src)

	cases := []struct {
		text string
		want TokenClass
	}{
		{"package", ClassKeyword},
		{"func", ClassKeyword},
		{"return", ClassKeyword},
		{"// doc", ClassComment}, // comment, without its trailing newline
		{"add", ClassFunction},   // declared function name
		{"make", ClassBuiltin},   // predeclared builtin
		{"int", ClassType},       // predeclared type
		{"byte", ClassType},      // predeclared type
		{"error", ClassType},     // predeclared type
		{"\"hi\"", ClassString},  // string literal
		{"42", ClassNumber},      // int literal
		{"true", ClassConstant},  // predeclared constant
		{"nil", ClassConstant},   // predeclared zero value
		{"a", ClassIdent},        // plain identifier (parameter)
	}
	for _, c := range cases {
		if got := classOf(tokens, c.text); got != c.want {
			t.Errorf("classOf(%q) = %q, want %q", c.text, got, c.want)
		}
	}
}

func TestFunctionCallClassified(t *testing.T) {
	// fmt is a plain identifier, Println is a call.
	tokens := TokenizeGo("fmt.Println(x)\n")
	if got := classOf(tokens, "fmt"); got != ClassIdent {
		t.Errorf("fmt class = %q, want %q", got, ClassIdent)
	}
	if got := classOf(tokens, "Println"); got != ClassFunction {
		t.Errorf("Println class = %q, want %q", got, ClassFunction)
	}
}

func TestTokenizeGoEmpty(t *testing.T) {
	if got := TokenizeGo(""); got != nil {
		t.Errorf("TokenizeGo(\"\") = %v, want nil", got)
	}
}
