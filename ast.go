package css

// Node represents a node in the CSS3 abstract syntax tree.
type Node interface {
	node()
}

func (_ *Stylesheet) node()    {}
func (_ Rules) node()          {}
func (_ *AtRule) node()        {}
func (_ *QualifiedRule) node() {}
func (_ Declarations) node()   {}
func (_ *Declaration) node()   {}

// Stylesheet represents a top-level CSS3 stylesheet.
type Stylesheet struct {
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
	Block   SimpleBlock
}

// QualifiedRule represents an unnamed rule that includes a prelude and block.
type QualifiedRule struct {
	Prelude ComponentValues
	Block   SimpleBlock
}

// Declarations represents a list of declarations.
type Declarations []*Declaration

// Declaration represents a name/value pair.
type Declaration struct {
	Name   string
	Values ComponentValues
}

// SimpleBlock represents a {-block, [-block, or (-block.
type SimpleBlock struct {
	Values ComponentValues
}

// Function represents a function call with a list of arguments.
type Function struct {
	Name   string
	Values ComponentValues
}

// ComponentValues represents a list of component values.
type ComponentValues []*ComponentValue

// ComponentValue represents a component value.
type ComponentValue struct {
}
