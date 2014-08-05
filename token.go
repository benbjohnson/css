package css

// Token represents a lexical token.
type Token struct {
	// Type represents the type of token.
	Type TokenType

	// Flag is set after parsing an ident-token, function-token,
	// at-keyword-token, hash-token, string-token, and url-token.
	// It is set to either "id" or "unrestricted".
	//
	// This is also set after scanning a numeric token and can
	// be set to "integer" or "number".
	Flag string

	// Value is the literal representation of the last read token.
	Value string

	// This numeric value is set after scanning a number-token,
	// a percentage-token, or a dimension-token. The unit is
	// returned for dimension tokens.
	Number float64
	Unit   string

	// Start and End are set after each unicode-range token.
	Start int
	End   int

	// Ending represents the ending code point of a string.
	Ending rune

	// Pos represents the position of the token in the stream.
	Pos Pos
}

// TokenType represents a type of CSS3 token.
type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
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

var types = [...]string{
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
func (tok TokenType) String() string {
	if tok >= 0 && tok < TokenType(len(types)) {
		return types[tok]
	}
	return ""
}

// Pos specifies the line and character position of a token.
// The Char and Line are both zero-based indexes.
type Pos struct {
	Char int
	Line int
}
