package meteor

import (
	//"net/http"
)

/** GENERIC Responder */
func GenericResponder() *genericResponder {
	return &genericResponder{}
}

// genericResponder
type genericResponder struct {
	responder
}