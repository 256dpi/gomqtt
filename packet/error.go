package packet

import "fmt"

// Error represents decoding and encoding errors.
type Error struct {
	format    string
	arguments []interface{}
}

func makeError(format string, arguments ...interface{}) *Error {
	return &Error{format, arguments}
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf(e.format, e.arguments...)
}
