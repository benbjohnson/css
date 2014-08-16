package css

import (
	"bytes"
	"fmt"
	"io"
)

// TODO(benbjohnson): Allow collapsing whitespace.

// Printer represents a configurable CSS printer.
type Printer struct{}

func (p *Printer) Print(w io.Writer, n Node) (err error) {
	switch n := n.(type) {
	case *StyleSheet:
		if n == nil {
			return nil
		}
		for i, r := range n.Rules {
			if i > 0 {
				_, err = w.Write([]byte{' '})
			}
			_ = p.Print(w, r)
		}

	case Rules:
		if n == nil {
			return nil
		}
		for i, r := range n {
			if i > 0 {
				_, _ = w.Write([]byte{' '})
			}
			err = p.Print(w, r)
		}

	case *AtRule:
		if n == nil {
			return nil
		}
		_, _ = w.Write([]byte{'@'})
		_, _ = w.Write([]byte(n.Name))
		if len(n.Prelude) > 0 {
			_ = p.Print(w, n.Prelude)
		}
		if n.Block != nil {
			err = p.Print(w, n.Block)
		} else {
			_, err = w.Write([]byte{';'})
		}

	case *QualifiedRule:
		if n == nil {
			return nil
		}
		_ = p.Print(w, n.Prelude)
		err = p.Print(w, n.Block)

	case *Declaration:
		if n == nil {
			return nil
		}
		_, _ = w.Write([]byte(n.Name))
		_, _ = w.Write([]byte{':'})
		err = p.Print(w, n.Values)
		if n.Important {
			_, err = w.Write([]byte("!important"))
		}

	case Declarations:
		if n == nil {
			return nil
		}
		for i, v := range n {
			if i > 0 {
				_, _ = w.Write([]byte{' '})
			}
			_ = p.Print(w, v)
			_, err = w.Write([]byte{';'})
		}

	case ComponentValues:
		if n == nil {
			return nil
		}
		for _, v := range n {
			err = p.Print(w, v)
		}

	case *SimpleBlock:
		if n == nil {
			return nil
		}
		switch n.Token.Tok {
		case LBraceToken:
			_, _ = w.Write([]byte{'{'})
		case LBrackToken:
			_, _ = w.Write([]byte{'['})
		case LParenToken:
			_, _ = w.Write([]byte{'('})
		}

		_ = p.Print(w, n.Values)

		switch n.Token.Tok {
		case LBraceToken:
			_, _ = w.Write([]byte{'}'})
		case LBrackToken:
			_, _ = w.Write([]byte{']'})
		case LParenToken:
			_, _ = w.Write([]byte{')'})
		}

	case *Function:
		if n == nil {
			return nil
		}
		_, _ = w.Write([]byte(n.Name))
		_, _ = w.Write([]byte{'('})
		_ = p.Print(w, n.Values)
		_, err = w.Write([]byte{')'})

	case *Token:
		if n == nil {
			return nil
		}
		switch n.Tok {
		case IdentToken:
			_, err = w.Write([]byte(n.Value))
		case FunctionToken:
			_, err = w.Write([]byte(n.Value + "("))
		case AtKeywordToken:
			_, err = w.Write([]byte("@" + n.Value))
		case HashToken:
			_, err = w.Write([]byte("#" + n.Value))
		case StringToken:
			_, err = w.Write([]byte(string(n.Ending) + n.Value + string(n.Ending)))
		case BadStringToken:
			_, err = w.Write([]byte("''"))
		case URLToken:
			_, err = w.Write([]byte("url(" + n.Value + ")"))
		case BadURLToken:
			_, err = w.Write([]byte("url()"))
		case DelimToken, NumberToken, PercentageToken, DimensionToken, WhitespaceToken:
			_, err = w.Write([]byte(n.Value))
		case UnicodeRangeToken:
			if n.Start == n.End {
				_, err = fmt.Fprintf(w, "U+%06x", n.Start)
			} else {
				_, err = fmt.Fprintf(w, "U+%06x-U+%06x", n.Start, n.End)
			}
		case IncludeMatchToken:
			_, err = w.Write([]byte("~="))
		case DashMatchToken:
			_, err = w.Write([]byte("|="))
		case PrefixMatchToken:
			_, err = w.Write([]byte("^="))
		case SuffixMatchToken:
			_, err = w.Write([]byte("$="))
		case SubstringMatchToken:
			_, err = w.Write([]byte("*="))
		case ColumnToken:
			_, err = w.Write([]byte("||"))
		case CDOToken:
			_, err = w.Write([]byte("<!--"))
		case CDCToken:
			_, err = w.Write([]byte("-->"))
		case ColonToken:
			_, err = w.Write([]byte{':'})
		case SemicolonToken:
			_, err = w.Write([]byte{';'})
		case CommaToken:
			_, err = w.Write([]byte{','})
		case LBrackToken:
			_, err = w.Write([]byte{'['})
		case RBrackToken:
			_, err = w.Write([]byte{']'})
		case LParenToken:
			_, err = w.Write([]byte{'('})
		case RParenToken:
			_, err = w.Write([]byte{')'})
		case LBraceToken:
			_, err = w.Write([]byte{'{'})
		case RBraceToken:
			_, err = w.Write([]byte{'}'})
		case EOFToken:
			_, err = w.Write([]byte("EOF"))
		}
	}

	return
}

// print pretty prints an AST node to a string using the default configuration.
func print(n Node) string {
	var p Printer
	var buf bytes.Buffer
	_ = p.Print(&buf, n)
	return buf.String()
}
