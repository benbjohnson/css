package css

import (
	"fmt"
	"os"
)

// Error represents a parse error.
type Error struct {
	Message string
	Pos     Pos
}

// Error returns the formatted string error message.
func (e *Error) Error() string {
	return e.Message
}

func warn(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
}

func warnf(msg string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", v...)
}
