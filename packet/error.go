package packet

import "fmt"

// Error represents decoding and encoding errors.
type Error struct {
	Type      Type
	Mode      Mode
	Err       error
	Format    string
	Arguments []interface{}
}

func wrapError(t Type, m Mode, err error) *Error {
	return &Error{
		Type: t,
		Mode: m,
		Err:  err,
	}
}

func makeError(t Type, m Mode, fmt string, args ...interface{}) *Error {
	return &Error{
		Type:      t,
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

func invalidPacketID(t Type, m Mode) error {
	return makeError(t, m, "invalid packet id")
}
