package css_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/benbjohnson/css"
)

// Ensure than the scanner returns appropriate tokens and literals.
func TestScanner_Scan(t *testing.T) {
	var tests = []struct {
		s      string
		tok    css.Token
		typ    string
		value  string
		num    float64
		unit   string
		start  int
		end    int
		ending rune
		err    string
	}{
		{s: ``, tok: css.EOF},
		{s: `   `, tok: css.WHITESPACE, value: `   `},

		{s: `""`, tok: css.STRING, value: ``, ending: '"'},
		{s: `"`, tok: css.STRING, value: ``, ending: '"'},
		{s: `"foo`, tok: css.STRING, value: `foo`, ending: '"'},
		{s: `"hello world"`, tok: css.STRING, value: `hello world`, ending: '"'},
		{s: `'hello world'`, tok: css.STRING, value: `hello world`, ending: '\''},
		{s: "'foo\\\nbar'", tok: css.STRING, value: "foo\nbar", ending: '\''},
		{s: `'foo\ bar'`, tok: css.STRING, value: `foo bar`, ending: '\''},
		{s: `'foo\\bar'`, tok: css.STRING, value: `foo\bar`, ending: '\''},
		{s: `'frosty the \2603'`, tok: css.STRING, value: `frosty the ☃`, ending: '\''},

		{s: `0`, tok: css.NUMBER, typ: "integer", value: `0`, num: 0.0},
		{s: `1.0`, tok: css.NUMBER, typ: "number", value: `1.0`, num: 1.0},
		{s: `1.123`, tok: css.NUMBER, typ: "number", value: `1.123`, num: 1.123},
		{s: `.001`, tok: css.NUMBER, typ: "number", value: `.001`, num: 0.001},
		{s: `-.001`, tok: css.NUMBER, typ: "number", value: `-.001`, num: -0.001},
		{s: `10000`, tok: css.NUMBER, typ: "integer", value: `10000`, num: 10000},
		{s: `10000.`, tok: css.NUMBER, typ: "integer", value: `10000`, num: 10000},
		{s: `100E`, tok: css.NUMBER, typ: "integer", value: `100`, num: 100},
		{s: `100E+`, tok: css.NUMBER, typ: "integer", value: `100`, num: 100},
		{s: `100E-`, tok: css.NUMBER, typ: "integer", value: `100`, num: 100},
		{s: `1E2`, tok: css.NUMBER, typ: "number", value: `1E2`, num: 100},
		{s: `1.5E2`, tok: css.NUMBER, typ: "number", value: `1.5E2`, num: 150},
		{s: `1.5E+2`, tok: css.NUMBER, typ: "number", value: `1.5E+2`, num: 150},
		{s: `1.5E-2`, tok: css.NUMBER, typ: "number", value: `1.5E-2`, num: 0.015},
		{s: `+100`, tok: css.NUMBER, typ: "integer", value: `+100`, num: 100},
		{s: `+1.0`, tok: css.NUMBER, typ: "number", value: `+1.0`, num: 1},
		{s: `-100`, tok: css.NUMBER, typ: "integer", value: `-100`, num: -100},
		{s: `-1.0`, tok: css.NUMBER, typ: "number", value: `-1.0`, num: -1},
		{s: `-`, tok: css.DELIM, value: `-`},

		{s: `100em`, tok: css.DIMENSION, typ: "integer", value: `100em`, num: 100, unit: "em"},
		{s: `-1.2in`, tok: css.DIMENSION, typ: "number", value: `-1.2in`, num: -1.2, unit: "in"},

		{s: `100%`, tok: css.PERCENTAGE, typ: "integer", value: `100%`, num: 100},
		{s: `-0.2%`, tok: css.PERCENTAGE, typ: "number", value: `-0.2%`, num: -0.2},

		{s: `#foo`, tok: css.HASH, value: `foo`, typ: "id"},
		{s: `#foo\2603 bar`, tok: css.HASH, value: `foo☃bar`, typ: "id"},
		{s: `#-x`, tok: css.HASH, value: `-x`, typ: "id"},
		{s: `#_x`, tok: css.HASH, value: `_x`, typ: "id"},
		{s: `#18273`, tok: css.HASH, value: `18273`},
		{s: `#`, tok: css.DELIM, value: `#`},

		{s: `/`, tok: css.DELIM, value: `/`},
		{s: `/* this is * a comment */#`, tok: css.DELIM, value: "#"},

		{s: `<`, tok: css.DELIM, value: "<"},
		{s: `<!`, tok: css.DELIM, value: "<"},
		{s: `<!-`, tok: css.DELIM, value: "<"},
		{s: `<!--`, tok: css.CDO, value: ""},

		{s: `@`, tok: css.DELIM, value: "@"},
		{s: `@foo`, tok: css.ATKEYWORD, value: "foo"},

		{s: `\2603`, tok: css.IDENT, value: "☃"},
		{s: `\`, tok: css.IDENT, value: ""},
		{s: `\ `, tok: css.IDENT, value: " "},
		{s: "\\\n", tok: css.DELIM, value: `\`, err: "unescaped \\"},

		{s: `$=`, tok: css.SUFFIXMATCH, value: ``},
		{s: `$X`, tok: css.DELIM, value: `$`},
		{s: `$`, tok: css.DELIM, value: `$`},

		{s: `*=`, tok: css.SUBSTRINGMATCH, value: ``},
		{s: `*X`, tok: css.DELIM, value: `*`},
		{s: `*`, tok: css.DELIM, value: `*`},

		{s: `^=`, tok: css.PREFIXMATCH, value: ``},
		{s: `^X`, tok: css.DELIM, value: `^`},
		{s: `^`, tok: css.DELIM, value: `^`},

		{s: `~=`, tok: css.INCLUDEMATCH, value: ``},
		{s: `~X`, tok: css.DELIM, value: `~`},
		{s: `~`, tok: css.DELIM, value: `~`},

		{s: `,`, tok: css.COMMA, value: ``},
		{s: `:`, tok: css.COLON, value: ``},
		{s: `;`, tok: css.SEMICOLON, value: ``},
		{s: `(`, tok: css.LPAREN, value: ``},
		{s: `)`, tok: css.RPAREN, value: ``},
		{s: `[`, tok: css.LBRACK, value: ``},
		{s: `]`, tok: css.RBRACK, value: ``},
		{s: `{`, tok: css.LBRACE, value: ``},
		{s: `}`, tok: css.RBRACE, value: ``},
	}

	for i, tt := range tests {
		//if i != 48 {
		//	continue
		//}

		// Set test defaults.
		if tt.typ == "" {
			tt.typ = "unrestricted"
		}

		// Scan token.
		s := css.NewScanner(bytes.NewBufferString(tt.s))
		_, tok := s.Scan()

		// Verify properties.
		if tok != tt.tok {
			t.Errorf("%d. <%q> tok: => got %q, want %q", i, tt.s, tok, tt.tok)
		} else if s.Type != tt.typ {
			t.Errorf("%d. <%q> type: got %q, want %q", i, tt.s, s.Type, tt.typ)
		} else if act, exp := fmt.Sprintf("%0.3f", s.Number), fmt.Sprintf("%0.3f", tt.num); exp != act {
			t.Errorf("%d. <%q> number: got %q, want %q", i, tt.s, act, exp)
		} else if s.Value != tt.value {
			t.Errorf("%d. <%q> value: got %q, want %q", i, tt.s, s.Value, tt.value)
		} else if s.Unit != tt.unit {
			t.Errorf("%d. <%q> unit: got %q, want %q", i, tt.s, s.Unit, tt.unit)
		} else if s.Start != tt.start {
			t.Errorf("%d. <%q> start: got %q, want %q", i, tt.s, s.Start, tt.start)
		} else if s.End != tt.end {
			t.Errorf("%d. <%q> end: got %q, want %q", i, tt.s, s.End, tt.end)
		} else if s.Ending != tt.ending {
			t.Errorf("%d. <%q> ending: got %q, want %q", i, tt.s, s.Ending, tt.ending)
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
