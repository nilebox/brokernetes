package controller

import (
	"fmt"
	"net/http"
)

type controllerError struct {
	Code        int
	Err         string
	Description string
}

func NewConflict(reason string, v ...interface{}) error {
	return &controllerError{
		Code: http.StatusConflict,
		Err:  fmt.Sprintf(reason, v...),
	}
}

func NewGone(reason string, v ...interface{}) error {
	return &controllerError{
		Code: http.StatusGone,
		Err:  fmt.Sprintf(reason, v...),
	}
}

func NewUnprocessableEntity() error {
	return &controllerError{
		Code:        http.StatusUnprocessableEntity,
		Err:         "AsyncRequired",
		Description: "This service plan requires client support for asynchronous service operations.",
	}
}

func NewBadRequest(reason string, v ...interface{}) error {
	return &controllerError{
		Code: http.StatusBadRequest,
		Err:  fmt.Sprintf(reason, v...),
	}
}

func (c *controllerError) Error() string {
	return fmt.Sprintf("%v %v: %v %v", c.Code, http.StatusText(c.Code), c.Err, c.Description)
}

func GetControllerError(e error) *controllerError {
	switch t := e.(type) {
	case *controllerError:
		return t
	}
	return nil
}
