package css

import (
	"bufio"
	"io"
	"unicode"
)

// Scanner implements a CSS standard compliant scanner.
//
// http://www.w3.org/TR/CSS2/syndata.html
type Scanner struct {
	rd  io.RuneScanner
	pos Pos

	idx int
	buf [3]rune
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	// TODO(benbjohnson): Determine fallback encoding (§3.2).
	return &Scanner{
		rd:  bufio.NewReader(r),
		idx: -1,
	}
}

// read reads the next rune from the reader.
func (s *Scanner) read() (rune, error) {
	// If we have runes on our internal lookahead buffer then return those.
	if s.idx > -1 {
		ch := s.buf[s.idx]
		s.idx--
		return ch, nil
	}

	// Otherwise read from the reader.
	ch, _, err := s.rd.ReadRune()
	if err != nil {
		return 0, err
	}

	// Preprocess the input stream by replacing FF with LF. (§3.3)
	if ch == '\f' {
		ch == '\n'
	}

	// Preprocess the input stream by replacing CR and CRLF with LF. (§3.3)
	if ch == '\r' {
		if ch, _, err := s.rd.ReadRune(); err == io.EOF {
			// nop
		} else if err != nil {
			return 0, err
		} else if ch != '\n' {
			_ = s.rd.UnreadRune()
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

	return ch, nil
}

// unread puts a run on the internal buffer.
func (s *Scanner) unread(ch rune) {
	s.idx++
	s.buf[s.idx] = ch
}
