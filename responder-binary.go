package meteor

import (
	"io/ioutil"
	"net/http"
)

/** Binary Responder */
func BinarySuccessResponder() *binaryResponder {
	return &binaryResponder{
		Success: &[]byte{},
	}
}

func BinaryResponder(failure interface{}) *binaryResponder {
	return &binaryResponder{
		Failure: failure,
		Success: &[]byte{},
	}
}

// binaryResponder
type binaryResponder struct {
	Request  *http.Request
	Response *http.Response
	Error    error
	Failure  interface{}
	Success  interface{}
}

func (r *binaryResponder) Respond(req *http.Request, resp *http.Response, err error) Responder {
	r.Request = req
	r.Response = resp
	r.Error = err

	return r
}

func (r *binaryResponder) DoResponse() (*http.Response, error) {
	defer r.Response.Body.Close()

	if r.Success != nil {
		r.Success, r.Error = ioutil.ReadAll(r.Response.Body)
	}

	if r.Error != nil {
		if r.Failure != nil {
			responder := JSONResponder(nil, r.Failure)
			r := responder.Respond(r.Request, r.Response, r.Error)
			return r.DoResponse()
		}
		return r.Response, r.Error
	}

	return r.Response, r.Error
}

func (r *binaryResponder) GetSuccess() interface{} {
	return r.Success
}

func (r *binaryResponder) GetFailure() interface{} {
	return r.Failure
}
