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

	fmt.Println(req.URL.String())
	//c.Responder(JSONSuccessResponder(successV))
	_, err = c.Do(req)
	//_, err = c.Do(req, &successV, nil)
	if err != nil {
		return err
	}
	return nil
}

//func (C *Common) Do(c *Service, successV interface{}) error {
//	req, err := c.Request()
//	if err != nil {
//		return err
//	}
//
//	fmt.Println(req.URL.String())
//	c.Responder(JSONSuccessResponder(successV))
//	_, err = c.Do(req)
//	//_, err = c.Do(req, &successV, nil)
//	if err != nil {
//		return err
//	}
//	return nil
//}

//func (C *Common) DoBinary(c *Service) (error, []byte) {
//
//
//	//fmt.Println(req.URL.String())
//
//	//var successV []byte
//	//c.Responder(BinarySuccessResponder(&successV))
//	err := C.Do(c)
//	//err := C.Do(c, successV)
//
//
//	//req, err := c.Request()
//	//if err != nil {
//	//	return err, nil
//	//}
//	//successV, err := c.DoBinary(req)
//
//
//
//	//if resp.Header.Get("Accept-Ranges") == "bytes" {
//	//	var buf []byte
//	//	buf, err = ioutil.ReadAll(resp.Body)
//	//	if isOk(resp.StatusCode) {
//	//		successV = buf
//	//	}
//	//}
//	//return err
//	//return err, successV
//	return err, []byte{}
//}


type AsyncRequest struct {
	Responder Responder
	Request *http.Request
	//Success interface{}
	//Failure interface{}
	Error    error
}

func (ar *AsyncRequest) GetRequest() *http.Request {
	return ar.Request
}

func (ar *AsyncRequest) GetSuccess() interface{} {
	return ar.Responder.GetSuccess()
}

func (ar *AsyncRequest) GetFailure() interface{} {
	return ar.Responder.GetFailure()
}

func (ar *AsyncRequest) GetError() error {
	return ar.Error
}

type AsyncResponse struct {
	responder Responder
	Response *http.Response
	Success  interface{}
	Failure  interface{}
	Error    error
}

func (ar *AsyncResponse) GetRequest() *http.Response {
	return ar.Response
}

func (ar *AsyncResponse) GetSuccess() interface{} {
	return ar.Success
}

func (ar *AsyncResponse) GetFailure() interface{} {
	return ar.Failure
}

func (ar *AsyncResponse) GetError() error {
	return ar.Error
}
