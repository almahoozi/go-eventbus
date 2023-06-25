package eventbus

import (
	"errors"
	"strings"
)

type Errors []error

func (e Errors) Error() string {
	var strs []string
	for _, err := range e {
		strs = append(strs, err.Error())
	}
	return strings.Join(strs, "\n")
}

var (
	ErrBusClosed = errors.New("bus is closed")
	// TODO: add more errors, for example to differentiate between publishing timeout and handler timeout, etc. as well as other internal errors
)
