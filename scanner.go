package css

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
)

// eof represents an EOF file byte.
var eof rune = 0

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

	// Ending represents the ending code point of a string.
	Ending rune

	rd  io.RuneReader
	pos Pos

	buf  [4]rune // circular buffer
	bufi int     // circular buffer index
	bufn int     // number of buffered characters
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	// TODO(benbjohnson): Determine fallback encoding (§3.2).
	return &Scanner{
		rd: bufio.NewReader(r),
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
	s.Ending = '\000'

	// Read next code point.
	ch := s.read()
	if ch == eof {
		tok = EOF
	} else if isWhitespace(ch) {
		tok = WHITESPACE
		s.Value = s.scanWhitespace()
	} else if ch == '"' || ch == '\'' {
		tok, s.Value = s.scanString()
	} else if ch == '#' {
		tok, s.Value = s.scanHash()
	} else if ch == '$' {
		tok, s.Value = s.scanSuffixMatch()
	} else if ch == '(' {
		tok = LPAREN
	} else if ch == ')' {
		tok = RPAREN
	} else if ch == '*' {
		tok, s.Value = s.scanSubstringMatch()
	}

	return
}

// scanWhitespace consumes the current code point and all subsequent whitespace.
func (s *Scanner) scanWhitespace() string {
	var buf bytes.Buffer
	_, _ = buf.WriteRune(s.peek())
	for {
		ch := s.read()
		if ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		}
		_, _ = buf.WriteRune(ch)
	}
	return buf.String()
}

// scanString consumes a quoted string. (§4.3.4)
//
// This assumes that the current token is a single or double quote.
// This function consumes all code points and escaped code points up until
// a matching, unescaped ending quote.
// An EOF closes out a string but does not return an error.
// A newline will close a string and returns a bad-string token.
func (s *Scanner) scanString() (Token, string) {
	var buf bytes.Buffer
	s.Ending = s.peek()
	for {
		ch := s.read()
		if ch == eof || ch == s.Ending {
			return STRING, buf.String()
		} else if ch == '\n' {
			s.unread()
			return BADSTRING, buf.String()
		} else if ch == '\\' {
			if s.peekEscape() {
				_, _ = buf.WriteRune(s.scanEscape())
				continue
			}
			if next := s.read(); next == eof {
				continue
			} else if next == '\n' {
				_, _ = buf.WriteRune(next)
			}
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}
}

// scanHash consumes a hash token.
//
// This assumes the current token is a '#' code point.
// It will return a hash token if the next code points are a name or valid escape.
// It will return a delim token otherwise.
// Hash tokens' type flag is set to "id" if its value is an identifier.
func (s *Scanner) scanHash() (Token, string) {
	// If there is a name following the hash then we have a hash token.
	if ch := s.read(); isName(ch) || s.peekEscape() {
		// If the name is an identifier then change the type.
		if s.peekIdent() {
			s.Type = "id"
		}
		return HASH, s.scanName()
	}

	// If there is no name following the hash symbol then return delim-token.
	s.unread()
	return DELIM, "#"
}

// scanSuffixMatch consumes a suffix-match token.
func (s *Scanner) scanSuffixMatch() (Token, string) {
	if next := s.read(); next == '=' {
		return SUFFIXMATCH, ""
	}
	s.unread()
	return DELIM, "$"
}

// scanSubstringMatch consumes a string-match token.
func (s *Scanner) scanSubstringMatch() (Token, string) {
	if next := s.read(); next == '=' {
		return SUBSTRINGMATCH, ""
	}
	s.unread()
	return DELIM, "*"
}

// scanName consumes a name.
// Consumes contiguous name code points and escaped code points.
func (s *Scanner) scanName() string {
	var buf bytes.Buffer
	_, _ = buf.WriteRune(s.peek())
	for {
		if ch := s.read(); isName(ch) {
			_, _ = buf.WriteRune(ch)
		} else if s.peekEscape() {
			_, _ = buf.WriteRune(s.scanEscape())
		} else {
			s.unread()
			return buf.String()
		}
	}
}

// scanEscape consumes an escaped code point.
func (s *Scanner) scanEscape() rune {
	var buf bytes.Buffer
	ch := s.read()
	if isHexDigit(ch) {
		_, _ = buf.WriteRune(ch)
		for i := 0; i < 5; i++ {
			if next := s.read(); next == eof || isWhitespace(next) {
				break
			} else if !isHexDigit(next) {
				s.unread()
				break
			} else {
				_, _ = buf.WriteRune(next)
			}
		}
		v, _ := strconv.ParseInt(buf.String(), 16, 0)
		return rune(v)
	} else if ch == eof {
		return '\uFFFD'
	} else {
		return ch
	}
}

// peekEscape checks if the next code points are a valid escape.
func (s *Scanner) peekEscape() bool {
	// If the current code point is not a backslash then this is not an escape.
	if s.peek() != '\\' {
		return false
	}

	// If the next code point is a newline then this is not an escape.
	next := s.read()
	if next != eof {
		s.unread()
	}
	return next != '\n'
}

// peekIdent checks if the next code points are a valid identifier.
func (s *Scanner) peekIdent() bool {
	return false // TODO(benbjohnson)
}

// read reads the next rune from the reader.
// This function will initially check for any characters that have been pushed
// back onto the lookahead buffer and return those. Otherwise it will read from
// the reader and do preprocessing to convert newline characters and NULL.
// EOF is converted to a zero rune (\000) and returned.
func (s *Scanner) read() rune {
	// If we have runes on our internal lookahead buffer then return those.
	if s.bufn > 0 {
		s.bufi = ((s.bufi + 1) % len(s.buf))
		s.bufn--
		return s.buf[s.bufi]
	}

	// Otherwise read from the reader.
	ch, _, err := s.rd.ReadRune()
	if err != nil {
		return eof
	}

	// Preprocess the input stream by replacing FF with LF. (§3.3)
	if ch == '\f' {
		ch = '\n'
	}

	// Preprocess the input stream by replacing CR and CRLF with LF. (§3.3)
	if ch == '\r' {
		if ch, _, err := s.rd.ReadRune(); err != nil {
			// nop
		} else if ch != '\n' {
			s.unread()
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

	// Add to circular buffer.
	s.bufi = ((s.bufi + 1) % len(s.buf))
	s.buf[s.bufi] = ch
	return ch
}

// unread puts a run on the internal buffer.
func (s *Scanner) unread() {
	s.bufi = ((s.bufi + len(s.buf) - 1) % len(s.buf))
	s.bufn++
}

// peek reads the current code point.
func (s *Scanner) peek() rune {
	return s.buf[s.bufi]
}

// isWhitespace returns true if the rune is a space, tab, or newline.
func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

// isLetter returns true if the rune is a letter.
func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// isDigit returns true if the rune is a digit.
func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

// isHexDigit returns true if the rune is a hex digit.
func isHexDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

// isNonASCII returns true if the rune is greater than U+0080.
func isNonASCII(ch rune) bool {
	return ch >= '\u0080'
}

// isNameStart returns true if the rune can start a name.
func isNameStart(ch rune) bool {
	return isLetter(ch) || isNonASCII(ch) || ch == '_'
}

// isName returns true if the character is a name code point.
func isName(ch rune) bool {
	return isNameStart(ch) || isDigit(ch) || ch == '-'
}
