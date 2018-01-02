package meteor

import (
	"net/http"
)

// Request interface for requests.
type Request interface {
	GetRequest() *http.Request
	GetSuccess() interface{}
	GetFailure() interface{}
	GetError() error
}

// Response interface for responses.
type Response interface {
	GetResponse() *http.Response
	GetSuccess() interface{}
	GetFailure() interface{}
	GetError() error
}

// ResponseProvider provides a modifier for the response.
type Responder interface {
	Respond(*http.Request, *http.Response, error) (Responder)
	DoResponse() (*http.Response, error)
	GetSuccess() interface{}
	GetFailure() interface{}
}

