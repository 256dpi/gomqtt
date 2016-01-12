package transport

import "fmt"

type ErrorCode int

const(
	_ ErrorCode = iota
	ExpectedClose
	DialError
	EncodeError
	DecodeError
	DetectionError
	ConnectionError
	ReadLimitExceeded
)

type Error interface {
	error

	// Code will return the corresponding error code.
	Code() ErrorCode

	// Err will return the original error.
	Err() error
}

type transportError struct {
	code ErrorCode
	err error
}

func newTransportError(code ErrorCode, err error) *transportError {
	return &transportError{
		code: code,
		err: err,
	}
}

func (err *transportError) Error() string {
	switch err.code {
	case ExpectedClose:
		return fmt.Sprintf("expected close: %s", err.err.Error())
	case DialError:
		return fmt.Sprintf("dial error: %s", err.err.Error())
	case EncodeError:
		return fmt.Sprintf("encode error: %s", err.err.Error())
	case DecodeError:
		return fmt.Sprintf("decode error: %s", err.err.Error())
	case DetectionError:
		return fmt.Sprintf("detection error: %s", err.err.Error())
	case ConnectionError:
		return fmt.Sprintf("connection error: %s", err.err.Error())
	case ReadLimitExceeded:
		return fmt.Sprintf("read limit exceeded: %s", err.err.Error())
	}

	return fmt.Sprintf("unknown error: %s", err.err.Error())
}

func (err *transportError) Code() ErrorCode {
	return err.code
}

func (err *transportError) Err() error {
	return err.err
}
