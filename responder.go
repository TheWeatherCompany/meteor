package meteor

import (
	"net/http"
)

type Request interface {
	GetRequest() *http.Request
	GetSuccess() interface{}
	GetFailure() interface{}
	GetError() error
}

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
	//GetError() error
}

