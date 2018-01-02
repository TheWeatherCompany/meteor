package meteor

import (
	"fmt"
	"net/http"
)

type Common struct{}

func (C *Common) Do(c *Service) error {
	req, err := c.Request()
	if err != nil {
		return err
	}

	//fmt.Println(req.URL.String())
	_, err = c.Do(req)
	if err != nil {
		return err
	}
	return nil
}

// AsyncRequest is the type for asynchronous requests.
// Implements Request interface
type AsyncRequest struct {
	Responder Responder
	Request *http.Request
	Error    error
}

// GetRequest gets the request.
func (ar *AsyncRequest) GetRequest() *http.Request {
	return ar.Request
}

// GetSuccess gets the success struct.
func (ar *AsyncRequest) GetSuccess() interface{} {
	return ar.Responder.GetSuccess()
}

// GetFailure gets the failure struct.
func (ar *AsyncRequest) GetFailure() interface{} {
	return ar.Responder.GetFailure()
}

// GetError gets the error.
func (ar *AsyncRequest) GetError() error {
	return ar.Error
}

// AsyncResponse is the type for asynchronous responses.
// Implements Response interface.
type AsyncResponse struct {
	responder Responder
	Response *http.Response
	Success  interface{}
	Failure  interface{}
	Error    error
}

// GetRequest gets the request.
func (ar *AsyncResponse) GetRequest() *http.Response {
	return ar.Response
}

// GetSuccess gets the success struct.
func (ar *AsyncResponse) GetSuccess() interface{} {
	return ar.Success
}

// GetFailure gets the failure struct.
func (ar *AsyncResponse) GetFailure() interface{} {
	return ar.Failure
}

// GetError gets the error.
func (ar *AsyncResponse) GetError() error {
	return ar.Error
}
