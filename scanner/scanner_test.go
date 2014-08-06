package scanner_test

import (
	"bytes"
	"flag"
	"reflect"
	"testing"

	"github.com/benbjohnson/css/scanner"
	"github.com/benbjohnson/css/token"
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
		tok token.Token
		err string
	}{
		{s: ``, tok: &token.EOF{}},
		{s: `   `, tok: &token.Whitespace{Value: `   `, Pos: token.Pos{1, 0}}},

		{s: `""`, tok: &token.String{Value: ``, Ending: '"', Pos: token.Pos{1, 0}}},
		{s: `"`, tok: &token.String{Value: ``, Ending: '"', Pos: token.Pos{1, 0}}},
		{s: `"foo`, tok: &token.String{Value: `foo`, Ending: '"', Pos: token.Pos{1, 0}}},
		{s: `"hello world"`, tok: &token.String{Value: `hello world`, Ending: '"', Pos: token.Pos{1, 0}}},
		{s: `'hello world'`, tok: &token.String{Value: `hello world`, Ending: '\'', Pos: token.Pos{1, 0}}},
		{s: "'foo\\\nbar'", tok: &token.String{Value: "foo\nbar", Ending: '\'', Pos: token.Pos{1, 0}}},
		{s: `'foo\ bar'`, tok: &token.String{Value: `foo bar`, Ending: '\'', Pos: token.Pos{1, 0}}},
		{s: `'foo\\bar'`, tok: &token.String{Value: `foo\bar`, Ending: '\'', Pos: token.Pos{1, 0}}},
		{s: `'frosty the \2603'`, tok: &token.String{Value: `frosty the ☃`, Ending: '\'', Pos: token.Pos{1, 0}}},

		{s: `0`, tok: &token.Number{Type: "integer", Value: `0`, Number: 0.0, Pos: token.Pos{1, 0}}},
		{s: `1.0`, tok: &token.Number{Type: "number", Value: `1.0`, Number: 1.0, Pos: token.Pos{1, 0}}},
		{s: `1.123`, tok: &token.Number{Type: "number", Value: `1.123`, Number: 1.123, Pos: token.Pos{1, 0}}},
		{s: `.001`, tok: &token.Number{Type: "number", Value: `.001`, Number: 0.001, Pos: token.Pos{1, 0}}},
		{s: `-.001`, tok: &token.Number{Type: "number", Value: `-.001`, Number: -0.001, Pos: token.Pos{1, 0}}},
		{s: `10000`, tok: &token.Number{Type: "integer", Value: `10000`, Number: 10000, Pos: token.Pos{1, 0}}},
		{s: `10000.`, tok: &token.Number{Type: "integer", Value: `10000`, Number: 10000, Pos: token.Pos{1, 0}}},
		{s: `100E`, tok: &token.Dimension{Type: "integer", Value: `100E`, Number: 100, Unit: "E", Pos: token.Pos{1, 0}}},
		{s: `100E+`, tok: &token.Dimension{Type: "integer", Value: `100E`, Number: 100, Unit: "E", Pos: token.Pos{1, 0}}},
		{s: `100E-`, tok: &token.Dimension{Type: "integer", Value: `100E-`, Number: 100, Unit: "E-", Pos: token.Pos{1, 0}}},
		{s: `1E2`, tok: &token.Number{Type: "number", Value: `1E2`, Number: 100, Pos: token.Pos{1, 0}}},
		{s: `1.5E2`, tok: &token.Number{Type: "number", Value: `1.5E2`, Number: 150, Pos: token.Pos{1, 0}}},
		{s: `1.5E+2`, tok: &token.Number{Type: "number", Value: `1.5E+2`, Number: 150, Pos: token.Pos{1, 0}}},
		{s: `1.5E-2`, tok: &token.Number{Type: "number", Value: `1.5E-2`, Number: 0.015, Pos: token.Pos{1, 0}}},
		{s: `+100`, tok: &token.Number{Type: "integer", Value: `+100`, Number: 100, Pos: token.Pos{1, 0}}},
		{s: `+1.0`, tok: &token.Number{Type: "number", Value: `+1.0`, Number: 1, Pos: token.Pos{1, 0}}},
		{s: `-100`, tok: &token.Number{Type: "integer", Value: `-100`, Number: -100, Pos: token.Pos{1, 0}}},
		{s: `-1.0`, tok: &token.Number{Type: "number", Value: `-1.0`, Number: -1, Pos: token.Pos{1, 0}}},
		{s: `-`, tok: &token.Delim{Value: `-`, Pos: token.Pos{1, 0}}},

		{s: `url`, tok: &token.Ident{Value: `url`, Pos: token.Pos{1, 0}}},
		{s: `myIdent`, tok: &token.Ident{Value: `myIdent`, Pos: token.Pos{1, 0}}},
		{s: `my\2603`, tok: &token.Ident{Value: `my☃`, Pos: token.Pos{1, 0}}},

		{s: `url(`, tok: &token.URL{Value: ``, Pos: token.Pos{1, 0}}},
		{s: `url(foo`, tok: &token.URL{Value: `foo`, Pos: token.Pos{1, 0}}},
		{s: `url(http://foo.com#bar?baz=bat)`, tok: &token.URL{Value: `http://foo.com#bar?baz=bat`, Pos: token.Pos{1, 0}}},
		{s: `url(  foo`, tok: &token.URL{Value: `foo`, Pos: token.Pos{1, 0}}},
		{s: `url(  foo  `, tok: &token.URL{Value: `foo`, Pos: token.Pos{1, 0}}},
		{s: `url(  \2603  `, tok: &token.URL{Value: `☃`, Pos: token.Pos{1, 0}}},
		{s: `url(foo)`, tok: &token.URL{Value: `foo`, Pos: token.Pos{1, 0}}},
		{s: `url("http://foo.com#bar?baz=bat")`, tok: &token.URL{Value: `http://foo.com#bar?baz=bat`, Pos: token.Pos{1, 0}}},
		{s: `url(  "foo"  `, tok: &token.URL{Value: `foo`, Pos: token.Pos{1, 0}}},
		{s: `url("foo"  `, tok: &token.URL{Value: `foo`, Pos: token.Pos{1, 0}}},
		{s: `url("foo")`, tok: &token.URL{Value: `foo`, Pos: token.Pos{1, 0}}},
		{s: `url("foo"x`, tok: &token.BadURL{Pos: token.Pos{1, 0}}},
		{s: `url("foo" x`, tok: &token.BadURL{Pos: token.Pos{1, 0}}},
		{s: `url(foo"`, tok: &token.BadURL{Pos: token.Pos{1, 0}}, err: `invalid url code point: " (U+0022)`},
		{s: `url(foo'`, tok: &token.BadURL{Pos: token.Pos{1, 0}}, err: `invalid url code point: ' (U+0027)`},
		{s: `url(foo(`, tok: &token.BadURL{Pos: token.Pos{1, 0}}, err: `invalid url code point: ( (U+0028)`},
		{s: "url(foo\001", tok: &token.BadURL{Pos: token.Pos{1, 0}}, err: "invalid url code point: \001 (U+0001)"},
		{s: "url(foo\\\n", tok: &token.BadURL{Pos: token.Pos{1, 0}}, err: `unescaped \ in url`},

		{s: `myFunc(`, tok: &token.Function{Value: `myFunc`, Pos: token.Pos{1, 0}}},

		{s: "u+A", tok: &token.UnicodeRange{Start: 10, End: 10, Pos: token.Pos{1, 0}}},
		{s: "u+00000A", tok: &token.UnicodeRange{Start: 10, End: 10, Pos: token.Pos{1, 0}}},
		{s: "u+000000A", tok: &token.UnicodeRange{Start: 0, End: 0, Pos: token.Pos{1, 0}}},
		{s: "u+1?", tok: &token.UnicodeRange{Start: 16, End: 31, Pos: token.Pos{1, 0}}},
		{s: "u+1?F", tok: &token.UnicodeRange{Start: 16, End: 31, Pos: token.Pos{1, 0}}},
		{s: "u+02-04", tok: &token.UnicodeRange{Start: 2, End: 4, Pos: token.Pos{1, 0}}},
		{s: "u+02-04?", tok: &token.UnicodeRange{Start: 2, End: 4, Pos: token.Pos{1, 0}}},
		{s: "u+02-0000004", tok: &token.UnicodeRange{Start: 2, End: 0, Pos: token.Pos{1, 0}}},

		{s: `100em`, tok: &token.Dimension{Type: "integer", Value: `100em`, Number: 100, Unit: "em", Pos: token.Pos{1, 0}}},
		{s: `-1.2in`, tok: &token.Dimension{Type: "number", Value: `-1.2in`, Number: -1.2, Unit: "in", Pos: token.Pos{1, 0}}},

		{s: `100%`, tok: &token.Percentage{Type: "integer", Value: `100%`, Number: 100, Pos: token.Pos{1, 0}}},
		{s: `-0.2%`, tok: &token.Percentage{Type: "number", Value: `-0.2%`, Number: -0.2, Pos: token.Pos{1, 0}}},

		{s: `#foo`, tok: &token.Hash{Value: `foo`, Type: "id", Pos: token.Pos{1, 0}}},
		{s: `#foo\2603 bar`, tok: &token.Hash{Value: `foo☃bar`, Type: "id", Pos: token.Pos{1, 0}}},
		{s: `#-x`, tok: &token.Hash{Value: `-x`, Type: "id", Pos: token.Pos{1, 0}}},
		{s: `#_x`, tok: &token.Hash{Value: `_x`, Type: "id", Pos: token.Pos{1, 0}}},
		{s: `#18273`, tok: &token.Hash{Value: `18273`, Type: "unrestricted", Pos: token.Pos{1, 0}}},
		{s: `#`, tok: &token.Delim{Value: `#`, Pos: token.Pos{1, 0}}},

		{s: `/`, tok: &token.Delim{Value: `/`, Pos: token.Pos{1, 0}}},
		{s: `/* this is * a comment */#`, tok: &token.Delim{Value: "#", Pos: token.Pos{26, 0}}},

		{s: `<`, tok: &token.Delim{Value: "<", Pos: token.Pos{1, 0}}},
		{s: `<!`, tok: &token.Delim{Value: "<", Pos: token.Pos{1, 0}}},
		{s: `<!-`, tok: &token.Delim{Value: "<", Pos: token.Pos{1, 0}}},
		{s: `<!--`, tok: &token.CDO{Pos: token.Pos{1, 0}}},

		{s: `@`, tok: &token.Delim{Value: "@", Pos: token.Pos{1, 0}}},
		{s: `@foo`, tok: &token.AtKeyword{Value: "foo", Pos: token.Pos{1, 0}}},

		{s: `\2603`, tok: &token.Ident{Value: "☃", Pos: token.Pos{1, 0}}},
		{s: `\`, tok: &token.Ident{Value: "\uFFFD", Pos: token.Pos{1, 0}}},
		{s: `\ `, tok: &token.Ident{Value: " ", Pos: token.Pos{1, 0}}},
		{s: "\\\n", tok: &token.Delim{Value: `\`, Pos: token.Pos{1, 0}}, err: "unescaped \\"},

		{s: `$=`, tok: &token.SuffixMatch{Pos: token.Pos{1, 0}}},
		{s: `$X`, tok: &token.Delim{Value: `$`, Pos: token.Pos{1, 0}}},
		{s: `$`, tok: &token.Delim{Value: `$`, Pos: token.Pos{1, 0}}},

		{s: `*=`, tok: &token.SubstringMatch{Pos: token.Pos{1, 0}}},
		{s: `*X`, tok: &token.Delim{Value: `*`, Pos: token.Pos{1, 0}}},
		{s: `*`, tok: &token.Delim{Value: `*`, Pos: token.Pos{1, 0}}},

		{s: `^=`, tok: &token.PrefixMatch{Pos: token.Pos{1, 0}}},
		{s: `^X`, tok: &token.Delim{Value: `^`, Pos: token.Pos{1, 0}}},
		{s: `^`, tok: &token.Delim{Value: `^`, Pos: token.Pos{1, 0}}},

		{s: `~=`, tok: &token.IncludeMatch{Pos: token.Pos{1, 0}}},
		{s: `~X`, tok: &token.Delim{Value: `~`, Pos: token.Pos{1, 0}}},
		{s: `~`, tok: &token.Delim{Value: `~`, Pos: token.Pos{1, 0}}},

		{s: `|=`, tok: &token.DashMatch{Pos: token.Pos{1, 0}}},
		{s: `||`, tok: &token.Column{Pos: token.Pos{1, 0}}},
		{s: `|X`, tok: &token.Delim{Value: `|`, Pos: token.Pos{1, 0}}},
		{s: `|`, tok: &token.Delim{Value: `|`, Pos: token.Pos{1, 0}}},

		{s: `,`, tok: &token.Comma{Pos: token.Pos{1, 0}}},
		{s: `:`, tok: &token.Colon{Pos: token.Pos{1, 0}}},
		{s: `;`, tok: &token.Semicolon{Pos: token.Pos{1, 0}}},
		{s: `(`, tok: &token.LParen{Pos: token.Pos{1, 0}}},
		{s: `)`, tok: &token.RParen{Pos: token.Pos{1, 0}}},
		{s: `[`, tok: &token.LBrack{Pos: token.Pos{1, 0}}},
		{s: `]`, tok: &token.RBrack{Pos: token.Pos{1, 0}}},
		{s: `{`, tok: &token.LBrace{Pos: token.Pos{1, 0}}},
		{s: `}`, tok: &token.RBrace{Pos: token.Pos{1, 0}}},
	}

	for i, tt := range tests {
		// Skips over tests if test.iter is set.
		if *testiter > -1 && *testiter != i {
			continue
		}

		// Scan token.
		s := scanner.New(bytes.NewBufferString(tt.s))
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
