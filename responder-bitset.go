package meteor

import (
	"io/ioutil"
	"net/http"

	"github.com/willf/bitset"
)

/** Bitset Responder */
// BitsetSuccessResponder creates a bitset response with Success.
func BitsetSuccessResponder() *bitsetResponder {
	return &bitsetResponder{
		Success: &bitset.BitSet{},
		isOk:    isOk,
	}
}

// BitsetResponder creates a bitset response with Failure and Success.
func BitsetFailureResponder(failure interface{}) *bitsetResponder {
	return &bitsetResponder{
		Failure: failure,
		Success: &bitset.BitSet{},
		isOk:    isOk,
	}
}

// BitsetResponder creates a bitset response with Failure and Success.
func BitsetResponder(failure interface{}, isOKfn ...func(int, *http.Response) bool) *bitsetResponder {
	br := BitsetFailureResponder(failure)

	if len(isOKfn) > 0 {
		br.isOk = isOKfn[0]
	}

	return br
}

// bitsetResponder
type bitsetResponder responder

// isOk determines whether the HTTP Status Code is an OK Code (200-299)
// Uses isOK
func (r *bitsetResponder) IsOK(statusCode int, resp *http.Response) bool {
	if r.isOk != nil {
		return r.isOk(statusCode, resp)
	}
	return isOk(statusCode, resp)
}

// Respond creates the proper response object.
func (r *bitsetResponder) Respond(req *http.Request, resp *http.Response, err error) Responder {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Request = req
	r.Response = resp
	r.Error = err

	return r
}

// DoResponse does the actual response falling back on JSONResponse for errors.
func (r *bitsetResponder) DoResponse() (*http.Response, error) {
	r.mu.Lock()
	defer func(br *bitsetResponder) {
		br.Response.Body.Close()
		br.mu.Unlock()
	}(r)

	ok := r.isOk(r.Response.StatusCode, r.Response)
	if ok {
		var contents []byte
		contents, r.Error = ioutil.ReadAll(r.Response.Body)
		r.Success.(*bitset.BitSet).UnmarshalBinary(contents)
	} else if r.Failure != nil {
		r.Error = decodeResponseJSON(r.IsOK, r.Response, nil, r.Failure)
	}

	return r.Response, r.Error
}

// GetResponse gets the http response.
func (r *bitsetResponder) GetResponse() *http.Response {
	return r.Response
}

// GetSuccess gets the success struct.
func (r *bitsetResponder) GetSuccess() interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.Success
}

// GetFailure gets the failure struct.
func (r *bitsetResponder) GetFailure() interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.Failure
}

// GetError gets the error field.
func (r *bitsetResponder) GetError() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Error
}
