package parser

import (
	"fmt"

	"github.com/benbjohnson/css/ast"
	"github.com/benbjohnson/css/token"
)

// parser represents a CSS3 parser.
type parser struct {
	errors ErrorList
}

// ParseStyleSheet parses an input stream into a stylesheet.
func ParseStyleSheet(s Scanner) (*ast.StyleSheet, error) {
	var p parser
	ss := &ast.StyleSheet{}
	ss.Rules = p.consumeRules(s, true)
	return ss, p.error()
}

// ParseRule parses a list of rules.
func ParseRules(s Scanner) (ast.Rules, error) {
	var p parser
	a := p.consumeRules(s, false)
	return a, p.error()
}

// ParseRule parses a qualified rule or at-rule.
func ParseRule(s Scanner) (ast.Rule, error) {
	// TODO(benbjohnson): Consume the next input token.
	// TODO(benbjohnson): While the current token is whitespace, consume.
	// TODO(benbjohnson): If the current token is EOF, return syntax error.
	// TODO(benbjohnson): If the current token is at-keyword, consume at-rule.
	// TODO(benbjohnson): Otherwise consume a qualified rule. If nothing is returned, return syntax error.
	// TODO(benbjohnson): While current token is whitespace, consume.
	// TODO(benbjohnson): If current token is EOF, return rule. Otherwise return syntax error.
	return nil, nil
}

// ParseDeclaration parses a name/value declaration.
func ParseDeclaration(s Scanner) (*ast.Declaration, error) {
	var p parser

	// Skip over initial whitespace.
	p.skipWhitespace(s)

	// If the next token is not an ident then return an error.
	if _, ok := s.Scan().(*token.Ident); !ok {
		p.errors = append(p.errors, &Error{Message: fmt.Sprintf("expected ident, got %q", s.Current().String()), Pos: s.Current().Position()})
		return nil, p.error()
	}
	s.Unscan()

	// Consume a declaration. If nothing is returned, return syntax error.
	d := p.consumeDeclaration(s)
	if d == nil {
		p.errors = append(p.errors, &Error{Message: "expected declaration", Pos: s.Current().Position()})
	}

	return d, p.error()
}

// ParseDeclarations parses a list of declarations and at-rules.
func ParseDeclarations(s Scanner) (*ast.Declaration, error) {
	// TODO(benbjohnson): Consume a list of declarations.
	return nil, nil
}

// ParseComponentValue parses a component value.
func ParseComponentValue(s Scanner) (ast.ComponentValue, error) {
	var p parser

	// Skip over initial whitespace.
	p.skipWhitespace(s)

	// If the next token is EOF then return an error.
	if _, ok := s.Scan().(*token.EOF); ok {
		p.errors = append(p.errors, &Error{Message: "unexpected EOF", Pos: s.Current().Position()})
		return nil, p.error()
	}
	s.Unscan()

	// Consume component value.
	v := p.consumeComponentValue(s)
	if v == nil {
		p.errors = append(p.errors, &Error{Message: "expected component value", Pos: s.Current().Position()})
		return nil, p.error()
	}

	// Skip over any trailing whitespace.
	p.skipWhitespace(s)

	// If we're not at EOF then return a syntax error.
	if _, ok := s.Scan().(*token.EOF); !ok {
		s.Unscan()
		p.errors = append(p.errors, &Error{Message: fmt.Sprintf("expected EOF, got %q", s.Current().String()), Pos: s.Current().Position()})
		return nil, p.error()
	}

	return v, nil
}

// ParseComponentValues parses a list of component values.
func ParseComponentValues(s Scanner) (ast.ComponentValues, error) {
	var a ast.ComponentValues

	// Repeatedly consume a component value until EOF.
	var p parser
	for {
		v := p.consumeComponentValue(s)

		// If the value is an EOF, then exit.
		if v, ok := v.(*ast.Token); ok {
			if _, ok := v.Token.(*token.EOF); ok {
				break
			}
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
func (p *parser) consumeRules(s Scanner, toplevel bool) ast.Rules {
	var a ast.Rules
	for {
		tok := s.Scan()
		switch tok.(type) {
		case *token.Whitespace:
			// nop
		case *token.EOF:
			return a
		case *token.CDO, *token.CDC:
			if !toplevel {
				s.Unscan()
				if r := p.consumeQualifiedRule(s); r != nil {
					a = append(a, r)
				}
			}
		case *token.AtKeyword:
			s.Unscan()
			if r := p.consumeAtRule(s); r != nil {
				a = append(a, r)
			}
		default:
			if r := p.consumeQualifiedRule(s); r != nil {
				a = append(a, r)
			}
		}
	}
}

// consumeAtRule consumes a single at-rule. (§5.4.2)
func (p *parser) consumeAtRule(s Scanner) *ast.AtRule {
	r := &ast.AtRule{}

	// Set the name to the value of the current token.
	atkeyword := s.Current().(*token.AtKeyword)
	r.Name = atkeyword.Value

	// Repeatedly consume the next token.
	for {
		tok := s.Scan()
		switch tok.(type) {
		case *token.Semicolon, *token.EOF:
			return r
		case *token.LBrace:
			r.Block = p.consumeSimpleBlock(s)
			return r
		// TODO: simple block with an associated token of <{-token> ???
		default:
			s.Unscan()
			v := p.consumeComponentValue(s)
			r.Prelude = append(r.Prelude, v)
		}
	}
}

// consumeAtRule consumes a single qualified rule. (§5.4.3)
func (p *parser) consumeQualifiedRule(s Scanner) *ast.QualifiedRule {
	r := &ast.QualifiedRule{}

	// Repeatedly consume the next token.
	for {
		tok := s.Scan()
		switch tok := tok.(type) {
		case *token.EOF:
			p.errors = append(p.errors, &Error{Message: "unexpected EOF", Pos: tok.Pos})
			return nil
		case *token.LBrace:
			r.Block = p.consumeSimpleBlock(s)
			return r
		// TODO: simple block with an associated token of <{-token> ???
		default:
			s.Unscan()
			r.Prelude = append(r.Prelude, p.consumeComponentValue(s))
		}
	}
}

// consumeDeclarations consumes a list of declarations. (§5.4.4)
func (p *parser) consumeDeclarations(s Scanner) ast.Declarations {
	var a ast.Declarations

	// Repeatedly consume the next token.
	for {
		tok := s.Scan()
		switch tok := tok.(type) {
		case *token.Whitespace, *token.Semicolon:
			// nop
		case *token.EOF:
			return a
		case *token.AtKeyword:
			a = append(a, p.consumeAtRule(s))
		case *token.Ident:
			// Generate a list of tokens up to the next semicolon or EOF.
			s.Unscan()
			tokens := p.consumeDeclarationTokens(s)

			// Consume declaration using temporary list of tokens.
			if d := p.consumeDeclaration(NewTokenScanner(tokens)); d != nil {
				a = append(a, d)
			}

		default:
			// Any other token is a syntax error.
			p.errors = append(p.errors, &Error{Message: fmt.Sprintf("unexpected %s", tok.String()), Pos: tok.Position()})

			// Repeatedly consume a component values until semicolon or EOF.
			p.skipComponentValues(s)
		}
	}
}

// consumeDeclaration consumes a single declaration. (§5.4.5)
func (p *parser) consumeDeclaration(s Scanner) *ast.Declaration {
	d := &ast.Declaration{}

	// The first token must be an ident.
	ident := s.Scan().(*token.Ident)
	d.Name = ident.Value

	// Skip over whitespace.
	p.skipWhitespace(s)

	// The next token must be a colon.
	if _, ok := s.Scan().(*token.Colon); !ok {
		p.errors = append(p.errors, &Error{Message: fmt.Sprintf("expected colon, got %s", s.Current().String()), Pos: s.Current().Position()})
		return nil
	}

	// Consume the declaration value until EOF.
	for {
		tok := s.Scan()
		if _, ok := tok.(*token.EOF); ok {
			break
		}
		d.Values = append(d.Values, &ast.Token{tok})
	}

	// Check last two non-whitespace tokens for "!important".
	d.Values, d.Important = cleanImportantFlag(d.Values)

	return d
}

// Checks if the last two non-whitespace tokens are a case-insensitive "!important".
// If so, it removes them and returns the "important" flag set to true.
func cleanImportantFlag(values ast.ComponentValues) (ast.ComponentValues, bool) {
	return values, false // TODO(benbjohnson)
}

// consumeComponentValue consumes a single component value. (§5.4.6)
func (p *parser) consumeComponentValue(s Scanner) ast.ComponentValue {
	tok := s.Scan()
	switch tok.(type) {
	case *token.LBrace, *token.LBrack, *token.LParen:
		return p.consumeSimpleBlock(s)
	case *token.Function:
		return p.consumeFunction(s)
	default:
		return &ast.Token{tok}
	}
}

// consumeSimpleBlock consumes a simple block. (§5.4.7)
func (p *parser) consumeSimpleBlock(s Scanner) *ast.SimpleBlock {
	b := &ast.SimpleBlock{}

	// Set the block's associated token to the current token.
	b.Token = s.Current()

	for {
		tok := s.Scan()

		// If this token is EOF or the mirror of the starting token then return.
		switch tok.(type) {
		case *token.EOF:
			return b
		case *token.RBrack:
			if _, ok := b.Token.(*token.LBrack); ok {
				return b
			}
		case *token.RBrace:
			if _, ok := b.Token.(*token.LBrace); ok {
				return b
			}
		case *token.RParen:
			if _, ok := b.Token.(*token.LParen); ok {
				return b
			}
		}

		// Otherwise consume a component value.
		s.Unscan()
		b.Values = append(b.Values, p.consumeComponentValue(s))
	}
}

// consumeFunction consumes a function. (§5.4.8)
func (p *parser) consumeFunction(s Scanner) *ast.Function {
	f := &ast.Function{}

	// Set the name to the first token.
	f.Name = s.Current().(*token.Function).Value

	for {
		tok := s.Scan()

		// If this token is EOF or the mirror of the starting token then return.
		switch tok.(type) {
		case *token.EOF, *token.RParen:
			return f
		}

		// Otherwise consume a component value.
		s.Unscan()
		f.Values = append(f.Values, p.consumeComponentValue(s))
	}
}

// consumeDeclarationTokens collects contiguous non-semicolon and non-EOF tokens.
func (p *parser) consumeDeclarationTokens(s Scanner) []token.Token {
	var a []token.Token
	for {
		tok := s.Scan()
		switch tok.(type) {
		case *token.Semicolon, *token.EOF:
			s.Unscan()
			return a
		}
		a = append(a, tok)
	}
}

// skipComponentValues consumes all component values until a semicolon or EOF.
func (p *parser) skipComponentValues(s Scanner) {
	for {
		v := p.consumeComponentValue(s)
		if tok, ok := v.(*ast.Token); ok {
			switch tok.Token.(type) {
			case *token.Semicolon, *token.EOF:
				return
			}
		}
	}
}

// skipWhitespace skips over all contiguous whitespace tokes.
func (p *parser) skipWhitespace(s Scanner) {
	for {
		if _, ok := s.Scan().(*token.Whitespace); !ok {
			s.Unscan()
			return
		}
	}
}

// Scanner represents a type that can retrieve the next token.
type Scanner interface {
	Current() token.Token
	Scan() token.Token
	Unscan()
}

// TokenScanner represents a scanner for a fixed list of tokens.
type TokenScanner struct {
	i      int
	tokens []token.Token
}

// NewTokenScanner returns a new instance of TokenScanner.
func NewTokenScanner(tokens []token.Token) *TokenScanner {
	return &TokenScanner{tokens: tokens}
}

// Current returns the current token.
func (s *TokenScanner) Current() token.Token {
	if s.i > len(s.tokens) {
		return &token.EOF{}
	}
	return s.tokens[s.i]
}

// Scan returns the next token.
func (s *TokenScanner) Scan() token.Token {
	if s.i <= len(s.tokens) {
		s.i++
	}
	return s.Current()
}

// Unscan moves back one token.
func (s *TokenScanner) Unscan() {
	if s.i > -1 {
		s.i--
	}
}

// Error represents a syntax error.
type Error struct {
	Message string
	Pos     token.Pos
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
