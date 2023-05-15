package error_code

import "fmt"

type StatusError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewStatus(code int, message string) *StatusError {
	return &StatusError{
		Code:    code,
		Message: message,
	}
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("[%v] %v", e.Code, e.Message)
}

var (
	InternalError = NewStatus(1, "internal error")

	InvalidParam = NewStatus(10000, "invalid param")

	RegisterNameDuplicate = NewStatus(10001, "username is already registered")
)
