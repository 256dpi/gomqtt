package packet

import "fmt"

// Error represents decoding and encoding errors.
type Error struct {
	Type      Type
	Err       error
	Format    string
	Arguments []interface{}
}

func wrapError(t Type, err error) *Error {
	return &Error{
		Type: t,
		Err:  err,
	}
}

func makeError(t Type, fmt string, args ...interface{}) *Error {
	return &Error{
		Type:      t,
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

func invalidPacketID(t Type) error {
	return makeError(t, "invalid packet id")
}
