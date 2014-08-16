package css

import (
	"reflect"
	"testing"
)

// Ensure that all nodes implement the Node interface.
func TestNode(t *testing.T) {
	var a []Node
	a = append(a, &StyleSheet{}, &AtRule{}, &QualifiedRule{}, &Declaration{})
	a = append(a, &SimpleBlock{}, &Function{}, &Token{})
	a = append(a, Rules{}, Declarations{}, ComponentValues{})
	for _, n := range a {
		n.node()
	}
}

// Ensure that all rules implement the Rule interface.
func TestRule(t *testing.T) {
	a := []Rule{&AtRule{}, &QualifiedRule{}}
	for _, r := range a {
		r.rule()
	}
}

// Ensure that all component values implement the ComponentValue interface.
func TestComponentValue(t *testing.T) {
	a := []ComponentValue{&SimpleBlock{}, &Function{}, &Token{}}
	for _, v := range a {
		v.componentValue()
	}
}

// Ensure that node positions can be retrieved.
func TestPosition(t *testing.T) {
	var tests = []struct {
		in  Node
		pos Pos
	}{
		{in: &StyleSheet{Rules: Rules{&QualifiedRule{Pos: Pos{1, 2}}}}, pos: Pos{1, 2}},
		{in: Rules{&AtRule{Pos: Pos{1, 2}}}, pos: Pos{1, 2}},
		{in: Rules{}, pos: Pos{}},
		{in: &QualifiedRule{Pos: Pos{1, 2}}, pos: Pos{1, 2}},
		{in: &AtRule{Pos: Pos{1, 2}}, pos: Pos{1, 2}},
		{in: Declarations{&AtRule{Pos: Pos{1, 2}}}, pos: Pos{1, 2}},
		{in: Declarations{&Declaration{Pos: Pos{1, 2}}}, pos: Pos{1, 2}},
		{in: Declarations{}, pos: Pos{}},
		{in: ComponentValues{&SimpleBlock{Pos: Pos{1, 2}}}, pos: Pos{1, 2}},
		{in: ComponentValues{&Function{Pos: Pos{1, 2}}}, pos: Pos{1, 2}},
		{in: ComponentValues{&Token{Pos: Pos{1, 2}}}, pos: Pos{1, 2}},
		{in: ComponentValues{}, pos: Pos{}},
		{in: &SimpleBlock{Pos: Pos{1, 2}}, pos: Pos{1, 2}},
		{in: &Function{Pos: Pos{1, 2}}, pos: Pos{1, 2}},
		{in: &Token{Pos: Pos{1, 2}}, pos: Pos{1, 2}},
	}

	for _, tt := range tests {
		if pos := Position(tt.in); !reflect.DeepEqual(tt.pos, pos) {
			t.Errorf("expected: %#v, got: %#v", tt.pos, pos)
		}
	}
}

// Ensure that an error list can be properly formatted.
func TestErrorList_Error(t *testing.T) {
	var tests = []struct {
		in ErrorList
		s  string
	}{
		{in: nil, s: "no errors"},
		{in: ErrorList{}, s: "no errors"},
		{in: ErrorList{&Error{Message: "foo"}}, s: "foo"},
		{in: ErrorList{&Error{Message: "foo"}, &Error{Message: "bar"}}, s: "foo (and 1 more errors)"},
	}

	for _, tt := range tests {
		if s := tt.in.Error(); tt.s != s {
			t.Errorf("expected: %s, got: %s", tt.s, s)
		}
	}

}

// TODO(benbjohnson): TestPosition_*
