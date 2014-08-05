package css

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// eof represents an EOF file byte.
var eof rune = -1

// Scanner implements a CSS3 standard compliant scanner.
//
// This implementation only allows UTF-8 encoding.
// @charset directives will be ignored.
type Scanner struct {
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
	return &Scanner{
		rd: bufio.NewReader(r),
	}
}

func (s *Scanner) Scan() *Token {
	for {
		tok := &Token{Pos: s.pos}

		// Read next code point.
		ch := s.read()
		if ch == eof {
			tok.Type = EOF
		} else if isWhitespace(ch) {
			tok.Type, tok.Value = WHITESPACE, s.scanWhitespace()
		} else if ch == '"' || ch == '\'' {
			tok.Type, tok.Value, tok.Ending = s.scanString()
		} else if ch == '#' {
			tok.Type, tok.Value, tok.Flag = s.scanHash()
		} else if ch == '$' {
			tok.Type, tok.Value = s.scanSuffixMatch()
		} else if ch == '*' {
			tok.Type, tok.Value = s.scanSubstringMatch()
		} else if ch == '^' {
			tok.Type, tok.Value = s.scanPrefixMatch()
		} else if ch == '~' {
			tok.Type, tok.Value = s.scanIncludeMatch()
		} else if ch == ',' {
			tok.Type = COMMA
		} else if ch == '-' {
			// Scan then next two tokens and unread back to the hyphen.
			ch1, ch2 := s.read(), s.read()
			s.unread(3)

			// If we have a digit next, it's a numeric token. If it's an identifier
			// then scan an identifier, and if it's a "->" then it's a CDC.
			if isDigit(ch1) || ch1 == '.' {
				tok.Type, tok.Number, tok.Value, tok.Flag, tok.Unit = s.scanNumeric()
			} else if s.peekIdent() {
				tok.Type, tok.Value = s.scanIdent()
			} else if ch1 == '-' && ch2 == '>' {
				tok.Type = CDC
			} else {
				tok.Type, tok.Value = DELIM, "-"
			}
		} else if ch == '/' {
			// Comments are ignored by the scanner so this will leave the tok
			// set to ILLEGAL and the outer for loop will iterate again.
			if ch1 := s.read(); ch1 == '*' {
				s.scanComment()
				tok.Type = ILLEGAL
			} else {
				s.unread(1)
				tok.Type, tok.Value = DELIM, "/"
			}
		} else if ch == ':' {
			tok.Type = COLON
		} else if ch == ';' {
			tok.Type = SEMICOLON
		} else if ch == '<' {
			// Attempt to read a comment open ("<!--").
			// If it's not possible then then rollback and return DELIM.
			if ch0 := s.read(); ch0 == '!' {
				if ch1 := s.read(); ch1 == '-' {
					if ch2 := s.read(); ch2 == '-' {
						tok.Type = CDO
						return tok
					}
					s.unread(1)
				}
				s.unread(1)
			}
			s.unread(1)
			tok.Type, tok.Value = DELIM, "<"
		} else if ch == '@' {
			// This is an at-keyword token if an identifier follows.
			// Otherwise it's just a DELIM.
			if s.read(); s.peekIdent() {
				tok.Type, tok.Value = ATKEYWORD, s.scanName()
			} else {
				tok.Type, tok.Value = DELIM, "@"
			}
		} else if ch == '(' {
			tok.Type = LPAREN
		} else if ch == ')' {
			tok.Type = RPAREN
		} else if ch == '[' {
			tok.Type = LBRACK
		} else if ch == ']' {
			tok.Type = RBRACK
		} else if ch == '{' {
			tok.Type = LBRACE
		} else if ch == '}' {
			tok.Type = RBRACE
		} else if ch == '\\' {
			// Return a valid escape, if possible.
			// Otherwise this is a parse error but continue on as a DELIM.
			if s.peekEscape() {
				tok.Type, tok.Value = s.scanIdent()
			} else {
				s.Errors = append(s.Errors, &Error{Message: "unescaped \\", Pos: s.pos})
				tok.Type, tok.Value = DELIM, "\\"
			}
		} else if ch == '+' || ch == '.' || isDigit(ch) {
			s.unread(1)
			tok.Type, tok.Number, tok.Value, tok.Flag, tok.Unit = s.scanNumeric()
		} else if ch == 'u' || ch == 'U' {
			// Peek "+[0-9a-f]" or "+?", consume next code point, consume unicode-range.
			ch1, ch2 := s.read(), s.read()
			if ch1 == '+' && (isHexDigit(ch2) || ch2 == '?') {
				s.unread(1)
				tok.Type = UNICODERANGE
				tok.Start, tok.End = s.scanUnicodeRange()
			} else {
				// Otherwise reconsume as ident.
				s.unread(2)
				tok.Type, tok.Value = s.scanIdent()
			}
		} else if isNameStart(ch) {
			tok.Type, tok.Value = s.scanIdent()
		} else if ch == '|' {
			// If the next token is an equals sign, it's a dash token.
			// If the next token is a pipe, it's a column token.
			// Otherwise, just treat this pipe as a delim token.
			if ch1 := s.read(); ch1 == '=' {
				tok.Type, tok.Value = DASHMATCH, ""
			} else if ch1 == '|' {
				tok.Type, tok.Value = COLUMN, ""
			} else {
				s.unread(1)
				tok.Type, tok.Value = DELIM, string(ch)
			}
		} else {
			tok.Type, tok.Value = DELIM, string(ch)
		}

		// C-style comments are ignored so they return an ILLEGAL token.
		// If this occurs then just try again until we hit a real token or EOF.
		if tok.Type != ILLEGAL {
			return tok
		}
	}
}

// scanWhitespace consumes the current code point and all subsequent whitespace.
func (s *Scanner) scanWhitespace() string {
	var buf bytes.Buffer
	_, _ = buf.WriteRune(s.curr())
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
func (s *Scanner) scanString() (typ TokenType, value string, ending rune) {
	var buf bytes.Buffer
	ending = s.curr()
	for {
		ch := s.read()
		if ch == eof || ch == ending {
			typ, value = STRING, buf.String()
			return
		} else if ch == '\n' {
			s.unread(1)
			typ, value = BADSTRING, buf.String()
			return
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
func (s *Scanner) scanNumeric() (typ TokenType, num float64, repr, flag, unit string) {
	num, flag, repr = s.scanNumber()

	// If the number is immediately followed by an identifier then scan dimension.
	if s.read(); s.peekIdent() {
		typ = DIMENSION
		unit = s.scanName()
		repr += unit
		return
	} else {
		s.unread(1)
	}

	// If the number is followed by a percent sign then return a percentage.
	if ch := s.read(); ch == '%' {
		typ = PERCENTAGE
		repr += "%"
		return
	} else {
		s.unread(1)
	}

	// Otherwise return a number token.
	typ = NUMBER
	return
}

// scanNumber consumes a number.
func (s *Scanner) scanNumber() (num float64, flag, repr string) {
	var buf bytes.Buffer
	flag = "integer"

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
			flag = "number"
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
				flag = "number"
				_, _ = buf.WriteRune(ch0)
				_, _ = buf.WriteRune(ch1)
				_, _ = buf.WriteRune(ch2)
			} else {
				s.unread(3)
			}
		} else if isDigit(ch1) {
			flag = "number"
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
func (s *Scanner) scanHash() (typ TokenType, value, flag string) {
	// If there is a name following the hash then we have a hash token.
	if ch := s.read(); isName(ch) || s.peekEscape() {
		flag = "unrestricted"

		// If the name is an identifier then change the type.
		if s.peekIdent() {
			flag = "id"
		}
		return HASH, s.scanName(), flag
	}

	// If there is no name following the hash symbol then return delim-token.
	s.unread(1)
	return DELIM, "#", ""
}

// scanSuffixMatch consumes a suffix-match token.
func (s *Scanner) scanSuffixMatch() (TokenType, string) {
	if next := s.read(); next == '=' {
		return SUFFIXMATCH, ""
	}
	s.unread(1)
	return DELIM, "$"
}

// scanSubstringMatch consumes a substring-match token.
func (s *Scanner) scanSubstringMatch() (TokenType, string) {
	if next := s.read(); next == '=' {
		return SUBSTRINGMATCH, ""
	}
	s.unread(1)
	return DELIM, "*"
}

// scanPrefixMatch consumes a prefix-match token.
func (s *Scanner) scanPrefixMatch() (TokenType, string) {
	if next := s.read(); next == '=' {
		return PREFIXMATCH, ""
	}
	s.unread(1)
	return DELIM, "^"
}

// scanIncludeMatch consumes an include-match token.
func (s *Scanner) scanIncludeMatch() (TokenType, string) {
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
func (s *Scanner) scanIdent() (TokenType, string) {
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
func (s *Scanner) scanURL() (TokenType, string) {
	// Consume all whitespace after the "(".
	if ch := s.read(); isWhitespace(ch) {
		s.scanWhitespace()
	} else {
		s.unread(1)
	}

	// Read the first non-whitespace character.
	// If it starts with a single or double quote then consume a string and
	// use the string's value as the URL.
	if ch := s.read(); ch == eof {
		return URL, ""
	} else if ch == '"' || ch == '\'' {
		// Scan the string as the value.
		tok, value, _ := s.scanString()

		// Scanning a bad-string causes a bad-url token.
		if tok == BADSTRING {
			s.scanBadURL()
			return BADURL, ""
		}

		// Scan whitespace after the string.
		if ch := s.read(); isWhitespace(ch) {
			s.scanWhitespace()
		}
		s.unread(1)

		// Scan right parenthesis.
		if ch := s.read(); ch != ')' && ch != eof {
			s.scanBadURL()
			return BADURL, ""
		}
		return URL, value
	}
	s.unread(1)

	// If we have a non-quote character then scan all non-whitespace, non-quote
	// and non-lparen code points to form the URL value.
	var buf bytes.Buffer
	for {
		ch := s.read()
		if ch == ')' || ch == eof {
			return URL, buf.String()
		} else if isWhitespace(ch) {
			s.scanWhitespace()
			if ch0 := s.read(); ch0 == ')' || ch0 == eof {
				return URL, buf.String()
			} else {
				s.scanBadURL()
				return BADURL, ""
			}
		} else if ch == '"' || ch == '\'' || ch == '(' || isNonPrintable(ch) {
			s.Errors = append(s.Errors, &Error{Message: fmt.Sprintf("invalid url code point: %c (%U)", ch, ch), Pos: s.pos})
			s.scanBadURL()
			return BADURL, ""
		} else if ch == '\\' {
			if s.peekEscape() {
				_, _ = buf.WriteRune(s.scanEscape())
			} else {
				s.Errors = append(s.Errors, &Error{Message: "unescaped \\ in url", Pos: s.pos})
				s.scanBadURL()
				return BADURL, ""
			}
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}
}

// scanBadURL recovers the scanner from a malformed URL token.
// We simply consume all non-) and non-eof characters and escaped code points.
// This function does not return anything.
func (s *Scanner) scanBadURL() {
	for {
		ch := s.read()
		if ch == ')' || ch == eof {
			return
		} else if s.peekEscape() {
			s.scanEscape()
		}
	}
}

// scanUnicodeRange consumes a unicode-range token.
func (s *Scanner) scanUnicodeRange() (start, end int) {
	var buf bytes.Buffer

	// Consume up to 6 hex digits first.
	for i := 0; i < 6; i++ {
		if ch := s.read(); isHexDigit(ch) {
			_, _ = buf.WriteRune(ch)
		} else {
			s.unread(1)
			break
		}
	}

	// Consume question marks to total 6 characters (hex digits + question marks).
	n := buf.Len()
	for i := 0; i < 6-n; i++ {
		if ch := s.read(); ch == '?' {
			_, _ = buf.WriteRune(ch)
		} else {
			s.unread(1)
			break
		}
	}

	// If we have any question marks then calculate the range.
	// To calculate the range, we replace "?" with "0" for the start and
	// we replace "?" with "F" for the end.
	if buf.Len() > n {
		start64, _ := strconv.ParseInt(strings.Replace(buf.String(), "?", "0", -1), 16, 0)
		end64, _ := strconv.ParseInt(strings.Replace(buf.String(), "?", "F", -1), 16, 0)
		start, end = int(start64), int(end64)
		return
	}

	// Otherwise calculate this token is the start of the range.
	start64, _ := strconv.ParseInt(buf.String(), 16, 0)
	start = int(start64)

	// If the next two code points are a "-" and a hex digit then consume the end.
	ch1, ch2 := s.read(), s.read()
	if ch1 == '-' && isHexDigit(ch2) {
		s.unread(1)

		// Consume up to 6 hex digits for the ending range.
		buf.Reset()
		for i := 0; i < 6; i++ {
			if ch := s.read(); isHexDigit(ch) {
				_, _ = buf.WriteRune(ch)
			} else {
				s.unread(1)
				break
			}
		}
		end64, _ := strconv.ParseInt(buf.String(), 16, 0)
		end = int(end64)
		return
	}
	s.unread(2)

	// Otherwise set the end value to the start value.
	end = start
	return
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
	if s.curr() != '\\' {
		return false
	}

	// If the next code point is a newline then this is not an escape.
	next := s.read()
	s.unread(1)
	return next != '\n'
}

// peekIdent checks if the next code points are a valid identifier.
func (s *Scanner) peekIdent() bool {
	if s.curr() == '-' {
		ch := s.read()
		s.unread(1)
		return isNameStart(ch) || s.peekEscape()
	} else if isNameStart(s.curr()) {
		return true
	} else if s.curr() == '\\' && s.peekEscape() {
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
		ch = eof
	} else {
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

// curr reads the current code point.
func (s *Scanner) curr() rune {
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

// isNonPrintable returns true if the character is non-printable.
func isNonPrintable(ch rune) bool {
	return (ch >= '\u0000' && ch <= '\u0008') || ch == '\u000B' || (ch >= '\u000E' && ch <= '\u001F') || ch == '\u007F'
}
