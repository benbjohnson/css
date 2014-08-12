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

// Tokenizer implements a CSS3 standard compliant tokenizer.
// The tokenizer will always return a Token pointer on Scan but returns
// a ComponentValue to comply with the Scanner interface.
//
// This implementation only allows UTF-8 encoding.
// @charset directives will be ignored.
type Tokenizer struct {
	// Errors contains a list of all errors that occur during scanning.
	Errors []*Error

	rd  io.RuneReader
	pos Pos

	tokbuf  *Token // last token read from the tokenizer.
	tokbufn bool   // whether the token buffer is in use.

	buf    [4]rune // circular buffer for runes
	bufpos [4]Pos  // circular buffer for position
	bufi   int     // circular buffer index
	bufn   int     // number of buffered characters
}

// New returns a new instance of Tokenizer.
func NewTokenizer(r io.Reader) *Tokenizer {
	return &Tokenizer{
		rd: bufio.NewReader(r),
	}
}

// Scan returns the next token from the reader.
func (t *Tokenizer) Scan() ComponentValue {
	// If unscan was the last call then return the previous token again.
	if t.tokbufn {
		t.tokbufn = false
		return t.tokbuf
	}

	// Otherwise read from the reader and save the token.
	tok := t.scan()
	t.tokbuf = tok
	return tok
}

func (t *Tokenizer) scan() *Token {
	for {
		// Read next code point.
		ch := t.read()
		pos := t.Pos()

		if ch == eof {
			return &Token{Tok: EOFToken, Pos: pos}
		} else if isWhitespace(ch) {
			return t.scanWhitespace()
		} else if ch == '"' || ch == '\'' {
			return t.scanString()
		} else if ch == '#' {
			return t.scanHash()
		} else if ch == '$' {
			if next := t.read(); next == '=' {
				return &Token{Tok: SuffixMatchToken, Pos: pos}
			}
			t.unread(1)
			return &Token{Tok: DelimToken, Value: string(ch), Pos: pos}
		} else if ch == '*' {
			if next := t.read(); next == '=' {
				return &Token{Tok: SubstringMatchToken, Pos: pos}
			}
			t.unread(1)
			return &Token{Tok: DelimToken, Value: string(ch), Pos: pos}
		} else if ch == '^' {
			if next := t.read(); next == '=' {
				return &Token{Tok: PrefixMatchToken, Pos: pos}
			}
			t.unread(1)
			return &Token{Tok: DelimToken, Value: string(ch), Pos: pos}
		} else if ch == '~' {
			if next := t.read(); next == '=' {
				return &Token{Tok: IncludeMatchToken, Pos: pos}
			}
			t.unread(1)
			return &Token{Tok: DelimToken, Value: string(ch), Pos: pos}
		} else if ch == ',' {
			return &Token{Tok: CommaToken, Pos: pos}
		} else if ch == '-' {
			// Scan then next two tokens and unread back to the hyphen.
			ch1, ch2 := t.read(), t.read()
			t.unread(3)

			// If we have a digit next, it's a numeric token. If it's an identifier
			// then scan an identifier, and if it's a "->" then it's a CDC.
			if isDigit(ch1) || ch1 == '.' {
				return t.scanNumeric(pos)
			} else if t.peekIdent() {
				return t.scanIdent()
			} else if ch1 == '-' && ch2 == '>' {
				return &Token{Tok: CDCToken, Pos: pos}
			} else {
				return &Token{Tok: DelimToken, Value: "-", Pos: pos}
			}
		} else if ch == '/' {
			// Comments are ignored by the scanner so restart the loop from
			// the end of the comment and get the next token.
			if ch1 := t.read(); ch1 == '*' {
				t.scanComment()
				continue
			}
			t.unread(1)
			return &Token{Tok: DelimToken, Value: "/", Pos: pos}
		} else if ch == ':' {
			return &Token{Tok: ColonToken, Pos: pos}
		} else if ch == ';' {
			return &Token{Tok: SemicolonToken, Pos: pos}
		} else if ch == '<' {
			// Attempt to read a comment open ("<!--").
			// If it's not possible then then rollback and return DELIM.
			if ch0 := t.read(); ch0 == '!' {
				if ch1 := t.read(); ch1 == '-' {
					if ch2 := t.read(); ch2 == '-' {
						return &Token{Tok: CDOToken, Pos: pos}
					}
					t.unread(1)
				}
				t.unread(1)
			}
			t.unread(1)
			return &Token{Tok: DelimToken, Value: "<", Pos: pos}
		} else if ch == '@' {
			// This is an at-keyword token if an identifier follows.
			// Otherwise it's just a DELIM.
			if t.read(); t.peekIdent() {
				return &Token{Tok: AtKeywordToken, Value: t.scanName(), Pos: pos}
			}
			return &Token{Tok: DelimToken, Value: "@", Pos: pos}
		} else if ch == '(' {
			return &Token{Tok: LParenToken, Pos: pos}
		} else if ch == ')' {
			return &Token{Tok: RParenToken, Pos: pos}
		} else if ch == '[' {
			return &Token{Tok: LBrackToken, Pos: pos}
		} else if ch == ']' {
			return &Token{Tok: RBrackToken, Pos: pos}
		} else if ch == '{' {
			return &Token{Tok: LBraceToken, Pos: pos}
		} else if ch == '}' {
			return &Token{Tok: RBraceToken, Pos: pos}
		} else if ch == '\\' {
			// Return a valid escape, if possible.
			if t.peekEscape() {
				return t.scanIdent()
			}
			// Otherwise this is a parse error but continue on as a DELIM.
			t.Errors = append(t.Errors, &Error{Message: "unescaped \\", Pos: t.Pos()})
			return &Token{Tok: DelimToken, Value: "\\", Pos: pos}
		} else if ch == '+' || ch == '.' || isDigit(ch) {
			t.unread(1)
			return t.scanNumeric(pos)
		} else if ch == 'u' || ch == 'U' {
			// Peek "+[0-9a-f]" or "+?", consume next code point, consume unicode-range.
			ch1, ch2 := t.read(), t.read()
			if ch1 == '+' && (isHexDigit(ch2) || ch2 == '?') {
				t.unread(1)
				return t.scanUnicodeRange()
			}
			// Otherwise reconsume as ident.
			t.unread(2)
			return t.scanIdent()
		} else if isNameStart(ch) {
			return t.scanIdent()
		} else if ch == '|' {
			// If the next token is an equals sign, it's a dash token.
			// If the next token is a pipe, it's a column token.
			// Otherwise, just treat this pipe as a delim token.
			if ch1 := t.read(); ch1 == '=' {
				return &Token{Tok: DashMatchToken, Pos: pos}
			} else if ch1 == '|' {
				return &Token{Tok: ColumnToken, Pos: pos}
			}
			t.unread(1)
			return &Token{Tok: DelimToken, Value: string(ch), Pos: pos}
		}
		return &Token{Tok: DelimToken, Value: string(ch), Pos: pos}
	}
}

// Unscan buffers the previous scan.
func (t *Tokenizer) Unscan() {
	t.tokbufn = true
}

// Current returns the current token.
func (t *Tokenizer) Current() ComponentValue {
	return t.tokbuf
}

// scanWhitespace consumes the current code point and all subsequent whitespace.
func (t *Tokenizer) scanWhitespace() *Token {
	pos := t.Pos()
	var buf bytes.Buffer
	_, _ = buf.WriteRune(t.curr())
	for {
		ch := t.read()
		if ch == eof {
			break
		} else if !isWhitespace(ch) {
			t.unread(1)
			break
		}
		_, _ = buf.WriteRune(ch)
	}
	return &Token{Tok: WhitespaceToken, Value: buf.String(), Pos: pos}
}

// scanString consumes a quoted string. (§4.3.4)
//
// This assumes that the current token is a single or double quote.
// This function consumes all code points and escaped code points up until
// a matching, unescaped ending quote.
// An EOF closes out a string but does not return an error.
// A newline will close a string and returns a bad-string token.
func (t *Tokenizer) scanString() *Token {
	pos, ending := t.Pos(), t.curr()
	var buf bytes.Buffer
	for {
		ch := t.read()
		if ch == eof || ch == ending {
			return &Token{Tok: StringToken, Value: buf.String(), Ending: ending, Pos: pos}
		} else if ch == '\n' {
			t.unread(1)
			return &Token{Tok: BadStringToken, Pos: pos}
		} else if ch == '\\' {
			if t.peekEscape() {
				_, _ = buf.WriteRune(t.scanEscape())
				continue
			}
			if next := t.read(); next == eof {
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
func (t *Tokenizer) scanNumeric(pos Pos) *Token {
	num, typ, repr := t.scanNumber()

	// If the number is immediately followed by an identifier then scan dimension.
	if t.read(); t.peekIdent() {
		unit := t.scanName()
		return &Token{Tok: DimensionToken, Type: typ, Value: repr + unit, Number: num, Unit: unit, Pos: pos}
	} else {
		t.unread(1)
	}

	// If the number is followed by a percent sign then return a percentage.
	if ch := t.read(); ch == '%' {
		return &Token{Tok: PercentageToken, Type: typ, Value: repr + "%", Number: num, Pos: pos}
	} else {
		t.unread(1)
	}

	// Otherwise return a number token.
	return &Token{Tok: NumberToken, Type: typ, Value: repr, Number: num, Pos: pos}
}

// scanNumber consumes a number.
func (t *Tokenizer) scanNumber() (num float64, typ, repr string) {
	var buf bytes.Buffer
	typ = "integer"

	// If initial code point is + or - then store it.
	if ch := t.read(); ch == '+' || ch == '-' {
		_, _ = buf.WriteRune(ch)
	} else {
		t.unread(1)
	}

	// Read as many digits as possible.
	_, _ = buf.WriteString(t.scanDigits())

	// If next code points are a full stop and digit then consume them.
	if ch0 := t.read(); ch0 == '.' {
		if ch1 := t.read(); isDigit(ch1) {
			typ = "number"
			_, _ = buf.WriteRune(ch0)
			_, _ = buf.WriteRune(ch1)
			_, _ = buf.WriteString(t.scanDigits())
		} else {
			t.unread(2)
		}
	} else {
		t.unread(1)
	}

	// Consume scientific notation (e0, e+0, e-0, E0, E+0, E-0).
	if ch0 := t.read(); ch0 == 'e' || ch0 == 'E' {
		if ch1 := t.read(); ch1 == '+' || ch1 == '-' {
			if ch2 := t.read(); isDigit(ch2) {
				typ = "number"
				_, _ = buf.WriteRune(ch0)
				_, _ = buf.WriteRune(ch1)
				_, _ = buf.WriteRune(ch2)
			} else {
				t.unread(3)
			}
		} else if isDigit(ch1) {
			typ = "number"
			_, _ = buf.WriteRune(ch0)
			_, _ = buf.WriteRune(ch1)
		} else {
			t.unread(2)
		}
	} else {
		t.unread(1)
	}

	// Parse number.
	num, _ = strconv.ParseFloat(buf.String(), 64)
	repr = buf.String()
	return
}

// scanDigits consume a contiguous series of digits.
func (t *Tokenizer) scanDigits() string {
	var buf bytes.Buffer
	for {
		if ch := t.read(); isDigit(ch) {
			_, _ = buf.WriteRune(ch)
		} else {
			t.unread(1)
			break
		}
	}
	return buf.String()
}

// scanComment consumes all characters up to "*/", inclusive.
// This function assumes that the initial "/*" have just been consumed.
func (t *Tokenizer) scanComment() {
	for {
		ch0 := t.read()
		if ch0 == eof {
			break
		} else if ch0 == '*' {
			if ch1 := t.read(); ch1 == '/' {
				break
			} else {
				t.unread(1)
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
func (t *Tokenizer) scanHash() *Token {
	pos := t.Pos()

	// If there is a name following the hash then we have a hash token.
	if ch := t.read(); isName(ch) || t.peekEscape() {
		typ := "unrestricted"

		// If the name is an identifier then change the type.
		if t.peekIdent() {
			typ = "id"
		}
		return &Token{Tok: HashToken, Value: t.scanName(), Type: typ, Pos: pos}
	}
	t.unread(1)

	// If there is no name following the hash symbol then return delim-token.
	return &Token{Tok: DelimToken, Value: "#", Pos: pos}
}

// scanName consumes a name.
// Consumes contiguous name code points and escaped code points.
func (t *Tokenizer) scanName() string {
	var buf bytes.Buffer
	t.unread(1)
	for {
		if ch := t.read(); isName(ch) {
			_, _ = buf.WriteRune(ch)
		} else if t.peekEscape() {
			_, _ = buf.WriteRune(t.scanEscape())
		} else {
			t.unread(1)
			return buf.String()
		}
	}
}

// scanIdent consumes a ident-like token.
// This function can return an ident, function, url, or bad-url.
func (t *Tokenizer) scanIdent() *Token {
	pos := t.Pos()
	v := t.scanName()

	// Check if this is the start of a url token.
	if strings.ToLower(v) == "url" {
		if ch := t.read(); ch == '(' {
			return t.scanURL(pos)
		}
		t.unread(1)
	} else if ch := t.read(); ch == '(' {
		return &Token{Tok: FunctionToken, Value: v, Pos: pos}
	}
	t.unread(1)

	return &Token{Tok: IdentToken, Value: v, Pos: pos}
}

// scanURL consumes the contents of a URL function.
// This function assumes that the "url(" has just been consumed.
// This function can return a url or bad-url token.
func (t *Tokenizer) scanURL(pos Pos) *Token {
	// Consume all whitespace after the "(".
	if ch := t.read(); isWhitespace(ch) {
		t.scanWhitespace()
	} else {
		t.unread(1)
	}

	// Read the first non-whitespace character.
	// If it starts with a single or double quote then consume a string and
	// use the string's value as the URL.
	if ch := t.read(); ch == eof {
		return &Token{Tok: URLToken, Pos: pos}
	} else if ch == '"' || ch == '\'' {
		// Scan the string as the value.
		tok := t.scanString()

		// Scanning a bad-string causes a bad-url token.
		var value string
		if tok.Tok == StringToken {
			value = tok.Value
		} else if tok.Tok == BadStringToken {
			t.scanBadURL()
			return &Token{Tok: BadURLToken, Pos: pos}
		}

		// Scan whitespace after the string.
		if ch := t.read(); isWhitespace(ch) {
			t.scanWhitespace()
		}
		t.unread(1)

		// Scan right parenthesis.
		if ch := t.read(); ch != ')' && ch != eof {
			t.scanBadURL()
			return &Token{Tok: BadURLToken, Pos: pos}
		}
		return &Token{Tok: URLToken, Value: value, Pos: pos}
	}
	t.unread(1)

	// If we have a non-quote character then scan all non-whitespace, non-quote
	// and non-lparen code points to form the URL value.
	var buf bytes.Buffer
	for {
		ch := t.read()
		if ch == ')' || ch == eof {
			return &Token{Tok: URLToken, Value: buf.String(), Pos: pos}
		} else if isWhitespace(ch) {
			t.scanWhitespace()
			if ch0 := t.read(); ch0 == ')' || ch0 == eof {
				return &Token{Tok: URLToken, Value: buf.String(), Pos: pos}
			} else {
				t.scanBadURL()
				return &Token{Tok: BadURLToken, Pos: pos}
			}
		} else if ch == '"' || ch == '\'' || ch == '(' || isNonPrintable(ch) {
			t.Errors = append(t.Errors, &Error{Message: fmt.Sprintf("invalid url code point: %c (%U)", ch, ch), Pos: pos})
			t.scanBadURL()
			return &Token{Tok: BadURLToken, Pos: pos}
		} else if ch == '\\' {
			if t.peekEscape() {
				_, _ = buf.WriteRune(t.scanEscape())
			} else {
				t.Errors = append(t.Errors, &Error{Message: "unescaped \\ in url", Pos: t.Pos()})
				t.scanBadURL()
				return &Token{Tok: BadURLToken, Pos: pos}
			}
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}
}

// scanBadURL recovers the scanner from a malformed URL token.
// We simply consume all non-) and non-eof characters and escaped code points.
// This function does not return anything.
func (t *Tokenizer) scanBadURL() {
	for {
		ch := t.read()
		if ch == ')' || ch == eof {
			return
		} else if t.peekEscape() {
			t.scanEscape()
		}
	}
}

// scanUnicodeRange consumes a unicode-range token.
func (t *Tokenizer) scanUnicodeRange() *Token {
	var buf bytes.Buffer

	// Move the position back one since the "U" is already consumed.
	pos := t.Pos()
	pos.Char--

	// Consume up to 6 hex digits first.
	for i := 0; i < 6; i++ {
		if ch := t.read(); isHexDigit(ch) {
			_, _ = buf.WriteRune(ch)
		} else {
			t.unread(1)
			break
		}
	}

	// Consume question marks to total 6 characters (hex digits + question marks).
	n := buf.Len()
	for i := 0; i < 6-n; i++ {
		if ch := t.read(); ch == '?' {
			_, _ = buf.WriteRune(ch)
		} else {
			t.unread(1)
			break
		}
	}

	// If we have any question marks then calculate the range.
	// To calculate the range, we replace "?" with "0" for the start and
	// we replace "?" with "F" for the end.
	if buf.Len() > n {
		start64, _ := strconv.ParseInt(strings.Replace(buf.String(), "?", "0", -1), 16, 0)
		end64, _ := strconv.ParseInt(strings.Replace(buf.String(), "?", "F", -1), 16, 0)
		return &Token{Tok: UnicodeRangeToken, Start: int(start64), End: int(end64), Pos: pos}
	}

	// Otherwise calculate this token is the start of the range.
	start64, _ := strconv.ParseInt(buf.String(), 16, 0)

	// If the next two code points are a "-" and a hex digit then consume the end.
	ch1, ch2 := t.read(), t.read()
	if ch1 == '-' && isHexDigit(ch2) {
		t.unread(1)

		// Consume up to 6 hex digits for the ending range.
		buf.Reset()
		for i := 0; i < 6; i++ {
			if ch := t.read(); isHexDigit(ch) {
				_, _ = buf.WriteRune(ch)
			} else {
				t.unread(1)
				break
			}
		}
		end64, _ := strconv.ParseInt(buf.String(), 16, 0)
		return &Token{Tok: UnicodeRangeToken, Start: int(start64), End: int(end64), Pos: pos}
	}
	t.unread(2)

	// Otherwise set the end value to the start value.
	return &Token{Tok: UnicodeRangeToken, Start: int(start64), End: int(start64), Pos: pos}
}

// scanEscape consumes an escaped code point.
func (t *Tokenizer) scanEscape() rune {
	var buf bytes.Buffer
	ch := t.read()
	if isHexDigit(ch) {
		_, _ = buf.WriteRune(ch)
		for i := 0; i < 5; i++ {
			if next := t.read(); next == eof || isWhitespace(next) {
				break
			} else if !isHexDigit(next) {
				t.unread(1)
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
func (t *Tokenizer) peekEscape() bool {
	// If the current code point is not a backslash then this is not an escape.
	if t.curr() != '\\' {
		return false
	}

	// If the next code point is a newline then this is not an escape.
	next := t.read()
	t.unread(1)
	return next != '\n'
}

// peekIdent checks if the next code points are a valid identifier.
func (t *Tokenizer) peekIdent() bool {
	if t.curr() == '-' {
		ch := t.read()
		t.unread(1)
		return isNameStart(ch) || t.peekEscape()
	} else if isNameStart(t.curr()) {
		return true
	} else if t.curr() == '\\' && t.peekEscape() {
		return true
	}
	return false
}

// read reads the next rune from the reader.
// This function will initially check for any characters that have been pushed
// back onto the lookahead buffer and return those. Otherwise it will read from
// the reader and do preprocessing to convert newline characters and NULL.
// EOF is converted to a zero rune (\000) and returned.
func (t *Tokenizer) read() rune {
	// If we have runes on our internal lookahead buffer then return those.
	if t.bufn > 0 {
		t.bufi = ((t.bufi + 1) % len(t.buf))
		t.bufn--
		return t.buf[t.bufi]
	}

	// Otherwise read from the reader.
	ch, _, err := t.rd.ReadRune()
	pos := t.Pos()
	if err != nil {
		ch = eof
	} else {
		// Preprocess the input stream by replacing FF with LF. (§3.3)
		if ch == '\f' {
			ch = '\n'
		}

		// Preprocess the input stream by replacing CR and CRLF with LF. (§3.3)
		if ch == '\r' {
			if ch, _, err := t.rd.ReadRune(); err != nil {
				// nop
			} else if ch != '\n' {
				t.unread(1)
			}
			ch = '\n'
		}

		// Replace NULL with Unicode replacement character. (§3.3)
		if ch == '\000' {
			ch = '\uFFFD'
		}

		// Track scanner position.
		if ch == '\n' {
			pos.Line++
			pos.Char = 0
		} else {
			pos.Char++
		}
	}

	// Add to circular buffer.
	t.bufi = ((t.bufi + 1) % len(t.buf))
	t.buf[t.bufi] = ch
	t.bufpos[t.bufi] = pos
	return ch
}

// unread adds the previous n code points back onto the buffer.
func (t *Tokenizer) unread(n int) {
	for i := 0; i < n; i++ {
		t.bufi = ((t.bufi + len(t.buf) - 1) % len(t.buf))
		t.bufn++
	}
}

// curr reads the current code point.
func (t *Tokenizer) curr() rune {
	return t.buf[t.bufi]
}

// Pos reads the current position of the scanner.
func (t *Tokenizer) Pos() Pos {
	return t.bufpos[t.bufi]
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
