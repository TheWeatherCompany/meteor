package meteor

import (
	"net/http"
)

/** GENERIC Responder */
func GenericResponder() *genericResponder {
	return &genericResponder{}
}

// genericResponder
type genericResponder struct {
	Request  *http.Request
	Response *http.Response
	Error    error
}

func (r *genericResponder) Respond(req *http.Request, resp *http.Response, err error) Responder {
	r.Request = req
	r.Response = resp
	r.Error = err

	return r
}

func (r *genericResponder) DoResponse() (*http.Response, error) {
	return r.Response, r.Error
}

func (r *genericResponder) GetSuccess() interface{} {
	return nil
}

func (r *genericResponder) GetFailure() interface{} {
	return nil
}

