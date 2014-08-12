package css

import "fmt"

// Node represents a node in the CSS3 abstract syntax tree.
type Node interface {
	node()
}

func (_ *StyleSheet) node()     {}
func (_ Rules) node()           {}
func (_ *AtRule) node()         {}
func (_ *QualifiedRule) node()  {}
func (_ Declarations) node()    {}
func (_ *Declaration) node()    {}
func (_ ComponentValues) node() {}
func (_ *SimpleBlock) node()    {}
func (_ *Function) node()       {}
func (_ *Token) node()          {}

// StyleSheet represents a top-level CSS3 stylesheet.
type StyleSheet struct {
	Rules Rules
}

// Rules represents a list of rules.
type Rules []Rule

// Rule represents a qualified rule or at-rule.
type Rule interface {
	Node
	rule()
}

func (_ *AtRule) rule()        {}
func (_ *QualifiedRule) rule() {}

// AtRule represents a rule starting with an "@" symbol.
type AtRule struct {
	Name    string
	Prelude ComponentValues
	Block   *SimpleBlock
	Pos     Pos
}

// QualifiedRule represents an unnamed rule that includes a prelude and block.
type QualifiedRule struct {
	Prelude ComponentValues
	Block   *SimpleBlock
	Pos     Pos
}

// Declarations represents a list of declarations or at-rules.
type Declarations []Node

// Declaration represents a name/value pair.
type Declaration struct {
	Name      string
	Values    ComponentValues
	Important bool
	Pos       Pos
}

// ComponentValues represents a list of component values.
type ComponentValues []ComponentValue

// nonwhitespace returns the list of values without whitespace characters.
func (a ComponentValues) nonwhitespace() ComponentValues {
	var tmp ComponentValues
	for _, v := range a {
		if v, ok := v.(*Token); ok && v.Tok == WhitespaceToken {
			continue
		}
		tmp = append(tmp, v)
	}
	return tmp
}

// ComponentValue represents a component value.
type ComponentValue interface {
	Node
	componentValue()
}

func (_ *SimpleBlock) componentValue() {}
func (_ *Function) componentValue()    {}
func (_ *Token) componentValue()       {}

// SimpleBlock represents a {-block, [-block, or (-block.
type SimpleBlock struct {
	Token  *Token
	Values ComponentValues
	Pos    Pos
}

// Function represents a function call with a list of arguments.
type Function struct {
	Name   string
	Values ComponentValues
	Pos    Pos
}

// Token represents a lexical token.
type Token struct {
	// The type of token.
	Tok Tok

	// A flag set for ident-like tokens to either "id" or "unrestricted".
	// Also set for numeric tokens to either "integer" or "number"
	Type string

	// The literal value of the token as parsed.
	Value string

	// The rune used to close the token. Used for string tokens.
	Ending rune

	// The numeric value and unit used for numeric tokens.
	Number float64
	Unit   string

	// Beginning and ending range for a unicode-range token.
	Start int
	End   int

	// Position of the token in the source document.
	Pos Pos
}

// Tok represents a lexical token type.
type Tok int

const (
	IdentToken Tok = iota + 1
	FunctionToken
	AtKeywordToken
	HashToken
	StringToken
	BadStringToken
	URLToken
	BadURLToken
	DelimToken
	NumberToken
	PercentageToken
	DimensionToken
	UnicodeRangeToken
	IncludeMatchToken
	DashMatchToken
	PrefixMatchToken
	SuffixMatchToken
	SubstringMatchToken
	ColumnToken
	WhitespaceToken
	CDOToken
	CDCToken
	ColonToken
	SemicolonToken
	CommaToken
	LBrackToken
	RBrackToken
	LParenToken
	RParenToken
	LBraceToken
	RBraceToken
	EOFToken
)

// Pos specifies the line and character position of a token.
// The Char and Line are both zero-based indexes.
type Pos struct {
	Char int
	Line int
}

// Position returns the position for a given Node.
func Position(n Node) Pos {
	switch n := n.(type) {
	case *StyleSheet:
		return Position(n.Rules)
	case Rules:
		if len(n) > 0 {
			return Position(n[0])
		}
	case *AtRule:
		return n.Pos
	case *QualifiedRule:
		return n.Pos
	case Declarations:
		if len(n) > 0 {
			return Position(n[0])
		}
	case *Declaration:
		return n.Pos
	case ComponentValues:
		if len(n) > 0 {
			return Position(n[0])
		}
	case *SimpleBlock:
		return n.Pos
	case *Function:
		return n.Pos
	case *Token:
		return n.Pos
	}
	return Pos{}
}

// Error represents a syntax error.
type Error struct {
	Message string
	Pos     Pos
}

// Error returns the formatted string error message.
func (e *Error) Error() string {
	return e.Message
}

// ErrorList represents a list of syntax errors.
type ErrorList []error

// Error returns the formatted string error message.
func (a ErrorList) Error() string {
	switch len(a) {
	case 0:
		return "no errors"
	case 1:
		return a[0].Error()
	}
	return fmt.Sprintf("%s (and %d more errors)", a[0], len(a)-1)
}
