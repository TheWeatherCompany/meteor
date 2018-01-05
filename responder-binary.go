package meteor

import (
	"io/ioutil"
	"net/http"
)

/** Binary Responder */
// BinarySuccessResponder creates a binary response with Success.
func BinarySuccessResponder() *binaryResponder {
	return &binaryResponder{
		Success: &[]byte{},
	}
}

// BinaryResponder creates a binary response with Failure and Success.
func BinaryResponder(failure interface{}) *binaryResponder {
	return &binaryResponder{
		Failure: failure,
		Success: &[]byte{},
	}
}

// binaryResponder
type binaryResponder responder

// Respond creates the proper response object.
func (r *binaryResponder) Respond(req *http.Request, resp *http.Response, err error) Responder {
	r.Request = req
	r.Response = resp
	r.Error = err

	return r
}

// DoResponse does the actual response falling back on JSONResponse for errors.
func (r *binaryResponder) DoResponse() (*http.Response, error) {
	defer r.Response.Body.Close()

	if isOk(r.Response.StatusCode) {
		r.Success, r.Error = ioutil.ReadAll(r.Response.Body)
	} else if r.Failure != nil {
		r.Error = decodeResponseJSON(r.Response, nil, r.Failure)
	}

	return r.Response, r.Error
}

// GetSuccess gets the success struct.
func (r *binaryResponder) GetSuccess() interface{} {
	return r.Success
}

// GetFailure gets the failure struct.
func (r *binaryResponder) GetFailure() interface{} {
	return r.Failure
}
