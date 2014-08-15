package css

import (
	"fmt"
	"strings"
)

// parser represents a CSS3 parser.
type parser struct {
	errors ErrorList
}

// ParseStyleSheet parses an input stream into a stylesheet.
func ParseStyleSheet(s *Scanner) (*StyleSheet, error) {
	var p parser
	ss := &StyleSheet{}
	ss.Rules = p.consumeRules(&scanner{s}, true)
	return ss, p.error()
}

// ParseRule parses a list of rules.
func ParseRules(s *Scanner) (Rules, error) {
	var p parser
	a := p.consumeRules(&scanner{s}, false)
	return a, p.error()
}

// ParseRule parses a qualified rule or at-rule.
func ParseRule(s *Scanner) (Rule, error) {
	var p parser
	var r Rule

	// Skip over initial whitespace.
	p.skipWhitespace(&scanner{s})

	// If the next token is EOF, return syntax error.
	// If the next token is at-keyword, consume an at-rule.
	// Otherwise consume a qualified rule. If nothing is returned, return error.
	tok := s.Scan()
	if tok.Tok == EOFToken {
		p.errors = append(p.errors, &Error{Message: "unexpected EOF", Pos: Position(s.Current())})
		return nil, p.error()
	} else if tok.Tok == AtKeywordToken {
		r = p.consumeAtRule(&scanner{s})
	} else {
		s.Unscan()
		r = p.consumeQualifiedRule(&scanner{s})
	}

	// Skip over trailing whitespace.
	p.skipWhitespace(&scanner{s})

	if tok := s.Scan(); tok.Tok != EOFToken {
		p.errors = append(p.errors, &Error{Message: fmt.Sprintf("expected EOF, got %s", print(s.Current())), Pos: Position(s.Current())})
		return nil, p.error()
	}

	return r, p.error()
}

// ParseDeclaration parses a name/value declaration.
func ParseDeclaration(s *Scanner) (*Declaration, error) {
	var p parser

	// Skip over initial whitespace.
	p.skipWhitespace(&scanner{s})

	// If the next token is not an ident then return an error.
	if tok := s.Scan(); tok.Tok != IdentToken {
		p.errors = append(p.errors, &Error{Message: fmt.Sprintf("expected ident, got %s", print(s.Current())), Pos: Position(s.Current())})
		return nil, p.error()
	}
	s.Unscan()

	// Consume a declaration.
	d := p.consumeDeclaration(&scanner{s})

	return d, p.error()
}

// ParseDeclarations parses a list of declarations and at-rules.
func ParseDeclarations(s *Scanner) (Declarations, error) {
	var p parser
	a := p.consumeDeclarations(&scanner{s})
	return a, p.error()
}

// ParseComponentValue parses a component value.
func ParseComponentValue(s *Scanner) (ComponentValue, error) {
	var p parser

	// Skip over initial whitespace.
	p.skipWhitespace(&scanner{s})

	// If the next token is EOF then return an error.
	if tok := s.Scan(); tok.Tok == EOFToken {
		p.errors = append(p.errors, &Error{Message: "unexpected EOF", Pos: Position(s.Current())})
		return nil, p.error()
	}
	s.Unscan()

	// Consume component value.
	v := p.consumeComponentValue(&scanner{s})

	// Skip over any trailing whitespace.
	p.skipWhitespace(&scanner{s})

	// If we're not at EOF then return a syntax error.
	if tok := s.Scan(); tok.Tok != EOFToken {
		s.Unscan()
		p.errors = append(p.errors, &Error{Message: fmt.Sprintf("expected EOF, got %s", print(s.Current())), Pos: Position(s.Current())})
		return nil, p.error()
	}

	return v, nil
}

// ParseComponentValues parses a list of component values.
func ParseComponentValues(s *Scanner) (ComponentValues, error) {
	var a ComponentValues

	// Repeatedly consume a component value until EOF.
	var p parser
	for {
		v := p.consumeComponentValue(&scanner{s})

		// If the value is an EOF, then exit.
		if tok, ok := v.(*Token); ok && tok.Tok == EOFToken {
			break
		}

		// Otherwise append to list of component values.
		a = append(a, v)
	}

	return a, nil
}

// Errors returns the error on the parser.
// Returns nil if there are no errors.
func (p *parser) error() error {
	if len(p.errors) == 0 {
		return nil
	}
	return p.errors
}

// consumeRules consumes a list of rules from a token stream. (§5.4.1)
func (p *parser) consumeRules(s componentValueScanner, toplevel bool) Rules {
	var a Rules
	for {
		tok := s.Scan()
		switch tok := tok.(type) {
		case *Token:
			switch tok.Tok {
			case WhitespaceToken:
				// nop
			case EOFToken:
				return a
			case CDOToken, CDCToken:
				if !toplevel {
					s.Unscan()
					if r := p.consumeQualifiedRule(s); r != nil {
						a = append(a, r)
					}
				}
			case AtKeywordToken:
				if r := p.consumeAtRule(s); r != nil {
					a = append(a, r)
				}
			default:
				s.Unscan()
				if r := p.consumeQualifiedRule(s); r != nil {
					a = append(a, r)
				}
			}
		default:
			s.Unscan()
			if r := p.consumeQualifiedRule(s); r != nil {
				a = append(a, r)
			}
		}
	}
}

// consumeAtRule consumes a single at-rule. (§5.4.2)
func (p *parser) consumeAtRule(s componentValueScanner) *AtRule {
	r := &AtRule{}

	// Set the name to the value of the current token.
	r.Name = s.Current().(*Token).Value

	// Repeatedly consume the next token.
	for {
		tok := s.Scan()
		switch tok := tok.(type) {
		case *Token:
			switch tok.Tok {
			case SemicolonToken, EOFToken:
				return r
			case LBraceToken:
				r.Block = p.consumeSimpleBlock(s)
				return r
			}
		case *SimpleBlock:
			if tok.Token.Tok == LBraceToken {
				r.Block = p.consumeSimpleBlock(s)
				return r
			}
		}
		s.Unscan()
		v := p.consumeComponentValue(s)
		r.Prelude = append(r.Prelude, v)
	}
}

// consumeAtRule consumes a single qualified rule. (§5.4.3)
func (p *parser) consumeQualifiedRule(s componentValueScanner) *QualifiedRule {
	r := &QualifiedRule{}

	// Repeatedly consume the next token.
	for {
		tok := s.Scan()
		switch tok := tok.(type) {
		case *Token:
			switch tok.Tok {
			case EOFToken:
				p.errors = append(p.errors, &Error{Message: "unexpected EOF", Pos: tok.Pos})
				return nil
			case LBraceToken:
				r.Block = p.consumeSimpleBlock(s)
				return r
			}
		case *SimpleBlock:
			if tok.Token.Tok == LBraceToken {
				r.Block = p.consumeSimpleBlock(s)
				return r
			}
		}
		s.Unscan()
		r.Prelude = append(r.Prelude, p.consumeComponentValue(s))
	}
}

// consumeDeclarations consumes a list of declarations. (§5.4.4)
func (p *parser) consumeDeclarations(s componentValueScanner) Declarations {
	var a Declarations

	// Repeatedly consume the next token.
	for {
		tok := s.Scan()
		switch tok := tok.(type) {
		case *Token:
			switch tok.Tok {
			case WhitespaceToken, SemicolonToken:
				// nop
				continue
			case EOFToken:
				return a
			case AtKeywordToken:
				a = append(a, p.consumeAtRule(s))
				continue
			case IdentToken:
				// Generate a list of tokens up to the next semicolon or EOF.
				s.Unscan()
				values := p.consumeDeclarationValues(s)

				// Consume declaration using temporary list of tokens.
				if d := p.consumeDeclaration(newComponentValueList(values)); d != nil {
					a = append(a, d)
				}
				continue
			}
		}

		// Any other token is a syntax error.
		p.errors = append(p.errors, &Error{Message: fmt.Sprintf("unexpected: %s", print(tok)), Pos: Position(tok)})

		// Repeatedly consume a component values until semicolon or EOF.
		p.skipComponentValues(s)
	}
}

// consumeDeclaration consumes a single declaration. (§5.4.5)
func (p *parser) consumeDeclaration(s componentValueScanner) *Declaration {
	d := &Declaration{}

	// The first token must be an ident.
	d.Name = s.Scan().(*Token).Value

	// Skip over whitespace.
	p.skipWhitespace(s)

	// The next token must be a colon.
	if tok := s.Scan().(*Token); tok.Tok != ColonToken {
		p.errors = append(p.errors, &Error{Message: fmt.Sprintf("expected colon, got %s", print(s.Current())), Pos: Position(s.Current())})
		return nil
	}

	// Consume the declaration value until EOF.
	for {
		tok := s.Scan()
		if tok, ok := tok.(*Token); ok && tok.Tok == EOFToken {
			break
		}
		d.Values = append(d.Values, tok)
	}

	// Check last two non-whitespace tokens for "!important".
	d.Values, d.Important = cleanImportantFlag(d.Values)

	return d
}

// Checks if the last two non-whitespace tokens are a case-insensitive "!important".
// If so, it removes them and returns the "important" flag set to true.
func cleanImportantFlag(values ComponentValues) (ComponentValues, bool) {
	a := values.nonwhitespace()
	if len(a) < 2 {
		return values, false
	}

	// Check last two tokens for "!important".
	if tok, ok := a[len(a)-2].(*Token); !ok || tok.Tok != DelimToken || tok.Value != "!" {
		return values, false
	}
	if tok, ok := a[len(a)-1].(*Token); !ok || tok.Tok != IdentToken || strings.ToLower(tok.Value) != "important" {
		return values, false
	}

	// Trim "!important" tokens off values.
	for i, v := range values {
		if v == a[len(a)-2] {
			values = values[i:]
			break
		}
	}

	return values, true
}

// consumeComponentValue consumes a single component value. (§5.4.6)
func (p *parser) consumeComponentValue(s componentValueScanner) ComponentValue {
	tok := s.Scan()
	if tok, ok := tok.(*Token); ok {
		switch tok.Tok {
		case LBraceToken, LBrackToken, LParenToken:
			return p.consumeSimpleBlock(s)
		case FunctionToken:
			return p.consumeFunction(s)
		}
	}
	return tok
}

// consumeSimpleBlock consumes a simple block. (§5.4.7)
func (p *parser) consumeSimpleBlock(s componentValueScanner) *SimpleBlock {
	b := &SimpleBlock{}

	// Set the block's associated token to the current token.
	b.Token = s.Current().(*Token)

	for {
		tok := s.Scan()

		// If this token is EOF or the mirror of the starting token then return.
		if tok, ok := tok.(*Token); ok {
			switch tok.Tok {
			case EOFToken:
				return b
			case RBrackToken:
				if b.Token.Tok == LBrackToken {
					return b
				}
			case RBraceToken:
				if b.Token.Tok == LBraceToken {
					return b
				}
			case RParenToken:
				if b.Token.Tok == LParenToken {
					return b
				}
			}
		}

		// Otherwise consume a component value.
		s.Unscan()
		b.Values = append(b.Values, p.consumeComponentValue(s))
	}
}

// consumeFunction consumes a function. (§5.4.8)
func (p *parser) consumeFunction(s componentValueScanner) *Function {
	f := &Function{}

	// Set the name to the first token.
	f.Name = s.Current().(*Token).Value

	for {
		tok := s.Scan()

		// If this token is EOF or the mirror of the starting token then return.
		if tok, ok := tok.(*Token); ok && (tok.Tok == EOFToken || tok.Tok == RParenToken) {
			return f
		}

		// Otherwise consume a component value.
		s.Unscan()
		f.Values = append(f.Values, p.consumeComponentValue(s))
	}
}

// consumeDeclarationTokens collects contiguous non-semicolon and non-EOF tokens.
func (p *parser) consumeDeclarationValues(s componentValueScanner) ComponentValues {
	var a ComponentValues
	for {
		tok := s.Scan()
		if tok, ok := tok.(*Token); ok && (tok.Tok == SemicolonToken || tok.Tok == EOFToken) {
			s.Unscan()
			return a
		}
		a = append(a, tok)
	}
}

// skipComponentValues consumes all component values until a semicolon or EOF.
func (p *parser) skipComponentValues(s componentValueScanner) {
	for {
		v := p.consumeComponentValue(s)
		if tok, ok := v.(*Token); ok {
			switch tok.Tok {
			case SemicolonToken, EOFToken:
				return
			}
		}
	}
}

// skipWhitespace skips over all contiguous whitespace tokes.
func (p *parser) skipWhitespace(s componentValueScanner) {
	for {
		if tok, ok := s.Scan().(*Token); ok && tok.Tok != WhitespaceToken {
			s.Unscan()
			return
		}
	}
}

// componentValueScanner represents a type that can retrieve the next component value.
type componentValueScanner interface {
	Current() ComponentValue
	Scan() ComponentValue
	Unscan()
}

// componentValueList represents a scanner for a fixed list of component values.
type componentValueList struct {
	i      int
	values ComponentValues
}

// newComponentValueList returns a new instance of componentValueList.
func newComponentValueList(a ComponentValues) *componentValueList {
	return &componentValueList{i: -1, values: a}
}

// Current returns the current component value.
func (l *componentValueList) Current() ComponentValue {
	if l.i >= len(l.values) {
		return &Token{Tok: EOFToken}
	}
	return l.values[l.i]
}

// Scan returns the next component value.
func (l *componentValueList) Scan() ComponentValue {
	if l.i < len(l.values) {
		l.i++
	}
	return l.Current()
}

// Unscan moves back one component value.
func (l *componentValueList) Unscan() {
	if l.i > -1 {
		l.i--
	}
}
