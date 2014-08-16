package css_test

import (
	"bytes"
	"flag"
	"reflect"
	"testing"

	"github.com/benbjohnson/css"
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
		tok css.ComponentValue
		err string
	}{
		{s: ``, tok: &css.Token{Tok: css.EOFToken}},
		{s: `   `, tok: &css.Token{Tok: css.WhitespaceToken, Value: `   `, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: " \n", tok: &css.Token{Tok: css.WhitespaceToken, Value: " \n", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: " \f", tok: &css.Token{Tok: css.WhitespaceToken, Value: " \n", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: " \r", tok: &css.Token{Tok: css.WhitespaceToken, Value: " \n", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: " \r ", tok: &css.Token{Tok: css.WhitespaceToken, Value: " \n", Pos: css.Pos{Char: 1, Line: 0}}},

		{s: `""`, tok: &css.Token{Tok: css.StringToken, Value: ``, Ending: '"', Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `"`, tok: &css.Token{Tok: css.StringToken, Value: ``, Ending: '"', Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `"foo`, tok: &css.Token{Tok: css.StringToken, Value: `foo`, Ending: '"', Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `"hello world"`, tok: &css.Token{Tok: css.StringToken, Value: `hello world`, Ending: '"', Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `'hello world'`, tok: &css.Token{Tok: css.StringToken, Value: `hello world`, Ending: '\'', Pos: css.Pos{Char: 1, Line: 0}}},
		{s: "'foo\\\nbar'", tok: &css.Token{Tok: css.StringToken, Value: "foo\nbar", Ending: '\'', Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `'foo\ bar'`, tok: &css.Token{Tok: css.StringToken, Value: `foo bar`, Ending: '\'', Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `'foo\\bar'`, tok: &css.Token{Tok: css.StringToken, Value: `foo\bar`, Ending: '\'', Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `'foo\`, tok: &css.Token{Tok: css.StringToken, Value: `foo`, Ending: '\'', Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `'frosty the \2603'`, tok: &css.Token{Tok: css.StringToken, Value: `frosty the ☃`, Ending: '\'', Pos: css.Pos{Char: 1, Line: 0}}},
		{s: "'foo bar\n", tok: &css.Token{Tok: css.BadStringToken, Value: ``, Pos: css.Pos{Char: 1, Line: 0}}},

		{s: `0`, tok: &css.Token{Tok: css.NumberToken, Type: "integer", Value: `0`, Number: 0.0, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `1.0`, tok: &css.Token{Tok: css.NumberToken, Type: "number", Value: `1.0`, Number: 1.0, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `1.123`, tok: &css.Token{Tok: css.NumberToken, Type: "number", Value: `1.123`, Number: 1.123, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `.001`, tok: &css.Token{Tok: css.NumberToken, Type: "number", Value: `.001`, Number: 0.001, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `-.001`, tok: &css.Token{Tok: css.NumberToken, Type: "number", Value: `-.001`, Number: -0.001, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `10000`, tok: &css.Token{Tok: css.NumberToken, Type: "integer", Value: `10000`, Number: 10000, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `10000.`, tok: &css.Token{Tok: css.NumberToken, Type: "integer", Value: `10000`, Number: 10000, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `100E`, tok: &css.Token{Tok: css.DimensionToken, Type: "integer", Value: `100E`, Number: 100, Unit: "E", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `100E+`, tok: &css.Token{Tok: css.DimensionToken, Type: "integer", Value: `100E`, Number: 100, Unit: "E", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `100E-`, tok: &css.Token{Tok: css.DimensionToken, Type: "integer", Value: `100E-`, Number: 100, Unit: "E-", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `1E2`, tok: &css.Token{Tok: css.NumberToken, Type: "number", Value: `1E2`, Number: 100, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `1.5E2`, tok: &css.Token{Tok: css.NumberToken, Type: "number", Value: `1.5E2`, Number: 150, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `1.5E+2`, tok: &css.Token{Tok: css.NumberToken, Type: "number", Value: `1.5E+2`, Number: 150, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `1.5E-2`, tok: &css.Token{Tok: css.NumberToken, Type: "number", Value: `1.5E-2`, Number: 0.015, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `+100`, tok: &css.Token{Tok: css.NumberToken, Type: "integer", Value: `+100`, Number: 100, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `+1.0`, tok: &css.Token{Tok: css.NumberToken, Type: "number", Value: `+1.0`, Number: 1, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `-100`, tok: &css.Token{Tok: css.NumberToken, Type: "integer", Value: `-100`, Number: -100, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `-1.0`, tok: &css.Token{Tok: css.NumberToken, Type: "number", Value: `-1.0`, Number: -1, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `-`, tok: &css.Token{Tok: css.DelimToken, Value: `-`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `-.`, tok: &css.Token{Tok: css.DelimToken, Value: `-`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `.`, tok: &css.Token{Tok: css.DelimToken, Value: `.`, Pos: css.Pos{Char: 1, Line: 0}}},

		{s: `url`, tok: &css.Token{Tok: css.IdentToken, Value: `url`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `-url`, tok: &css.Token{Tok: css.IdentToken, Value: `-url`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `myIdent`, tok: &css.Token{Tok: css.IdentToken, Value: `myIdent`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `my\2603`, tok: &css.Token{Tok: css.IdentToken, Value: `my☃`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `\2603`, tok: &css.Token{Tok: css.IdentToken, Value: `☃`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: "\000", tok: &css.Token{Tok: css.IdentToken, Value: "\uFFFD", Pos: css.Pos{Char: 1, Line: 0}}},

		{s: `url(`, tok: &css.Token{Tok: css.URLToken, Value: ``, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `url(foo`, tok: &css.Token{Tok: css.URLToken, Value: `foo`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `url(http://foo.com#bar?baz=bat)`, tok: &css.Token{Tok: css.URLToken, Value: `http://foo.com#bar?baz=bat`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `url(  foo`, tok: &css.Token{Tok: css.URLToken, Value: `foo`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `url(  foo  `, tok: &css.Token{Tok: css.URLToken, Value: `foo`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `url(  \2603  `, tok: &css.Token{Tok: css.URLToken, Value: `☃`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `url(foo)`, tok: &css.Token{Tok: css.URLToken, Value: `foo`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `url("http://foo.com#bar?baz=bat")`, tok: &css.Token{Tok: css.URLToken, Value: `http://foo.com#bar?baz=bat`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `url(  "foo"  `, tok: &css.Token{Tok: css.URLToken, Value: `foo`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `url("foo"  `, tok: &css.Token{Tok: css.URLToken, Value: `foo`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `url("foo")`, tok: &css.Token{Tok: css.URLToken, Value: `foo`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `url("foo"x`, tok: &css.Token{Tok: css.BadURLToken, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `url("foo" x`, tok: &css.Token{Tok: css.BadURLToken, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: "url('foo\n", tok: &css.Token{Tok: css.BadURLToken, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `url(foo"`, tok: &css.Token{Tok: css.BadURLToken, Pos: css.Pos{Char: 1, Line: 0}}, err: `invalid url code point: " (U+0022)`},
		{s: `url(foo bar)`, tok: &css.Token{Tok: css.BadURLToken, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `url(foo'`, tok: &css.Token{Tok: css.BadURLToken, Pos: css.Pos{Char: 1, Line: 0}}, err: `invalid url code point: ' (U+0027)`},
		{s: `url(foo(`, tok: &css.Token{Tok: css.BadURLToken, Pos: css.Pos{Char: 1, Line: 0}}, err: `invalid url code point: ( (U+0028)`},
		{s: "url(foo\001 \\2603", tok: &css.Token{Tok: css.BadURLToken, Pos: css.Pos{Char: 1, Line: 0}}, err: "invalid url code point: \001 (U+0001)"},
		{s: "url(foo\\\n", tok: &css.Token{Tok: css.BadURLToken, Pos: css.Pos{Char: 1, Line: 0}}, err: `unescaped \ in url`},
		{s: "url(foo\001 \001", tok: &css.Token{Tok: css.BadURLToken, Pos: css.Pos{Char: 1, Line: 0}}, err: "invalid url code point: \001 (U+0001)"},

		{s: `myFunc(`, tok: &css.Token{Tok: css.FunctionToken, Value: `myFunc`, Pos: css.Pos{Char: 1, Line: 0}}},

		{s: "u+A", tok: &css.Token{Tok: css.UnicodeRangeToken, Start: 10, End: 10, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: "u+00000A", tok: &css.Token{Tok: css.UnicodeRangeToken, Start: 10, End: 10, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: "u+000000A", tok: &css.Token{Tok: css.UnicodeRangeToken, Start: 0, End: 0, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: "u+1?", tok: &css.Token{Tok: css.UnicodeRangeToken, Start: 16, End: 31, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: "u+1?F", tok: &css.Token{Tok: css.UnicodeRangeToken, Start: 16, End: 31, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: "u+02-04", tok: &css.Token{Tok: css.UnicodeRangeToken, Start: 2, End: 4, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: "u+02-04?", tok: &css.Token{Tok: css.UnicodeRangeToken, Start: 2, End: 4, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: "u+02-0000004", tok: &css.Token{Tok: css.UnicodeRangeToken, Start: 2, End: 0, Pos: css.Pos{Char: 1, Line: 0}}},

		{s: `100em`, tok: &css.Token{Tok: css.DimensionToken, Type: "integer", Value: `100em`, Number: 100, Unit: "em", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `-1.2in`, tok: &css.Token{Tok: css.DimensionToken, Type: "number", Value: `-1.2in`, Number: -1.2, Unit: "in", Pos: css.Pos{Char: 1, Line: 0}}},

		{s: `100%`, tok: &css.Token{Tok: css.PercentageToken, Type: "integer", Value: `100%`, Number: 100, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `-0.2%`, tok: &css.Token{Tok: css.PercentageToken, Type: "number", Value: `-0.2%`, Number: -0.2, Pos: css.Pos{Char: 1, Line: 0}}},

		{s: `#foo`, tok: &css.Token{Tok: css.HashToken, Value: `foo`, Type: "id", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `#foo\2603 bar`, tok: &css.Token{Tok: css.HashToken, Value: `foo☃bar`, Type: "id", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `#-x`, tok: &css.Token{Tok: css.HashToken, Value: `-x`, Type: "id", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `#_x`, tok: &css.Token{Tok: css.HashToken, Value: `_x`, Type: "id", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `#18273`, tok: &css.Token{Tok: css.HashToken, Value: `18273`, Type: "unrestricted", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `#`, tok: &css.Token{Tok: css.DelimToken, Value: `#`, Pos: css.Pos{Char: 1, Line: 0}}},

		{s: `/`, tok: &css.Token{Tok: css.DelimToken, Value: `/`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `/* this is * a comment */#`, tok: &css.Token{Tok: css.DelimToken, Value: "#", Pos: css.Pos{Char: 26, Line: 0}}},
		{s: `/* this is a comment`, tok: &css.Token{Tok: css.EOFToken, Pos: css.Pos{Char: 20, Line: 0}}},

		{s: `<`, tok: &css.Token{Tok: css.DelimToken, Value: "<", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `<!`, tok: &css.Token{Tok: css.DelimToken, Value: "<", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `<!-`, tok: &css.Token{Tok: css.DelimToken, Value: "<", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `<!--`, tok: &css.Token{Tok: css.CDOToken, Pos: css.Pos{Char: 1, Line: 0}}},

		{s: `@`, tok: &css.Token{Tok: css.DelimToken, Value: "@", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `@foo`, tok: &css.Token{Tok: css.AtKeywordToken, Value: "foo", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `@\2603`, tok: &css.Token{Tok: css.AtKeywordToken, Value: "☃", Pos: css.Pos{Char: 1, Line: 0}}},

		{s: `\2603`, tok: &css.Token{Tok: css.IdentToken, Value: "☃", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `\`, tok: &css.Token{Tok: css.IdentToken, Value: "\uFFFD", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `\ `, tok: &css.Token{Tok: css.IdentToken, Value: " ", Pos: css.Pos{Char: 1, Line: 0}}},
		{s: "\\\n", tok: &css.Token{Tok: css.DelimToken, Value: `\`, Pos: css.Pos{Char: 1, Line: 0}}, err: "unescaped \\"},

		{s: `$=`, tok: &css.Token{Tok: css.SuffixMatchToken, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `$X`, tok: &css.Token{Tok: css.DelimToken, Value: `$`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `$`, tok: &css.Token{Tok: css.DelimToken, Value: `$`, Pos: css.Pos{Char: 1, Line: 0}}},

		{s: `*=`, tok: &css.Token{Tok: css.SubstringMatchToken, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `*X`, tok: &css.Token{Tok: css.DelimToken, Value: `*`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `*`, tok: &css.Token{Tok: css.DelimToken, Value: `*`, Pos: css.Pos{Char: 1, Line: 0}}},

		{s: `^=`, tok: &css.Token{Tok: css.PrefixMatchToken, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `^X`, tok: &css.Token{Tok: css.DelimToken, Value: `^`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `^`, tok: &css.Token{Tok: css.DelimToken, Value: `^`, Pos: css.Pos{Char: 1, Line: 0}}},

		{s: `~=`, tok: &css.Token{Tok: css.IncludeMatchToken, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `~X`, tok: &css.Token{Tok: css.DelimToken, Value: `~`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `~`, tok: &css.Token{Tok: css.DelimToken, Value: `~`, Pos: css.Pos{Char: 1, Line: 0}}},

		{s: `|=`, tok: &css.Token{Tok: css.DashMatchToken, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `||`, tok: &css.Token{Tok: css.ColumnToken, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `|X`, tok: &css.Token{Tok: css.DelimToken, Value: `|`, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `|`, tok: &css.Token{Tok: css.DelimToken, Value: `|`, Pos: css.Pos{Char: 1, Line: 0}}},

		{s: `,`, tok: &css.Token{Tok: css.CommaToken, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `:`, tok: &css.Token{Tok: css.ColonToken, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `;`, tok: &css.Token{Tok: css.SemicolonToken, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `(`, tok: &css.Token{Tok: css.LParenToken, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `)`, tok: &css.Token{Tok: css.RParenToken, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `[`, tok: &css.Token{Tok: css.LBrackToken, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `]`, tok: &css.Token{Tok: css.RBrackToken, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `{`, tok: &css.Token{Tok: css.LBraceToken, Pos: css.Pos{Char: 1, Line: 0}}},
		{s: `}`, tok: &css.Token{Tok: css.RBraceToken, Pos: css.Pos{Char: 1, Line: 0}}},
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
			t.Errorf("%d. <%q> tok: =>\n\ngot %#v\n\nwant %#v\n\n", i, tt.s, tok, tt.tok)
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
