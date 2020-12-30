package packet

import "fmt"

// Method represents the called method.
type Method string

// The available methods.
const (
	ENCODE Method = "Encode"
	DECODE Method = "Decode"
)

// Error represents decoding and encoding errors.
type Error struct {
	Type      Type
	Method    Method
	Mode      Mode
	Err       error
	Format    string
	Arguments []interface{}
}

func wrapError(t Type, mt Method, m Mode, err error) *Error {
	return &Error{
		Type:   t,
		Method: mt,
		Mode:   m,
		Err:    err,
	}
}

func makeError(t Type, mt Method, m Mode, fmt string, args ...interface{}) *Error {
	return &Error{
		Type:      t,
		Method:    mt,
		Mode:      m,
		Format:    fmt,
		Arguments: args,
	}
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	return fmt.Sprintf(e.Format, e.Arguments...)
}
