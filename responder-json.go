package meteor

import (
	"io"
	"io/ioutil"
	"net/http"
	"encoding/json"
)

/** JSON Responder */
// JSONSuccessResponder creates a json response with Success.
func JSONSuccessResponder(success interface{}) *jsonResponder {
	return &jsonResponder{
		Success: success,
	}
}

// JSONResponder creates a json response with Failure and Success.
func JSONResponder(success, failure interface{}) *jsonResponder {
	return &jsonResponder{
		Failure: failure,
		Success: success,
	}
}

// jsonResponder
type jsonResponder responder

// Respond creates the proper response object.
func (r *jsonResponder) Respond(req *http.Request, resp *http.Response, err error) Responder {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Request = req
	r.Response = resp
	r.Error = err

	return r
}

// DoResponse does the actual response decoding from json.
func (r *jsonResponder) DoResponse() (*http.Response, error) {
	if r.GetSuccess() != nil || r.GetFailure() != nil {
		r.mu.Lock()
		r.Error = decodeResponseJSON(r.Response, r.Success, r.Failure)
		r.mu.Unlock()
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Response, r.Error
}

// GetSuccess gets the success struct.
func (r *jsonResponder) GetSuccess() interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Success
}

// GetFailure gets the failure struct.
func (r *jsonResponder) GetFailure() interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Failure
}


// decodeResponse decodes response Body into the value pointed to by successV
// if the response is a success (2XX) or into the value pointed to by failureV
// otherwise. If the successV or failureV argument to decode into is nil,
// decoding is skipped.
// Caller is responsible for closing the resp.Body.
func decodeResponseJSON(resp *http.Response, successV, failureV interface{}) (err error) {
	if isOk(resp.StatusCode) && successV != nil {
		err = decodeResponseBodyJSON(resp, successV)
		if err != nil {
			successV, err = ioutil.ReadAll(resp.Body)
		}
		return err
	} else if failureV != nil {
		err = decodeResponseBodyJSON(resp, failureV)
		if err != nil {
			failureV, err = ioutil.ReadAll(resp.Body)
		}
		return err
	}
	return nil
}

// decodeResponseBodyJSON JSON decodes a Response Body into the value pointed
// to by v.
// Caller must provide a non-nil v and close the resp.Body.
func decodeResponseBodyJSON(resp *http.Response, v interface{}) (err error) {
	if w, ok := v.(io.Writer); ok {
		io.Copy(w, resp.Body)
	} else {
		err = json.NewDecoder(resp.Body).Decode(v)
		if err == io.EOF {
			err = nil // ignore EOF errors caused by empty response body
		}
	}
	return err
}

