package css

import (
	"fmt"
	"strings"
)

// Parser represents a CSS3 parser.
type Parser struct {
	Errors ErrorList
}

// ParseStyleSheet parses an input stream into a stylesheet.
func (p *Parser) ParseStyleSheet(s *Scanner) *StyleSheet {
	ss := &StyleSheet{}
	ss.Rules = p.consumeRules(s, true)
	return ss
}

// ParseRule parses a list of rules.
func (p *Parser) ParseRules(s *Scanner) Rules {
	return p.consumeRules(s, false)
}

// ParseRule parses a qualified rule or at-rule.
func (p *Parser) ParseRule(s *Scanner) Rule {
	var r Rule

	// Skip over initial whitespace.
	p.skipWhitespace(&scanner{s})

	// If the next token is EOF, return syntax error.
	// If the next token is at-keyword, consume an at-rule.
	// Otherwise consume a qualified rule. If nothing is returned, return error.
	tok := s.Scan()
	if tok.Tok == EOFToken {
		p.Errors = append(p.Errors, &Error{Message: "unexpected EOF", Pos: Position(s.current())})
		return nil, p.error()
	} else if tok.Tok == AtKeywordToken {
		r = p.consumeAtRule(s)
	} else {
		s.unscan()
		r = p.consumeQualifiedRule(&scanner{s})
	}

	// Skip over trailing whitespace.
	p.skipWhitespace(&scanner{s})

	if tok := s.Scan(); tok.Tok != EOFToken {
		p.Errors = append(p.Errors, &Error{Message: fmt.Sprintf("expected EOF, got %s", print(s.current())), Pos: Position(s.current())})
		return nil, p.error()
	}

	return r, p.error()
}

// ParseDeclaration parses a name/value declaration.
func (p *Parser) ParseDeclaration(s *Scanner) *Declaration {
	// Skip over initial whitespace.
	p.skipWhitespace(&scanner{s})

	// If the next token is not an ident then return an error.
	if tok := s.Scan(); tok.Tok != IdentToken {
		p.Errors = append(p.Errors, &Error{Message: fmt.Sprintf("expected ident, got %s", print(s.current())), Pos: Position(s.current())})
		return nil, p.error()
	}
	s.unscan()

	// Consume a declaration.
	return p.consumeDeclaration(&scanner{s})
}

// ParseDeclarations parses a list of declarations and at-rules.
func (p *Parser) ParseDeclarations(s *Scanner) Declarations {
	return p.consumeDeclarations(s)
}

// ParseComponentValue parses a component value.
func (p *Parser) ParseComponentValue(s *Scanner) ComponentValue {
	// Skip over initial whitespace.
	p.skipWhitespace(&scanner{s})

	// If the next token is EOF then return an error.
	if tok := s.Scan(); tok.Tok == EOFToken {
		p.Errors = append(p.Errors, &Error{Message: "unexpected EOF", Pos: Position(s.current())})
		return nil
	}
	s.unscan()

	// Consume component value.
	v := p.consumeComponentValue(&scanner{s})

	// Skip over any trailing whitespace.
	p.skipWhitespace(&scanner{s})

	// If we're not at EOF then return a syntax error.
	if tok := s.Scan(); tok.Tok != EOFToken {
		s.unscan()
		p.Errors = append(p.Errors, &Error{Message: fmt.Sprintf("expected EOF, got %s", print(s.current())), Pos: Position(s.current())})
		return nil
	}

	return v
}

// ParseComponentValues parses a list of component values.
func (p *Parser) ParseComponentValues(s *Scanner) ComponentValues {
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

// ConsumeRules consumes a list of rules from a token stream.
func (p *Parser) ConsumeRules(s ComponentValueScanner, topLevel bool) Rules {
	var a Rules
	for {
		tok := s.Scan()
		switch tok := tok.(type) {
		case *Token:
			switch tok.Tok {
			case WhitespaceToken:
				continue // nop
			case EOFToken:
				return a
			case CDOToken, CDCToken:
				if !topLevel {
					s.unscan()
					if r := p.ConsumeQualifiedRule(s); r != nil {
						a = append(a, r)
					}
					continue
				}
			case AtKeywordToken:
				if r := p.ConsumeAtRule(s); r != nil {
					a = append(a, r)
				}
				continue
			}
		}

		// Otherwise consume a qualified rule.
		s.Unscan()
		if r := p.ConsumeQualifiedRule(s); r != nil {
			a = append(a, r)
		}
	}
}

// ConsumeAtRule consumes a single at-rule.
func (p *Parser) ConsumeAtRule(s ComponentValueScanner) *AtRule {
	var r AtRule

	// Set the name to the value of the current token.
	// TODO(benbjohnson): Validate first token.
	r.Name = s.Current().(*Token).Value

	// Repeatedly consume the next token.
	for {
		tok := s.Scan()
		switch tok := tok.(type) {
		case *Token:
			switch tok.Tok {
			case SemicolonToken, EOFToken:
				return &r
			case LBraceToken:
				r.Block = p.ConsumeSimpleBlock(s)
				return &r
			}
		case *SimpleBlock:
			if tok.Token.Tok == LBraceToken {
				r.Block = p.ConsumeSimpleBlock(s)
				return &r
			}
		}

		// Otherwise consume a component value.
		s.unscan()
		v := p.ConsumeComponentValue(&scanner{s})
		r.Prelude = append(r.Prelude, v)
	}
}

// ConsumeQualifiedRule consumes a single qualified rule.
func (p *parser) ConsumeQualifiedRule(s ComponentValueScanner) *QualifiedRule {
	var r QualifiedRule

	// Repeatedly consume the next token.
	for {
		tok := s.Scan()
		switch tok := tok.(type) {
		case *Token:
			switch tok.Tok {
			case EOFToken:
				p.Errors = append(p.Errors, &Error{Message: "unexpected EOF", Pos: tok.Pos})
				return nil
			case LBraceToken:
				r.Block = p.ConsumeSimpleBlock(s)
				return &r
			}
		case *SimpleBlock:
			if tok.Token.Tok == LBraceToken {
				r.Block = p.ConsumeSimpleBlock(s)
				return &r
			}
		}
		s.Unscan()
		r.Prelude = append(r.Prelude, p.ConsumeComponentValue(s))
	}
}

// ConsumeDeclarations consumes a list of declarations.
func (p *Parser) ConsumeDeclarations(s ComponentValueScanner) Declarations {
	var a Declarations

	// Repeatedly consume the next token.
	for {
		tok := s.Scan()
		switch tok.Tok {
		case WhitespaceToken, SemicolonToken:
			// nop
			continue
		case EOFToken:
			return a
		case AtKeywordToken:
			a = append(a, p.ConsumeAtRule(s))
			continue
		case IdentToken:
			// Generate a list of tokens up to the next semicolon or EOF.
			s.unscan()
			values := p.ConsumeDeclarationValues(s)

			// Consume declaration using temporary list of tokens.
			if d := p.ConsumeDeclaration(&componentValueList{i: -1, values: values}); d != nil {
				a = append(a, d)
			}
			continue
		}

		// Any other token is a syntax error.
		p.Errors = append(p.Errors, &Error{Message: fmt.Sprintf("unexpected: %s", print(tok)), Pos: Position(tok)})

		// Repeatedly consume a component values until semicolon or EOF.
		p.skipComponentValues(&scanner{s})
	}
}

// ConsumeDeclaration consumes a single declaration.
func (p *Parser) ConsumeDeclaration(s ComponentValueScanner) *Declaration {
	var d Declaration

	// The first token must be an ident.
	// TODO(benbjohnson): Validate initial token.
	d.Name = s.Scan().(*Token).Value

	// Skip over whitespace.
	p.skipWhitespace(s)

	// The next token must be a colon.
	if tok := s.Scan().(*Token); tok.Tok != ColonToken {
		p.Errors = append(p.Errors, &Error{Message: fmt.Sprintf("expected colon, got %s", print(s.Current())), Pos: Position(s.Current())})
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

// ConsumeComponentValue consumes a single component value. (§5.4.6)
func (p *Parser) ConsumeComponentValue(s ComponentValueScanner) ComponentValue {
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

// ConsumeSimpleBlock consumes a simple block. (§5.4.7)
func (p *Parser) ConsumeSimpleBlock(s ComponentValueScanner) *SimpleBlock {
	b := &SimpleBlock{}

	// Set the block's associated token to the current token.
	// TODO(benbjohnson): Validate first token.
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
		b.Values = append(b.Values, p.ConsumeComponentValue(s))
	}
}

// ConsumeFunction consumes a function.
func (p *Parser) ConsumeFunction(s ComponentValueScanner) *Function {
	f := &Function{}

	// Set the name to the first token.
	// TODO(benbjohnson): Validate first token.
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
func (p *parser) skipComponentValues(s ComponentValueScanner) {
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
func (p *parser) skipWhitespace(s ComponentValueScanner) {
	for {
		if tok, ok := s.Scan().(*Token); ok && tok.Tok != WhitespaceToken {
			s.Unscan()
			return
		}
	}
}

// ComponentValueScanner represents a type that can retrieve the next component value.
type ComponentValueScanner interface {
	Current() ComponentValue
	Scan() ComponentValue
	Unscan()
}

// componentValueList represents a scanner for a fixed list of component values.
type componentValueList struct {
	i      int
	values ComponentValues
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
