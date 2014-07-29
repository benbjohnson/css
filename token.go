package css

// Token represents a lexical token.
type Token int

const (
	// Special tokens
	IllegalToken Token = iota
	EOF

	// CSS Standard tokens

)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
}

// String returns the string representation of the token.
func (tok Token) String() string {
	if tok >= 0 && tok < Token(len(tokens)) {
		return tokens[tok]
	}
	return ""
}

// Pos specifies the line and character position of a token.
// The Char and Line are both zero-based indexes.
type Pos struct {
	Char int
	Line int
}
