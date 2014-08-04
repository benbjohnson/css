package css

import (
	"io"
)

// parser represents a CSS3 parser.
type parser struct {
	scanner *Scanner

	buf  [2]Token // circular buffer
	bufi int      // circular buffer index
	bufn int      // number of buffered tokens
}

// ParseStylesheet parses an input stream into a stylesheet.
func (p *parser) ParseStylesheet(r io.Reader) *Stylesheet {
	return nil // TODO(benbjohnson)
}

// ParseRule parses a list of rules.
func (p *parser) ParseRules(r io.Reader) Rules {
	return nil // TODO(benbjohnson)
}

// ParseRule parses a qualified rule or at-rule.
func (p *parser) ParseRule(r io.Reader) Rule {
	return nil // TODO(benbjohnson)
}

// ParseDeclaration parses a name/value declaration.
func (p *parser) ParseDeclaration(r io.Reader) *Declaration {
	return nil // TODO(benbjohnson)
}

// ParseComponentValue parses a component value.
func (p *parser) ParseComponentValue(r io.Reader) *ComponentValue {
	return nil // TODO(benbjohnson)
}

// ParseComponentValues parses a list of component values.
func (p *parser) ParseComponentValues(r io.Reader) ComponentValues {
	return nil // TODO(benbjohnson)
}

// scan reads the next token from the scanner.
func (p *parser) scan() Token {
	// If we have tokens on our internal lookahead buffer then return those.
	if p.bufn > 0 {
		p.bufi = ((p.bufi + 1) % len(p.buf))
		p.bufn--
		return p.buf[p.bufi]
	}

	// Otherwise read from the scanner.
	_, tok := p.scanner.Scan()

	// Add to circular buffer.
	p.bufi = ((p.bufi + 1) % len(p.buf))
	p.buf[p.bufi] = tok
	return tok
}

// unscan adds the previous token back onto the buffer.
func (p *Scanner) unscan() {
	p.bufi = ((p.bufi + len(p.buf) - 1) % len(p.buf))
	p.bufn++
}

// curr reads the current token.
func (p *parser) curr() Token {
	return p.buf[p.bufi]
}
