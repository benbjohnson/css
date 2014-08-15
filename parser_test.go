package css_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/benbjohnson/css"
)

// Ensure that a list of rules can be parsed into an AST.
func TestParseRules(t *testing.T) {
	var tests = []ParserTest{
		{in: `foo { padding: 10px; }`, out: `foo { padding: 10px; }`},
		{in: `@import url(/css/screen.css) screen, projection;`, out: `@import url(/css/screen.css) screen, projection;`},
		{in: `@xxx; foo { padding: 10 0; }`, out: `@xxx; foo { padding: 10 0; }`},
		{in: `<!-- comment --> foo { }`, out: `<!-- comment --> foo { }`},
	}

	for _, tt := range tests {
		var p css.Parser
		v := p.ParseRules(css.NewScanner(strings.NewReader(tt.in)))
		tt.Assert(t, v, p.Errors)
	}
}

// Ensure that a rule can be parsed into an AST.
func TestParseRule(t *testing.T) {
	var tests = []ParserTest{
		{in: `foo { padding: 10px; }`, out: `foo { padding: 10px; }`},
		{in: `foo { padding: 10px; `, out: `foo { padding: 10px; }`},
		{in: `  #foo bar, .baz bat {}  `, out: `#foo bar, .baz bat {}`},
		{in: `@media (max-width: 600px) { .nav { display: none; }}`, out: `@media (max-width: 600px) { .nav { display: none; }}`},

		{in: ``, err: `unexpected EOF`},
		{in: `  `, err: `unexpected EOF`},
		{in: `foo {} bar`, err: `expected EOF, got bar`},
	}

	for _, tt := range tests {
		var p css.Parser
		v := p.ParseRule(css.NewScanner(strings.NewReader(tt.in)))
		tt.Assert(t, v, p.Errors)
	}
}

// Ensure that a declaration can be parsed into an AST.
func TestParseDeclaration(t *testing.T) {
	var tests = []ParserTest{
		{in: `foo: bar`, out: `foo: bar`},

		{in: ``, err: `expected ident, got EOF`},
		{in: ` foo bar`, err: `expected colon, got bar`},
	}

	for _, tt := range tests {
		var p css.Parser
		v := p.ParseDeclaration(css.NewScanner(strings.NewReader(tt.in)))
		tt.Assert(t, v, p.Errors)
	}
}

// Ensure that a list of declarations can be parsed into an AST.
func TestParseDeclarations(t *testing.T) {
	var tests = []ParserTest{
		{in: `foo: bar`, out: `foo: bar;`},
		{in: `font-size: 20px; font-weight:bold`, out: `font-size: 20px; font-weight:bold;`},
	}

	for _, tt := range tests {
		var p css.Parser
		v := p.ParseDeclarations(css.NewScanner(strings.NewReader(tt.in)))
		tt.Assert(t, v, p.Errors)
	}
}

// Ensure that component values can be parsed into the correct AST.
func TestParseComponentValue(t *testing.T) {
	var tests = []ParserTest{
		{in: `foo`, out: `foo`},
		{in: `  :`, out: `:`},
		{in: `  :   `, out: `:`},
		{in: `{}`, out: `{}`},
		{in: `{foo: bar}`, out: `{foo: bar}`},
		{in: `{foo: {bar}}`, out: `{foo: {bar}}`},
		{in: ` [12.34]`, out: `[12.34]`},
		{in: ` [12.34]`, out: `[12.34]`},
		{in: ` fun(12, 34, "foo")`, out: `fun(12, 34, "foo")`},
		{in: ` fun("hello"`, out: `fun("hello")`},

		{in: ``, err: `unexpected EOF`},
		{in: ` foo bar`, err: `expected EOF, got bar`},
	}

	for _, tt := range tests {
		var p css.Parser
		v := p.ParseComponentValue(css.NewScanner(strings.NewReader(tt.in)))
		tt.Assert(t, v, p.Errors)
	}
}

// Ensure that a list of component values can be parsed into the correct AST.
func TestParseComponentValues(t *testing.T) {
	var tests = []ParserTest{
		{in: `foo bar`, out: `foo bar`},
		{in: `foo func(bar) { baz }`, out: `foo func(bar) { baz }`},
	}

	for _, tt := range tests {
		var p css.Parser
		v := p.ParseComponentValues(css.NewScanner(strings.NewReader(tt.in)))
		tt.Assert(t, v, p.Errors)
	}
}

// ParserTest represents a generic framework for table tests against the parser.
type ParserTest struct {
	in  string // input CSS
	out string // matches against generated CSS
	err string // stringified error, empty string if no error.
}

// Assert validates the node against the output CSS and checks for errors.
func (tt *ParserTest) Assert(t *testing.T, n css.Node, errors css.ErrorList) {
	var errstring string
	if len(errors) > 0 {
		errstring = errors.Error()
	}

	if tt.err != "" || errstring != "" {
		if tt.err != errstring {
			t.Errorf("<%q> error: exp=%q, got=%q", tt.in, tt.err, errstring)
		}
	} else if n == nil {
		t.Errorf("<%q> expected value", tt.in)
	} else if print(n) != tt.out {
		t.Errorf("<%q>\n\nexp: %s\n\ngot: %s", tt.in, tt.out, print(n))
	}
}

// print pretty prints an AST node to a string using the default configuration.
func print(n css.Node) string {
	var buf bytes.Buffer
	var p css.Printer
	_ = p.Fprint(&buf, n)
	return buf.String()
}
