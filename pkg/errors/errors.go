package errors

import (
	"fmt"
	"runtime"
	"strings"
)

type CustomError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
	Stack   string `json:"-"`
}

func (e *CustomError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func New(code int, message string) *CustomError {
	return &CustomError{
		Code:    code,
		Message: message,
		Stack:   getStackTrace(),
	}
}

func Wrap(err error, code int, message string) *CustomError {
	return &CustomError{
		Code:    code,
		Message: message,
		Err:     err,
		Stack:   getStackTrace(),
	}
}

func getStackTrace() string {
	var pc [32]uintptr
	n := runtime.Callers(3, pc[:])
	frames := runtime.CallersFrames(pc[:n])
	var builder strings.Builder

	for {
		frame, more := frames.Next()
		builder.WriteString(fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}
	return builder.String()
}
