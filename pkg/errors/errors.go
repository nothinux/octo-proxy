package errors

import "fmt"

type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

func New(title, text string) error {
	errString := fmt.Sprintf("[%s] %s", title, text)
	return &errorString{errString}
}
