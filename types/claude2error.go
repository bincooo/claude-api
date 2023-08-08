package types

import "fmt"

type Claude2Error struct {
	ErrorType Claude2ErrorType `json:"error"`
	Detail    string           `json:"detail"`
}

type Claude2ErrorType struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (c Claude2Error) Error() string {
	return fmt.Sprintf("[Claude2Error::%s]%s: %s", c.ErrorType.Type, c.ErrorType.Message, c.Detail)
}
