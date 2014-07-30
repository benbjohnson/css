package css

// Token represents a lexical token.
type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
	EOF

	// CSS Standard tokens
	IDENT
	FUNCTION
	ATKEYWORD
	HASH
	STRING
	BADSTRING
	URL
	BADURL
	DELIM
	NUMBER
	PERCENTAGE
	DIMENSION
	UNICODERANGE
	INCLUDEMATCH
	DASHMATCH
	PREFIXMATCH
	SUFFIXMATCH
	SUBSTRINGMATCH
	COLUMN
	WHITESPACE
	CDO
	CDC
	COLON
	SEMICOLON
	COMMA
	LBRACK
	RBRACK
	LPAREN
	RPAREN
	LBRACE
	RBRACE
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",

	IDENT:          "ident",
	FUNCTION:       "function",
	ATKEYWORD:      "at-keyword",
	HASH:           "hash",
	STRING:         "string",
	BADSTRING:      "bad-string",
	URL:            "url",
	BADURL:         "bad-url",
	DELIM:          "delim",
	NUMBER:         "number",
	PERCENTAGE:     "percentage",
	DIMENSION:      "dimension",
	UNICODERANGE:   "unicode-range",
	INCLUDEMATCH:   "include-match",
	DASHMATCH:      "dash-match",
	PREFIXMATCH:    "prefix-match",
	SUFFIXMATCH:    "suffix-match",
	SUBSTRINGMATCH: "substring-match",
	COLUMN:         "column",
	WHITESPACE:     "whitespace",
	CDO:            "CDO",
	CDC:            "CDC",
	COLON:          "colon",
	SEMICOLON:      "semicolon",
	COMMA:          "comma",
	LBRACK:         "[",
	RBRACK:         "]",
	LPAREN:         "(",
	RPAREN:         ")",
	LBRACE:         "{",
	RBRACE:         "}",
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
