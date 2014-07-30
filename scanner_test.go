package css_test

import (
	"bytes"
	"testing"

	"github.com/benbjohnson/css"
)

// Ensure than the scanner returns appropriate tokens and literals.
func TestScanner_Scan(t *testing.T) {
	var tests = []struct {
		str   string
		tok   css.Token
		value string
	}{
		{``, css.EOF, ``},
		{`   `, css.WHITESPACE, `   `},
		{`""`, css.STRING, ``},
		{`"`, css.STRING, ``},
		{`"foo`, css.STRING, `foo`},
		{`"hello world"`, css.STRING, `hello world`},
		{`'hello world'`, css.STRING, `hello world`},
		{"'foo\\\nbar'", css.STRING, "foo\nbar"},
		{`'foo\ bar'`, css.STRING, `foo\ bar`},
	}

	for i, tt := range tests {
		s := css.NewScanner(bytes.NewBufferString(tt.str))
		_, tok := s.Scan()
		if tok != tt.tok {
			t.Errorf("%d. tok: => got %q, want %q", i, tok, tt.tok)
		} else if s.Value != tt.value {
			t.Errorf("%d. value: got %q, want %q", i, s.Value, tt.value)
		}
	}
}
