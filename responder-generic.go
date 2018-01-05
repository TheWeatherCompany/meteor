package meteor

import (
	//"net/http"
)

/** GENERIC Responder */
func GenericResponder() *genericResponder {
	return &genericResponder{}
}

// genericResponder
type genericResponder struct {
	responder
}

//// Respond creates the proper response object.
//func (r *genericResponder) Respond(req *http.Request, resp *http.Response, err error) Responder {
//	r.Request = req
//	r.Response = resp
//	r.Error = err
//
//	return r
//}
//
//// DoResponse does the actual response.
//func (r *genericResponder) DoResponse() (*http.Response, error) {
//	return r.Response, r.Error
//}
//
//// GetSuccess gets the success struct.
//func (r *genericResponder) GetSuccess() interface{} {
//	return r.Success
//}
//
//// GetFailure gets the failure struct.
//func (r *genericResponder) GetFailure() interface{} {
//	return r.Failure
//}

