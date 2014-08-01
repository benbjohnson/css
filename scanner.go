package css

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
)

// eof represents an EOF file byte.
var eof rune = 0

// Scanner implements a CSS3 standard compliant scanner.
type Scanner struct {
	// Type is set after parsing an ident-token, function-token,
	// at-keyword-token, hash-token, string-token, and url-token.
	// It is set to either "id" or "unrestricted".
	//
	// This is also set after scanning a numeric token and can
	// be set to "integer" or "number".
	Type string

	// Value is the literal representation of the last read token.
	Value string

	// This numeric value is set after scanning a number-token,
	// a percentage-token, or a dimension-token. The unit is
	// returned for dimension tokens.
	Number float64
	Unit   string

	// Start and End are set after each unicode-range token.
	Start int
	End   int

	// Ending represents the ending code point of a string.
	Ending rune

	// Errors contains a list of all errors that occur during scanning.
	Errors []*Error

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
	s.Number = 0
	s.Start = 0
	s.End = 0
	s.Ending = '\000'

	for {
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
		} else if ch == '*' {
			tok, s.Value = s.scanSubstringMatch()
		} else if ch == '^' {
			tok, s.Value = s.scanPrefixMatch()
		} else if ch == '~' {
			tok, s.Value = s.scanIncludeMatch()
		} else if ch == ',' {
			tok = COMMA
		} else if ch == '-' {
			// Scan then next two tokens and unread back to the hyphen.
			ch1, ch2 := s.read(), s.read()
			s.unread(3)

			// If we have a digit next, it's a numeric token. If it's an identifier
			// then scan an identifier, and if it's a "->" then it's a CDC.
			if isDigit(ch1) || ch1 == '.' {
				tok, s.Number, s.Value, s.Type, s.Unit = s.scanNumeric()
			} else if s.peekIdent() {
				tok, s.Value = s.scanIdent()
			} else if ch1 == '-' && ch2 == '>' {
				tok = CDC
			} else {
				tok, s.Value = DELIM, "-"
			}
		} else if ch == '/' {
			// Comments are ignored by the scanner so this will leave the tok
			// set to ILLEGAL and the outer for loop will iterate again.
			if ch1 := s.read(); ch1 == '*' {
				s.scanComment()
				tok = ILLEGAL
			} else {
				s.unread(1)
				tok, s.Value = DELIM, "/"
			}
		} else if ch == ':' {
			tok = COLON
		} else if ch == ';' {
			tok = SEMICOLON
		} else if ch == '<' {
			// Attempt to read a comment open ("<!--").
			// If it's not possible then then rollback and return DELIM.
			if ch0 := s.read(); ch0 == '!' {
				if ch1 := s.read(); ch1 == '-' {
					if ch2 := s.read(); ch2 == '-' {
						tok = CDO
						break
					}
					s.unread(1)
				}
				s.unread(1)
			}
			s.unread(1)
			tok, s.Value = DELIM, "<"
		} else if ch == '@' {
			// This is an at-keyword token if an identifier follows.
			// Otherwise it's just a DELIM.
			if s.read(); s.peekIdent() {
				tok, s.Value = ATKEYWORD, s.scanName()
			} else {
				tok, s.Value = DELIM, "@"
			}
		} else if ch == '(' {
			tok = LPAREN
		} else if ch == ')' {
			tok = RPAREN
		} else if ch == '[' {
			tok = LBRACK
		} else if ch == ']' {
			tok = RBRACK
		} else if ch == '{' {
			tok = LBRACE
		} else if ch == '}' {
			tok = RBRACE
		} else if ch == '\\' {
			// Return a valid escape, if possible.
			// Otherwise this is a parse error but continue on as a DELIM.
			if s.peekEscape() {
				tok, s.Value = s.scanIdent()
			} else {
				s.Errors = append(s.Errors, &Error{Message: "unescaped \\", Pos: pos})
				tok, s.Value = DELIM, "\\"
			}
		} else if ch == '+' || ch == '.' || isDigit(ch) {
			s.unread(1)
			tok, s.Number, s.Value, s.Type, s.Unit = s.scanNumeric()
		} else if ch == 'u' || ch == 'U' {
			// TODO: Peek "+hex" or "+?", consume next code point, consume unicode-range.
			// TODO: Otherwise reconsume as ident.
		} else if isNameStart(ch) {
			// TODO: Reconsume as ident.
		} else if ch == '|' {
			// TODO: Peek "=" then dash-match token.
			// TODO: Peek "|" then column token.
			// TODO: Otherwise DELIM.
		} else {
			tok, s.Value = DELIM, string(ch)
		}

		// C-style comments are ignored so they return an ILLEGAL token.
		// If this occurs then just try again until we hit a real token or EOF.
		if tok != ILLEGAL {
			break
		}
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
			s.unread(1)
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
			s.unread(1)
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

// scanNumeric consumes a numeric token.
//
// This assumes that the current token is a +, -, . or digit.
func (s *Scanner) scanNumeric() (tok Token, num float64, repr, typ, unit string) {
	num, typ, repr = s.scanNumber()

	// If the number is immediately followed by an identifier then scan dimension.
	if s.read(); s.peekIdent() {
		tok = DIMENSION
		unit = s.scanName()
		repr += unit
		return
	} else {
		s.unread(1)
	}

	// If the number is followed by a percent sign then return a percentage.
	if ch := s.read(); ch == '%' {
		tok = PERCENTAGE
		repr += "%"
		return
	} else {
		s.unread(1)
	}

	// Otherwise return a number token.
	tok = NUMBER
	return
}

// scanNumber consumes a number.
func (s *Scanner) scanNumber() (num float64, typ, repr string) {
	var buf bytes.Buffer
	typ = "integer"

	// If initial code point is + or - then store it.
	if ch := s.read(); ch == '+' || ch == '-' {
		_, _ = buf.WriteRune(ch)
	} else {
		s.unread(1)
	}

	// Read as many digits as possible.
	_, _ = buf.WriteString(s.scanDigits())

	// If next code points are a full stop and digit then consume them.
	if ch0 := s.read(); ch0 == '.' {
		if ch1 := s.read(); isDigit(ch1) {
			typ = "number"
			_, _ = buf.WriteRune(ch0)
			_, _ = buf.WriteRune(ch1)
			_, _ = buf.WriteString(s.scanDigits())
		} else {
			s.unread(2)
		}
	} else {
		s.unread(1)
	}

	// Consume scientific notation (e0, e+0, e-0, E0, E+0, E-0).
	if ch0 := s.read(); ch0 == 'e' || ch0 == 'E' {
		if ch1 := s.read(); ch1 == '+' || ch1 == '-' {
			if ch2 := s.read(); isDigit(ch2) {
				typ = "number"
				_, _ = buf.WriteRune(ch0)
				_, _ = buf.WriteRune(ch1)
				_, _ = buf.WriteRune(ch2)
			} else {
				s.unread(3)
			}
		} else if isDigit(ch1) {
			typ = "number"
			_, _ = buf.WriteRune(ch0)
			_, _ = buf.WriteRune(ch1)
		} else {
			s.unread(2)
		}
	} else {
		s.unread(1)
	}

	// Parse number.
	num, _ = strconv.ParseFloat(buf.String(), 64)
	repr = buf.String()
	return
}

// scanDigits consume a contiguous series of digits.
func (s *Scanner) scanDigits() string {
	var buf bytes.Buffer
	for {
		if ch := s.read(); isDigit(ch) {
			_, _ = buf.WriteRune(ch)
		} else {
			s.unread(1)
			break
		}
	}
	return buf.String()
}

// scanComment consumes all characters up to "*/", inclusive.
// This function assumes that the initial "/*" have just been consumed.
func (s *Scanner) scanComment() {
	for {
		ch0 := s.read()
		if ch0 == eof {
			break
		} else if ch0 == '*' {
			if ch1 := s.read(); ch1 == '/' {
				break
			} else {
				s.unread(1)
			}
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
	s.unread(1)
	return DELIM, "#"
}

// scanSuffixMatch consumes a suffix-match token.
func (s *Scanner) scanSuffixMatch() (Token, string) {
	if next := s.read(); next == '=' {
		return SUFFIXMATCH, ""
	}
	s.unread(1)
	return DELIM, "$"
}

// scanSubstringMatch consumes a substring-match token.
func (s *Scanner) scanSubstringMatch() (Token, string) {
	if next := s.read(); next == '=' {
		return SUBSTRINGMATCH, ""
	}
	s.unread(1)
	return DELIM, "*"
}

// scanPrefixMatch consumes a prefix-match token.
func (s *Scanner) scanPrefixMatch() (Token, string) {
	if next := s.read(); next == '=' {
		return PREFIXMATCH, ""
	}
	s.unread(1)
	return DELIM, "^"
}

// scanIncludeMatch consumes an include-match token.
func (s *Scanner) scanIncludeMatch() (Token, string) {
	if next := s.read(); next == '=' {
		return INCLUDEMATCH, ""
	}
	s.unread(1)
	return DELIM, "~"
}

// scanName consumes a name.
// Consumes contiguous name code points and escaped code points.
func (s *Scanner) scanName() string {
	var buf bytes.Buffer
	s.unread(1)
	for {
		if ch := s.read(); isName(ch) {
			_, _ = buf.WriteRune(ch)
		} else if s.peekEscape() {
			_, _ = buf.WriteRune(s.scanEscape())
		} else {
			s.unread(1)
			return buf.String()
		}
	}
}

// scanIdent consumes a ident-like token.
// This function can return an ident, function, url, or bad-url.
func (s *Scanner) scanIdent() (Token, string) {
	v := s.scanName()

	// Check if this is the start of a url token.
	if strings.ToLower(v) == "url" {
		if ch := s.read(); ch == '(' {
			return s.scanURL()
		}
		s.unread(1)
	} else if ch := s.read(); ch == '(' {
		return FUNCTION, v
	}
	s.unread(1)

	return IDENT, v
}

// scanURL consumes the contents of a URL function.
// This function assumes that the "url(" has just been consumed.
// This function can return a url or bad-url token.
func (s *Scanner) scanURL() (Token, string) {
	return 0, "" // TODO(benbjohnson)
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
				s.unread(1)
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
	s.unread(1)
	return next != '\n'
}

// peekIdent checks if the next code points are a valid identifier.
func (s *Scanner) peekIdent() bool {
	if s.peek() == '-' {
		ch := s.read()
		s.unread(1)
		return isNameStart(ch) || s.peekEscape()
	} else if isNameStart(s.peek()) {
		return true
	} else if s.peek() == '\\' && s.peekEscape() {
		return true
	}
	return false
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
			s.unread(1)
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

// unread adds the previous n code points back onto the buffer.
func (s *Scanner) unread(n int) {
	for i := 0; i < n; i++ {
		s.bufi = ((s.bufi + len(s.buf) - 1) % len(s.buf))
		s.bufn++
	}
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
