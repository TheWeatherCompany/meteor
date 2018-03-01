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
		isOk: isOk,
	}
}

// BinaryResponder creates a binary response with Failure and Success.
func BinaryFailureResponder(failure interface{}) *binaryResponder {
	return &binaryResponder{
		Failure: failure,
		Success: &[]byte{},
		isOk: isOk,
	}
}

// BinaryResponder creates a binary response with Failure and Success.
func BinaryResponder(failure interface{}, isOKfn ...func(int, *http.Response) bool) *binaryResponder {
	br := BinaryFailureResponder(failure)

	if len(isOKfn) > 0 {
		br.isOk = isOKfn[0]
	}

	return br
}

// binaryResponder
type binaryResponder responder

// isOk determines whether the HTTP Status Code is an OK Code (200-299)
// Uses isOK
func (r *binaryResponder) IsOK(statusCode int, resp *http.Response) bool {
	if r.isOk != nil {
		return r.isOk(statusCode, resp)
	}
	return isOk(statusCode, resp)
}

// Respond creates the proper response object.
func (r *binaryResponder) Respond(req *http.Request, resp *http.Response, err error) Responder {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Request = req
	r.Response = resp
	r.Error = err

	return r
}

// DoResponse does the actual response falling back on JSONResponse for errors.
func (r *binaryResponder) DoResponse() (*http.Response, error) {
	r.mu.Lock()
	defer func(br *binaryResponder) {
		br.Response.Body.Close()
		br.mu.Unlock()
	}(r)

	ok := r.isOk(r.Response.StatusCode, r.Response)
	if ok {
		r.Success, r.Error = ioutil.ReadAll(r.Response.Body)
	} else if r.Failure != nil {
		r.Error = decodeResponseJSON(r.IsOK, r.Response, nil, r.Failure)
	}

	return r.Response, r.Error
}

// GetResponse gets the http response.
func (r *binaryResponder) GetResponse() *http.Response {
	return r.Response
}

// GetSuccess gets the success struct.
func (r *binaryResponder) GetSuccess() interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.Success
}

// GetFailure gets the failure struct.
func (r *binaryResponder) GetFailure() interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.Failure
}

// GetError gets the error field.
func (r *binaryResponder) GetError() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Error
}
