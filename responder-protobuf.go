package meteor

import (
	"io"
	"io/ioutil"
	"net/http"
)

/** Protobuf Responder */
// ProtobufSuccessResponder creates a protobuf response with Success.
func ProtobufSuccessResponder(success interface{}) *protobufResponder {
	return &protobufResponder{
		Success: success,
		isOk:    isOk,
	}
}

// ProtobufResponder creates a protobuf response with Failure and Success.
func ProtobufResponder(success, failure interface{}, isOKfn ...func(int, *http.Response) bool) *protobufResponder {
	jr := &protobufResponder{
		Failure: failure,
		Success: success,
		isOk:    isOk,
	}

	if len(isOKfn) > 0 {
		jr.isOk = isOKfn[0]
	}

	return jr
}

// protobufResponder
type protobufResponder responder

// isOk determines whether the HTTP Status Code is an OK Code (200-299)
// Uses isOK
func (r *protobufResponder) IsOK(statusCode int, resp *http.Response) bool {
	if r.isOk != nil {
		return r.isOk(statusCode, resp)
	}
	return isOk(statusCode, resp)
}

// Respond creates the proper response object.
func (r *protobufResponder) Respond(req *http.Request, resp *http.Response, err error) Responder {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Request = req
	r.Response = resp
	r.Error = err

	return r
}

// DoResponse does the actual response decoding from protobuf.
func (r *protobufResponder) DoResponse() (*http.Response, error) {
	if r.GetSuccess() != nil || r.GetFailure() != nil {
		r.mu.Lock()
		r.Error = decodeResponseProtobuf(r.IsOK, r.Response, r.Success, r.Failure)
		r.mu.Unlock()
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Response, r.Error
}

// GetResponse gets the http response.
func (r *protobufResponder) GetResponse() *http.Response {
	return r.Response
}

// GetSuccess gets the success struct.
func (r *protobufResponder) GetSuccess() interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Success
}

// GetFailure gets the failure struct.
func (r *protobufResponder) GetFailure() interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Failure
}

// GetError gets the error field.
func (r *protobufResponder) GetError() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Error
}

// decodeResponse decodes response Body into the value pointed to by successV
// if the response is a success (2XX) or into the value pointed to by failureV
// otherwise. If the successV or failureV argument to decode into is nil,
// decodig is skipped.
// Caller is responsible for closing the resp.Body.
func decodeResponseProtobuf(okFn func(int, *http.Response) bool, resp *http.Response, successV, failureV interface{}) (err error) {
	if okFn(resp.StatusCode, resp) && successV != nil {
		err = decodeResponseBodyProtobuf(resp, successV)
		// For some reason this is considered an ineffectual assignment
		if err != nil {
			successV, err = ioutil.ReadAll(resp.Body)
		}
		return err
	} else if failureV != nil {
		err = decodeResponseBodyProtobuf(resp, failureV)
		// For some reason this is considered an ineffectual assignment
		if err != nil {
			failureV, err = ioutil.ReadAll(resp.Body)
		}
		return err
	}
	return nil
}

// decodeResponseBodyProtobuf Protobuf decodes a Response Body into the value pointed
// to by v.
// Caller must provide a non-nil v and close the resp.Body.
func decodeResponseBodyProtobuf(resp *http.Response, v interface{}) (err error) {
	if w, ok := v.(io.Writer); ok {
		io.Copy(w, resp.Body)
	} else {
		//err = protobuf.NewDecoder(resp.Body).Decode(v)
		//if err == io.EOF {
		//	err = nil // ignore EOF errors caused by empty response body
		//}
	}
	return err
}
