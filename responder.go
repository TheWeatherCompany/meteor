package meteor

//go:generate moq -out responder_mocks_test.go . Responder

import (
	"net/http"
)

// ResponseProvider provides a modifier for the response.
type Responder interface {
	Respond(*http.Request, *http.Response, error) (Responder)
	DoResponse() (*http.Response, error)
	GetSuccess() interface{}
	GetFailure() interface{}
}

// responder
type responder struct {
	Request  *http.Request
	Response *http.Response
	Error    error
	Failure  interface{}
	Success  interface{}
}


// Respond creates the proper response object.
func (r *responder) Respond(req *http.Request, resp *http.Response, err error) Responder {
	r.Request = req
	r.Response = resp
	r.Error = err

	return r
}

// DoResponse does the actual response.
func (r *responder) DoResponse() (*http.Response, error) {
	return r.Response, r.Error
}

// GetSuccess gets the success struct.
func (r *responder) GetSuccess() interface{} {
	return r.Success
}

// GetFailure gets the failure struct.
func (r *responder) GetFailure() interface{} {
	return r.Failure
}

