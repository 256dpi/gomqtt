package packet

// Method represents the called method.
type Method string

// The available methods.
const (
	ENCODE Method = "Encode"
	DECODE Method = "Decode"
)

// Error represents decoding and encoding errors.
type Error struct {
	Type     Type
	Method   Method
	Mode     Mode
	Position int
	Err      error
	Message  string
}

func wrapError(t Type, mt Method, md Mode, pos int, err error) *Error {
	return &Error{
		Type:     t,
		Method:   mt,
		Mode:     md,
		Position: pos,
		Err:      err,
	}
}

func makeError(t Type, mt Method, md Mode, pos int, msg string) *Error {
	return &Error{
		Type:     t,
		Method:   mt,
		Mode:     md,
		Position: pos,
		Message:  msg,
	}
}

// Error implements the error interface.
func (e *Error) Error() string {
	// check error
	if e.Err != nil {
		return e.Err.Error()
	}

	return e.Message
}

// Unwrap allows unwrapping a wrapped error.
func (e *Error) Unwrap() error {
	return e.Err
}
