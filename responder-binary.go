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

// GetSuccess gets the success struct.
func (r *binaryResponder) GetSuccess() interface{} {
	return r.Success
}

// GetFailure gets the failure struct.
func (r *binaryResponder) GetFailure() interface{} {
	return r.Failure
}
