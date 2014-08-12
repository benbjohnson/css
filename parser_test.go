package css_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/benbjohnson/css"
)

// Ensure that a declaration can be parsed into an AST.
func TestParseDeclaration(t *testing.T) {
	var tests = []struct {
		s   string
		v   string
		err string
	}{
		{s: `foo: bar`, v: `foo: bar`},

		//{s: ``, err: `unexpected EOF`},
		//{s: ` foo bar`, err: `expected EOF, got "bar"`},
	}

	for i, tt := range tests {
		v, err := css.ParseDeclaration(css.NewTokenizer(strings.NewReader(tt.s)))
		if tt.err != "" || errstring(err) != "" {
			if tt.err != errstring(err) {
				t.Errorf("%d. <%q> error: exp=%q, got=%q", i, tt.s, tt.err, errstring(err))
			}
		} else if v == nil {
			t.Errorf("%d. <%q> expected value", i, tt.s)
		} else if print(v) != tt.v {
			t.Errorf("%d. <%q>\n\nexp: %s\n\ngot: %s", i, tt.s, tt.v, print(v))
		}
	}
}

// Ensure that component values can be parsed into the correct AST.
func TestParseComponentValue(t *testing.T) {
	var tests = []struct {
		s   string
		v   string
		err string
	}{
		{s: `foo`, v: `foo`},
		{s: `  :`, v: `:`},
		{s: `  :   `, v: `:`},
		{s: `{}`, v: `{}`},
		{s: `{foo: bar}`, v: `{foo: bar}`},
		{s: `{foo: {bar}}`, v: `{foo: {bar}}`},
		{s: ` [12.34]`, v: `[12.34]`},
		{s: ` [12.34]`, v: `[12.34]`},
		{s: ` fun(12, 34, "foo")`, v: `fun(12, 34, "foo")`},
		{s: ` fun("hello"`, v: `fun("hello")`},

		{s: ``, err: `unexpected EOF`},
		{s: ` foo bar`, err: `expected EOF, got bar`},
	}

	for i, tt := range tests {
		v, err := css.ParseComponentValue(css.NewTokenizer(strings.NewReader(tt.s)))
		if tt.err != "" || errstring(err) != "" {
			if tt.err != errstring(err) {
				t.Errorf("%d. <%q> error: exp=%q, got=%q", i, tt.s, tt.err, errstring(err))
			}
		} else if v == nil {
			t.Errorf("%d. <%q> expected value", i, tt.s)
		} else if print(v) != tt.v {
			t.Errorf("%d. <%q>\n\nexp: %s\n\ngot: %s", i, tt.s, tt.v, print(v))
		}
	}
}

// Ensure that a list of component values can be parsed into the correct AST.
func TestParseComponentValues(t *testing.T) {
	var tests = []struct {
		s   string
		v   string
		err string
	}{
		{s: `foo bar`, v: `foo bar`},
		{s: `foo func(bar) { baz }`, v: `foo func(bar) { baz }`},
	}

	for i, tt := range tests {
		v, err := css.ParseComponentValues(css.NewTokenizer(strings.NewReader(tt.s)))
		if tt.err != "" || errstring(err) != "" {
			if tt.err != errstring(err) {
				t.Errorf("%d. <%q> error: exp=%q, got=%q", i, tt.s, tt.err, errstring(err))
			}
		} else if v == nil {
			t.Errorf("%d. <%q> expected value", i, tt.s)
		} else if print(v) != tt.v {
			t.Errorf("%d. <%q>\n\nexp: %s\n\ngot: %s", i, tt.s, tt.v, print(v))
		}
	}
}

// errstring returns the string representation of the error.
func errstring(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

// print pretty prints an AST node to a string using the default configuration.
func print(n css.Node) string {
	var buf bytes.Buffer
	var p css.Printer
	_ = p.Fprint(&buf, n)
	return buf.String()
}
