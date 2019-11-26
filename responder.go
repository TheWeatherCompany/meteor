package meteor

//go:generate moq -out responder_mocks_test.go . Responder

import (
	"net/http"
	"sync"
)

// ResponseProvider provides a modifier for the response.
type Responder interface {
	Respond(*http.Request, *http.Response, error) Responder
	DoResponse() (*http.Response, error)
	GetResponse() *http.Response
	GetSuccess() interface{}
	GetFailure() interface{}
	IsOK(int, *http.Response) bool
	GetError() error
}

// responder
type responder struct {
	isOk     func(int, *http.Response) bool
	mu       sync.RWMutex
	Request  *http.Request
	Response *http.Response
	Error    error
	Failure  interface{}
	Success  interface{}
}

// isOk determines whether the HTTP Status Code is an OK Code (200-299)
// Uses isOK
func (r *responder) IsOK(statusCode int, resp *http.Response) bool {
	if r.isOk != nil {
		return r.isOk(statusCode, resp)
	}
	return isOk(statusCode, resp)
}

// Respond creates the proper response object.
func (r *responder) Respond(req *http.Request, resp *http.Response, err error) Responder {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Request = req
	r.Response = resp
	r.Error = err

	return r
}

// DoResponse does the actual response.
func (r *responder) DoResponse() (*http.Response, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Response, r.Error
}

// GetResponse gets the http response.
func (r *responder) GetResponse() *http.Response {
	return r.Response
}

// GetSuccess gets the success struct.
func (r *responder) GetSuccess() interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Success
}

// GetFailure gets the failure struct.
func (r *responder) GetFailure() interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Failure
}

// GetError gets the error field.
func (r *responder) GetError() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Error
}
