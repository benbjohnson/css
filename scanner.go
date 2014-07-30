package css

import (
	"bufio"
	"bytes"
	"io"
)

// Scanner implements a CSS3 standard compliant scanner.
type Scanner struct {
	// Type is set after parsing an ident-token, function-token,
	// at-keyword-token, hash-token, string-token, and url-token.
	// It is set to either "id" or "unrestricted".
	Type string

	// Value is the literal representation of the last read token.
	Value string

	// These numeric values are set after scanning a number-token,
	// a percentage-token, or a dimension-token.
	IntValue    int
	NumberValue int

	// Start and End are set after each unicode-range token.
	Start int
	End   int

	rd  io.RuneReader
	pos Pos

	ch  rune    // current code point
	idx int     // bufferred input index
	buf [3]rune // buffered input
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	// TODO(benbjohnson): Determine fallback encoding (§3.2).
	return &Scanner{
		rd:  bufio.NewReader(r),
		idx: -1,
	}
}

func (s *Scanner) Scan() (pos Pos, tok Token) {
	// Mark the start position for this scan.
	pos = s.pos

	// Initialize fields.
	s.Type = "unrestricted"
	s.Value = ""
	s.IntValue = 0
	s.NumberValue = 0
	s.Start = 0
	s.End = 0

	// Read next code point.
	ch, err := s.read()
	if err == io.EOF {
		tok = EOF
		return
	}

	// Scan all contiguous whitespace.
	if ch == ' ' || ch == '\t' || ch == '\n' {
		tok = WHITESPACE
		s.Value = s.scanWhitespace()
		return
	}

	// Scan a string if it starts with a quote.
	if ch == '"' || ch == '\'' {
		tok, s.Value = s.scanString()
		return
	}

	// Scan a hash token.
	if ch == '#' {
		// If there is a name following the hash then we have a hash token.
		if s.peekName() || s.peekEscape() {
			// If the name is an identifier then change the type.
			if s.peekIdent() {
				s.Type = "id"
			}
			tok, s.Value = HASH, s.scanName()
			return
		}

		// If there is no name following the hash symbol then return delim-token.
		tok, s.Value = DELIM, string(ch)
		return
	}

	// Scan a suffix-match token.
	if ch == '$' {
		if next, err := s.read(); err != nil {
			tok = EOF
		} else if next == '=' {
			tok, s.Value = SUFFIXMATCH, "$="
		} else {
			s.unread(next)
			tok, s.Value = DELIM, string(ch)
		}
		return
	}

	return
}

// scanWhitespace consumes the current code point and all subsequent whitespace.
func (s *Scanner) scanWhitespace() string {
	var buf bytes.Buffer
	_, _ = buf.WriteRune(s.ch)
	for {
		ch, err := s.read()
		if err == io.EOF {
			break
		} else if ch != ' ' && ch != '\t' && ch != '\n' {
			s.unread(ch)
			break
		}
		_, _ = buf.WriteRune(ch)
	}
	return buf.String()
}

// scanString consumes a quoted string.
func (s *Scanner) scanString() (Token, string) {
	var buf bytes.Buffer
	ending := s.ch
	for {
		ch, err := s.read()
		if err == io.EOF || ch == ending {
			return STRING, buf.String()
		} else if ch == '\n' {
			s.unread(ch)
			return BADSTRING, buf.String()
		} else if ch == '\\' {
			if next, err := s.read(); err == io.EOF {
				continue
			} else if next == '\n' {
				_, _ = buf.WriteRune(next)
			} else if s.peekEscape() {
				_, _ = buf.WriteRune(s.scanEscape())
			} else {
				s.unread(next)
				_, _ = buf.WriteRune(ch)
			}
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}
}

// scanName consumes a name.
func (s *Scanner) scanName() string {
	return "" // TODO(benbjohnson)
}

// scanEscape consumes an escaped code point.
func (s *Scanner) scanEscape() rune {
	return '☃' // TODO(benbjohnson)
}

// peekName checks if the next code point is a name code point.
func (s *Scanner) peekName() bool {
	return false // TODO(benbjohnson)
}

// peekEscape checks if the next code points are a valid escape.
func (s *Scanner) peekEscape() bool {
	return false // TODO(benbjohnson)
}

// peekIdent checks if the next code points are a valid identifier.
func (s *Scanner) peekIdent() bool {
	return false // TODO(benbjohnson)
}

// read reads the next rune from the reader.
func (s *Scanner) read() (rune, error) {
	// If we have runes on our internal lookahead buffer then return those.
	if s.idx > -1 {
		s.ch = s.buf[s.idx]
		s.idx--
		return s.ch, nil
	}

	// Otherwise read from the reader.
	ch, _, err := s.rd.ReadRune()
	if err != nil {
		return 0, err
	}

	// Preprocess the input stream by replacing FF with LF. (§3.3)
	if ch == '\f' {
		ch = '\n'
	}

	// Preprocess the input stream by replacing CR and CRLF with LF. (§3.3)
	if ch == '\r' {
		if ch, _, err := s.rd.ReadRune(); err == io.EOF {
			// nop
		} else if err != nil {
			return 0, err
		} else if ch != '\n' {
			s.unread(ch)
		}
		ch = '\n'
	}

	// Replace NULL with Unicode replacement character. (§3.3)
	if ch == '\000' {
		ch = '\uFFFD'
	}

	// Track scanner position.
	if ch == '\n' {
		s.pos.Line++
		s.pos.Char = 0
	} else {
		s.pos.Char++
	}

	s.ch = ch
	return ch, nil
}

// unread puts a run on the internal buffer.
func (s *Scanner) unread(ch rune) {
	s.idx++
	s.buf[s.idx] = ch
}
