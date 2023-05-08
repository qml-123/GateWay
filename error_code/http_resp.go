package error_code

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

var (
	InternalError = NewStatus(1, "internal error")

	InvalidParam = NewStatus(10, "invalid param")
)
