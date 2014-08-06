package token

// Token represents a lexical token.
type Token interface {
	token()
}

func (_ *Ident) token()          {}
func (_ *Function) token()       {}
func (_ *AtKeyword) token()      {}
func (_ *Hash) token()           {}
func (_ *String) token()         {}
func (_ *BadString) token()      {}
func (_ *URL) token()            {}
func (_ *BadURL) token()         {}
func (_ *Delim) token()          {}
func (_ *Number) token()         {}
func (_ *Percentage) token()     {}
func (_ *Dimension) token()      {}
func (_ *UnicodeRange) token()   {}
func (_ *IncludeMatch) token()   {}
func (_ *DashMatch) token()      {}
func (_ *PrefixMatch) token()    {}
func (_ *SuffixMatch) token()    {}
func (_ *SubstringMatch) token() {}
func (_ *Column) token()         {}
func (_ *Whitespace) token()     {}
func (_ *CDO) token()            {}
func (_ *CDC) token()            {}
func (_ *Colon) token()          {}
func (_ *Semicolon) token()      {}
func (_ *Comma) token()          {}
func (_ *LBrack) token()         {}
func (_ *RBrack) token()         {}
func (_ *LParen) token()         {}
func (_ *RParen) token()         {}
func (_ *LBrace) token()         {}
func (_ *RBrace) token()         {}
func (_ *EOF) token()            {}

type Ident struct {
	Type  string
	Value string
	Pos   Pos
}

type Function struct {
	Type  string
	Value string
	Pos   Pos
}

type AtKeyword struct {
	Type  string
	Value string
	Pos   Pos
}

type Hash struct {
	Type  string
	Value string
	Pos   Pos
}

type String struct {
	Type   string
	Ending rune
	Value  string
	Pos    Pos
}

type BadString struct {
	Pos Pos
}

type URL struct {
	Type  string
	Value string
	Pos   Pos
}

type BadURL struct {
	Pos Pos
}

type Delim struct {
	Value string
	Pos   Pos
}

type Number struct {
	Type   string
	Number float64
	Value  string
	Pos    Pos
}

type Percentage struct {
	Type   string
	Number float64
	Value  string
	Pos    Pos
}

type Dimension struct {
	Type   string
	Number float64
	Unit   string
	Value  string
	Pos    Pos
}

type UnicodeRange struct {
	Start int
	End   int
	Pos   Pos
}

type IncludeMatch struct {
	Pos Pos
}
type DashMatch struct {
	Pos Pos
}
type PrefixMatch struct {
	Pos Pos
}
type SuffixMatch struct {
	Pos Pos
}
type SubstringMatch struct {
	Pos Pos
}

type Column struct {
	Pos Pos
}

type Whitespace struct {
	Value string
	Pos   Pos
}

type CDO struct {
	Pos Pos
}
type CDC struct {
	Pos Pos
}

type Colon struct {
	Pos Pos
}
type Semicolon struct {
	Pos Pos
}
type Comma struct {
	Pos Pos
}
type LBrack struct {
	Pos Pos
}
type RBrack struct {
	Pos Pos
}
type LParen struct {
	Pos Pos
}
type RParen struct {
	Pos Pos
}
type LBrace struct {
	Pos Pos
}
type RBrace struct {
	Pos Pos
}

type EOF struct{}

// Pos specifies the line and character position of a token.
// The Char and Line are both zero-based indexes.
type Pos struct {
	Char int
	Line int
}
