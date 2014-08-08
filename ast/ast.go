package ast

import (
	"bytes"
	"fmt"

	"github.com/benbjohnson/css/token"
)

// Node represents a node in the CSS3 abstract syntax tree.
type Node interface {
	node()
	String() string
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

func (s *StyleSheet) String() string {
	var buf bytes.Buffer
	for _, r := range s.Rules {
		buf.WriteString(r.String())
		buf.WriteString("\n")
	}
	return buf.String()
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
}

func (r *AtRule) String() string {
	var buf bytes.Buffer
	buf.WriteString("@" + r.Name)
	if len(r.Prelude) > 0 {
		buf.WriteString(" " + r.Prelude.String())
	}
	if r.Block != nil {
		buf.WriteString(" " + r.Block.String())
	} else {
		buf.WriteString(";")
	}
	return buf.String()
}

// QualifiedRule represents an unnamed rule that includes a prelude and block.
type QualifiedRule struct {
	Prelude ComponentValues
	Block   *SimpleBlock
}

func (r *QualifiedRule) String() string {
	return r.Prelude.String() + r.Block.String()
}

// Declarations represents a list of declarations or at-rules.
type Declarations []Node

// Declaration represents a name/value pair.
type Declaration struct {
	Name   string
	Values ComponentValues
}

func (d *Declaration) String() string {
	return d.Name + ": " + d.Values.String()
}

// ComponentValues represents a list of component values.
type ComponentValues []ComponentValue

func (a ComponentValues) String() string {
	var buf bytes.Buffer
	for _, v := range a {
		buf.WriteString(v.String())
	}
	return buf.String()
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
	Token  token.Token
	Values ComponentValues
}

func (b *SimpleBlock) String() string {
	switch b.Token.(type) {
	case *token.LBrace:
		return "{" + b.Values.String() + "}"
	case *token.LBrack:
		return "[" + b.Values.String() + "]"
	case *token.LParen:
		return "(" + b.Values.String() + ")"
	}
	return "<>"
}

// Function represents a function call with a list of arguments.
type Function struct {
	Name   string
	Values ComponentValues
}

func (f *Function) String() string {
	return fmt.Sprintf("%s(%s)", f.Name, f.Values.String())
}

// Token represents a single token in the AST.
type Token struct {
	token.Token
}

func (t *Token) String() string {
	return t.Token.String()
}
