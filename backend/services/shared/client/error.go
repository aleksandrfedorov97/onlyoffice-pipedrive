package client

import (
	"errors"
	"fmt"
)

var ErrInvalidUrlFormat error = errors.New("url is not valid")

type UnexpectedStatusCodeError struct {
	Action string
	Code   int
}

func (e *UnexpectedStatusCodeError) Error() string {
	return fmt.Sprintf("could not perform zoom %s action. Status code: %d", e.Action, e.Code)
}
