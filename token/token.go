package token

import "fmt"

// Token represents a lexical token.
type Token interface {
	token()
	Position() Pos
	String() string
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

func (t *Ident) String() string { return t.Value }
func (t *Ident) Position() Pos  { return t.Pos }

type Function struct {
	Type  string
	Value string
	Pos   Pos
}

func (t *Function) String() string { return t.Value + "(" }
func (t *Function) Position() Pos  { return t.Pos }

type AtKeyword struct {
	Type  string
	Value string
	Pos   Pos
}

func (t *AtKeyword) String() string { return "@" + t.Value }
func (t *AtKeyword) Position() Pos  { return t.Pos }

type Hash struct {
	Type  string
	Value string
	Pos   Pos
}

func (t *Hash) String() string { return "#" + t.Value }
func (t *Hash) Position() Pos  { return t.Pos }

type String struct {
	Type   string
	Ending rune
	Value  string
	Pos    Pos
}

func (t *String) String() string { return string(t.Ending) + t.Value + string(t.Ending) }
func (t *String) Position() Pos  { return t.Pos }

type BadString struct{ Pos Pos }

func (_ *BadString) String() string { return "bad-string" }
func (t *BadString) Position() Pos  { return t.Pos }

type URL struct {
	Type  string
	Value string
	Pos   Pos
}

func (t *URL) String() string { return "url(" + t.Value + ")" }
func (t *URL) Position() Pos  { return t.Pos }

type BadURL struct{ Pos Pos }

func (t *BadURL) String() string { return "bad-url" }
func (t *BadURL) Position() Pos  { return t.Pos }

type Delim struct {
	Value string
	Pos   Pos
}

func (t *Delim) String() string { return t.Value }
func (t *Delim) Position() Pos  { return t.Pos }

type Number struct {
	Type   string
	Number float64
	Value  string
	Pos    Pos
}

func (t *Number) String() string { return t.Value }
func (t *Number) Position() Pos  { return t.Pos }

type Percentage struct {
	Type   string
	Number float64
	Value  string
	Pos    Pos
}

func (t *Percentage) String() string { return t.Value }
func (t *Percentage) Position() Pos  { return t.Pos }

type Dimension struct {
	Type   string
	Number float64
	Unit   string
	Value  string
	Pos    Pos
}

func (t *Dimension) String() string { return t.Value }
func (t *Dimension) Position() Pos  { return t.Pos }

type UnicodeRange struct {
	Start int
	End   int
	Pos   Pos
}

func (t *UnicodeRange) String() string { return fmt.Sprintf("U+%06x-U+%06x", t.Start, t.End) }
func (t *UnicodeRange) Position() Pos  { return t.Pos }

type IncludeMatch struct {
	Pos Pos
}

func (_ *IncludeMatch) String() string { return "~=" }
func (t *IncludeMatch) Position() Pos  { return t.Pos }

type DashMatch struct{ Pos Pos }

func (_ *DashMatch) String() string { return "|=" }
func (t *DashMatch) Position() Pos  { return t.Pos }

type PrefixMatch struct{ Pos Pos }

func (_ *PrefixMatch) String() string { return "^=" }
func (t *PrefixMatch) Position() Pos  { return t.Pos }

type SuffixMatch struct{ Pos Pos }

func (_ *SuffixMatch) String() string { return "$=" }
func (t *SuffixMatch) Position() Pos  { return t.Pos }

type SubstringMatch struct{ Pos Pos }

func (_ *SubstringMatch) String() string { return "*=" }
func (t *SubstringMatch) Position() Pos  { return t.Pos }

type Column struct{ Pos Pos }

func (_ *Column) String() string { return "||" }
func (t *Column) Position() Pos  { return t.Pos }

type Whitespace struct {
	Value string
	Pos   Pos
}

func (t *Whitespace) String() string { return t.Value }
func (t *Whitespace) Position() Pos  { return t.Pos }

type CDO struct{ Pos Pos }

func (_ *CDO) String() string { return "<!--" }
func (t *CDO) Position() Pos  { return t.Pos }

type CDC struct{ Pos Pos }

func (_ *CDC) String() string { return "-->" }
func (t *CDC) Position() Pos  { return t.Pos }

type Colon struct{ Pos Pos }

func (_ *Colon) String() string { return ":" }
func (t *Colon) Position() Pos  { return t.Pos }

type Semicolon struct{ Pos Pos }

func (_ *Semicolon) String() string { return ";" }
func (t *Semicolon) Position() Pos  { return t.Pos }

type Comma struct{ Pos Pos }

func (_ *Comma) String() string { return "," }
func (t *Comma) Position() Pos  { return t.Pos }

type LBrack struct{ Pos Pos }

func (_ *LBrack) String() string { return "[" }
func (t *LBrack) Position() Pos  { return t.Pos }

type RBrack struct{ Pos Pos }

func (_ *RBrack) String() string { return "]" }
func (t *RBrack) Position() Pos  { return t.Pos }

type LParen struct{ Pos Pos }

func (_ *LParen) String() string { return "(" }
func (t *LParen) Position() Pos  { return t.Pos }

type RParen struct{ Pos Pos }

func (_ *RParen) String() string { return ")" }
func (t *RParen) Position() Pos  { return t.Pos }

type LBrace struct{ Pos Pos }

func (_ *LBrace) String() string { return "{" }
func (t *LBrace) Position() Pos  { return t.Pos }

type RBrace struct{ Pos Pos }

func (_ *RBrace) String() string { return "}" }
func (t *RBrace) Position() Pos  { return t.Pos }

type EOF struct{ Pos Pos }

func (_ *EOF) String() string { return "EOF" }
func (t *EOF) Position() Pos  { return t.Pos }

// Pos specifies the line and character position of a token.
// The Char and Line are both zero-based indexes.
type Pos struct {
	Char int
	Line int
}
