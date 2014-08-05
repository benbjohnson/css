package scanner_test

import (
	"bytes"
	"flag"
	"reflect"
	"testing"

	"github.com/benbjohnson/css/scanner"
)

// testiter sets the table test iteration to run in isolation.
var testiter = flag.Int("test.iter", -1, "table test number")

func init() {
	flag.Parse()
}

// Ensure than the scanner returns appropriate tokens and literals.
func TestScanner_Scan(t *testing.T) {
	var tests = []struct {
		s   string
		tok *css.Token
		err string
	}{
		{s: ``, tok: &css.Token{Type: css.EOF}},
		{s: `   `, tok: &css.Token{Type: css.WHITESPACE, Value: `   `}},

		{s: `""`, tok: &css.Token{Type: css.STRING, Value: ``, Ending: '"'}},
		{s: `"`, tok: &css.Token{Type: css.STRING, Value: ``, Ending: '"'}},
		{s: `"foo`, tok: &css.Token{Type: css.STRING, Value: `foo`, Ending: '"'}},
		{s: `"hello world"`, tok: &css.Token{Type: css.STRING, Value: `hello world`, Ending: '"'}},
		{s: `'hello world'`, tok: &css.Token{Type: css.STRING, Value: `hello world`, Ending: '\''}},
		{s: "'foo\\\nbar'", tok: &css.Token{Type: css.STRING, Value: "foo\nbar", Ending: '\''}},
		{s: `'foo\ bar'`, tok: &css.Token{Type: css.STRING, Value: `foo bar`, Ending: '\''}},
		{s: `'foo\\bar'`, tok: &css.Token{Type: css.STRING, Value: `foo\bar`, Ending: '\''}},
		{s: `'frosty the \2603'`, tok: &css.Token{Type: css.STRING, Value: `frosty the ☃`, Ending: '\''}},

		{s: `0`, tok: &css.Token{Type: css.NUMBER, Flag: "integer", Value: `0`, Number: 0.0}},
		{s: `1.0`, tok: &css.Token{Type: css.NUMBER, Flag: "number", Value: `1.0`, Number: 1.0}},
		{s: `1.123`, tok: &css.Token{Type: css.NUMBER, Flag: "number", Value: `1.123`, Number: 1.123}},
		{s: `.001`, tok: &css.Token{Type: css.NUMBER, Flag: "number", Value: `.001`, Number: 0.001}},
		{s: `-.001`, tok: &css.Token{Type: css.NUMBER, Flag: "number", Value: `-.001`, Number: -0.001}},
		{s: `10000`, tok: &css.Token{Type: css.NUMBER, Flag: "integer", Value: `10000`, Number: 10000}},
		{s: `10000.`, tok: &css.Token{Type: css.NUMBER, Flag: "integer", Value: `10000`, Number: 10000}},
		{s: `100E`, tok: &css.Token{Type: css.DIMENSION, Flag: "integer", Value: `100E`, Number: 100, Unit: "E"}},
		{s: `100E+`, tok: &css.Token{Type: css.DIMENSION, Flag: "integer", Value: `100E`, Number: 100, Unit: "E"}},
		{s: `100E-`, tok: &css.Token{Type: css.DIMENSION, Flag: "integer", Value: `100E-`, Number: 100, Unit: "E-"}},
		{s: `1E2`, tok: &css.Token{Type: css.NUMBER, Flag: "number", Value: `1E2`, Number: 100}},
		{s: `1.5E2`, tok: &css.Token{Type: css.NUMBER, Flag: "number", Value: `1.5E2`, Number: 150}},
		{s: `1.5E+2`, tok: &css.Token{Type: css.NUMBER, Flag: "number", Value: `1.5E+2`, Number: 150}},
		{s: `1.5E-2`, tok: &css.Token{Type: css.NUMBER, Flag: "number", Value: `1.5E-2`, Number: 0.015}},
		{s: `+100`, tok: &css.Token{Type: css.NUMBER, Flag: "integer", Value: `+100`, Number: 100}},
		{s: `+1.0`, tok: &css.Token{Type: css.NUMBER, Flag: "number", Value: `+1.0`, Number: 1}},
		{s: `-100`, tok: &css.Token{Type: css.NUMBER, Flag: "integer", Value: `-100`, Number: -100}},
		{s: `-1.0`, tok: &css.Token{Type: css.NUMBER, Flag: "number", Value: `-1.0`, Number: -1}},
		{s: `-`, tok: &css.Token{Type: css.DELIM, Value: `-`}},

		{s: `url`, tok: &css.Token{Type: css.IDENT, Value: `url`}},
		{s: `myIdent`, tok: &css.Token{Type: css.IDENT, Value: `myIdent`}},
		{s: `my\2603`, tok: &css.Token{Type: css.IDENT, Value: `my☃`}},

		{s: `url(`, tok: &css.Token{Type: css.URL, Value: ``}},
		{s: `url(foo`, tok: &css.Token{Type: css.URL, Value: `foo`}},
		{s: `url(http://foo.com#bar?baz=bat)`, tok: &css.Token{Type: css.URL, Value: `http://foo.com#bar?baz=bat`}},
		{s: `url(  foo`, tok: &css.Token{Type: css.URL, Value: `foo`}},
		{s: `url(  foo  `, tok: &css.Token{Type: css.URL, Value: `foo`}},
		{s: `url(  \2603  `, tok: &css.Token{Type: css.URL, Value: `☃`}},
		{s: `url(foo)`, tok: &css.Token{Type: css.URL, Value: `foo`}},
		{s: `url("http://foo.com#bar?baz=bat")`, tok: &css.Token{Type: css.URL, Value: `http://foo.com#bar?baz=bat`}},
		{s: `url(  "foo"  `, tok: &css.Token{Type: css.URL, Value: `foo`}},
		{s: `url("foo"  `, tok: &css.Token{Type: css.URL, Value: `foo`}},
		{s: `url("foo")`, tok: &css.Token{Type: css.URL, Value: `foo`}},
		{s: `url("foo"x`, tok: &css.Token{Type: css.BADURL, Value: ``}},
		{s: `url("foo" x`, tok: &css.Token{Type: css.BADURL, Value: ``}},
		{s: `url(foo"`, tok: &css.Token{Type: css.BADURL, Value: ``}, err: `invalid url code point: " (U+0022)`},
		{s: `url(foo'`, tok: &css.Token{Type: css.BADURL, Value: ``}, err: `invalid url code point: ' (U+0027)`},
		{s: `url(foo(`, tok: &css.Token{Type: css.BADURL, Value: ``}, err: `invalid url code point: ( (U+0028)`},
		{s: "url(foo\001", tok: &css.Token{Type: css.BADURL, Value: ``}, err: "invalid url code point: \001 (U+0001)"},
		{s: "url(foo\\\n", tok: &css.Token{Type: css.BADURL, Value: ``}, err: `unescaped \ in url`},

		{s: `myFunc(`, tok: &css.Token{Type: css.FUNCTION, Value: `myFunc`}},

		{s: "u+A", tok: &css.Token{Type: css.UNICODERANGE, Start: 10, End: 10}},
		{s: "u+00000A", tok: &css.Token{Type: css.UNICODERANGE, Start: 10, End: 10}},
		{s: "u+000000A", tok: &css.Token{Type: css.UNICODERANGE, Start: 0, End: 0}},
		{s: "u+1?", tok: &css.Token{Type: css.UNICODERANGE, Start: 16, End: 31}},
		{s: "u+1?F", tok: &css.Token{Type: css.UNICODERANGE, Start: 16, End: 31}},
		{s: "u+02-04", tok: &css.Token{Type: css.UNICODERANGE, Start: 2, End: 4}},
		{s: "u+02-04?", tok: &css.Token{Type: css.UNICODERANGE, Start: 2, End: 4}},
		{s: "u+02-0000004", tok: &css.Token{Type: css.UNICODERANGE, Start: 2, End: 0}},

		{s: `100em`, tok: &css.Token{Type: css.DIMENSION, Flag: "integer", Value: `100em`, Number: 100, Unit: "em"}},
		{s: `-1.2in`, tok: &css.Token{Type: css.DIMENSION, Flag: "number", Value: `-1.2in`, Number: -1.2, Unit: "in"}},

		{s: `100%`, tok: &css.Token{Type: css.PERCENTAGE, Flag: "integer", Value: `100%`, Number: 100}},
		{s: `-0.2%`, tok: &css.Token{Type: css.PERCENTAGE, Flag: "number", Value: `-0.2%`, Number: -0.2}},

		{s: `#foo`, tok: &css.Token{Type: css.HASH, Value: `foo`, Flag: "id"}},
		{s: `#foo\2603 bar`, tok: &css.Token{Type: css.HASH, Value: `foo☃bar`, Flag: "id"}},
		{s: `#-x`, tok: &css.Token{Type: css.HASH, Value: `-x`, Flag: "id"}},
		{s: `#_x`, tok: &css.Token{Type: css.HASH, Value: `_x`, Flag: "id"}},
		{s: `#18273`, tok: &css.Token{Type: css.HASH, Value: `18273`, Flag: "unrestricted"}},
		{s: `#`, tok: &css.Token{Type: css.DELIM, Value: `#`}},

		{s: `/`, tok: &css.Token{Type: css.DELIM, Value: `/`}},
		{s: `/* this is * a comment */#`, tok: &css.Token{Type: css.DELIM, Value: "#", Pos: css.Pos{Char: 25, Line: 0}}},

		{s: `<`, tok: &css.Token{Type: css.DELIM, Value: "<"}},
		{s: `<!`, tok: &css.Token{Type: css.DELIM, Value: "<"}},
		{s: `<!-`, tok: &css.Token{Type: css.DELIM, Value: "<"}},
		{s: `<!--`, tok: &css.Token{Type: css.CDO, Value: ""}},

		{s: `@`, tok: &css.Token{Type: css.DELIM, Value: "@"}},
		{s: `@foo`, tok: &css.Token{Type: css.ATKEYWORD, Value: "foo"}},

		{s: `\2603`, tok: &css.Token{Type: css.IDENT, Value: "☃"}},
		{s: `\`, tok: &css.Token{Type: css.IDENT, Value: "\uFFFD"}},
		{s: `\ `, tok: &css.Token{Type: css.IDENT, Value: " "}},
		{s: "\\\n", tok: &css.Token{Type: css.DELIM, Value: `\`}, err: "unescaped \\"},

		{s: `$=`, tok: &css.Token{Type: css.SUFFIXMATCH, Value: ``}},
		{s: `$X`, tok: &css.Token{Type: css.DELIM, Value: `$`}},
		{s: `$`, tok: &css.Token{Type: css.DELIM, Value: `$`}},

		{s: `*=`, tok: &css.Token{Type: css.SUBSTRINGMATCH, Value: ``}},
		{s: `*X`, tok: &css.Token{Type: css.DELIM, Value: `*`}},
		{s: `*`, tok: &css.Token{Type: css.DELIM, Value: `*`}},

		{s: `^=`, tok: &css.Token{Type: css.PREFIXMATCH, Value: ``}},
		{s: `^X`, tok: &css.Token{Type: css.DELIM, Value: `^`}},
		{s: `^`, tok: &css.Token{Type: css.DELIM, Value: `^`}},

		{s: `~=`, tok: &css.Token{Type: css.INCLUDEMATCH, Value: ``}},
		{s: `~X`, tok: &css.Token{Type: css.DELIM, Value: `~`}},
		{s: `~`, tok: &css.Token{Type: css.DELIM, Value: `~`}},

		{s: `|=`, tok: &css.Token{Type: css.DASHMATCH, Value: ``}},
		{s: `||`, tok: &css.Token{Type: css.COLUMN, Value: ``}},
		{s: `|X`, tok: &css.Token{Type: css.DELIM, Value: `|`}},
		{s: `|`, tok: &css.Token{Type: css.DELIM, Value: `|`}},

		{s: `,`, tok: &css.Token{Type: css.COMMA, Value: ``}},
		{s: `:`, tok: &css.Token{Type: css.COLON, Value: ``}},
		{s: `;`, tok: &css.Token{Type: css.SEMICOLON, Value: ``}},
		{s: `(`, tok: &css.Token{Type: css.LPAREN, Value: ``}},
		{s: `)`, tok: &css.Token{Type: css.RPAREN, Value: ``}},
		{s: `[`, tok: &css.Token{Type: css.LBRACK, Value: ``}},
		{s: `]`, tok: &css.Token{Type: css.RBRACK, Value: ``}},
		{s: `{`, tok: &css.Token{Type: css.LBRACE, Value: ``}},
		{s: `}`, tok: &css.Token{Type: css.RBRACE, Value: ``}},
	}

	for i, tt := range tests {
		// Skips over tests if test.iter is set.
		if *testiter > -1 && *testiter != i {
			continue
		}

		// Scan token.
		s := css.NewScanner(bytes.NewBufferString(tt.s))
		tok := s.Scan()

		// Verify properties.
		if !reflect.DeepEqual(tok, tt.tok) {
			t.Errorf("%d. <%q> tok: => got %#v, want %#v", i, tt.s, tok, tt.tok)
		} else if tt.err != "" {
			if len(s.Errors) == 0 {
				t.Errorf("%d. <%q> error expected", i, tt.s)
			} else if len(s.Errors) > 1 {
				t.Errorf("%d. <%q> too many errors occurred", i, tt.s)
			} else if s.Errors[0].Message != tt.err {
				t.Errorf("%d. <%q> error: got %q, want %q", i, tt.s, s.Errors[0].Message, tt.err)
			}
		} else if tt.err == "" && len(s.Errors) > 0 {
			t.Errorf("%d. <%q> unexpected error: %q", i, tt.s, s.Errors[0].Message)
		}
	}
}
