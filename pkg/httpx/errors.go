package httpx

import "errors"

var (
	ErrNilReader          = errors.New("nil reader")
	ErrNilResponseWriter  = errors.New("nil response writer")
	ErrNilError           = errors.New("nil error")
	ErrBodyTooLarge       = errors.New("body exceeds configured limit")
	ErrMultipleJSONValues = errors.New("multiple JSON values are not allowed")
)
