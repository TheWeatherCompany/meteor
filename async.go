package meteor

//go:generate moq -out async_mocks_test.go . AsyncDoer

import "net/http"

// AsyncDoer does the work for an async job/task
// for use by async
type AsyncDoer interface {
	// Prepare prepares the AsyncDoer object to do the work.
	// This is where you would set up your object on a struct
	// in order to actually do the work.
	Prepare(int)

	// Do does the work of the AsyncDoer.
	Do() interface{}

	// ToStop outputs a string that declares whether the channel should stop.
	// Returning an empty string will not stop the channel.
	// Returning a string will record on the async struct.
	ToStop() string
}

// NewAsyncDoers is a nice wrapper for creating AsyncDoer slice.
func NewAsyncDoers(doers ...AsyncDoer) []AsyncDoer {
	return doers
}

// async
type async struct {
	service   *Service
	toDos     []AsyncDoer
	responses []interface{}
	ch        chan interface{}
	stopCh    chan struct{}
	toStop    chan string
	closed    bool
	StoppedBy string
	length    int
}

// IsClosed tells you whether the channel has been closed or not.
func (a *async) IsClosed() bool {
	return a.closed
}

// Moderator moderates stopping to ensure async channel(s) close.
func (a *async) Moderator() {
	a.StoppedBy = <-a.toStop
	close(a.stopCh)
	a.closed = true
}

// Restart restarts the toStop channel.
func (a *async) Restart() {
	a.closed = false
	a.toStop = make(chan string, 1)
	go a.Moderator()
}

// do does the work of the job/task by calling the AsyncDoer methods
// gracefully closing channels.
func (a *async) do(index int, item AsyncDoer) {
	select {
	case <-a.stopCh:
		return
	default:
	}

	item.Prepare(index)
	value := item.Do()
	if toStop := item.ToStop(); toStop != "" {
		select {
		case a.toStop <- toStop:
		default:
		}
		return
	}

	select {
	case <-a.stopCh:
		return
	case a.ch <- value:
	}
}

// Do performs the toDos asyncronously.
func (a *async) Do() []interface{} {
	for i, item := range a.toDos {
		go a.do(i, item)
	}

	for {
		select {
		case <-a.stopCh:
			return a.responses
		default:
		}

		select {
		case <-a.stopCh:
			return a.responses
		case r := <-a.ch:
			a.responses = append(a.responses, r)
			if len(a.responses) == a.length {
				select {
				case a.toStop <- "asyncAutoStopped":
				default:
				}
				return a.responses
			}
		}
	}
}

// GetResponses returns the responses of the AsyncDoers.
func (a *async) GetResponses() []interface{} {
	return a.responses
}

// NewAsync creates a new async service.
func NewAsync(service *Service, toDos []AsyncDoer, length ...int) *async {
	var l int
	if len(length) == 0 || (len(length) == 1 && length[0] == 0) {
		l = len(toDos)
	} else {
		l = length[0]
	}

	a := &async{
		service:   service,
		toDos:     toDos,
		responses: make([]interface{}, 0),
		ch:        make(chan interface{}, l),
		stopCh:    make(chan struct{}),
		toStop:    make(chan string, 1),
		closed:    false,
		length:    l,
	}
	go a.Moderator()
	return a
}

// AsyncRequest is the type for asynchronous requests.
// Implements Request interface
type AsyncRequest struct {
	responder Responder
	Request   *http.Request
	Error     error
	service   *Service
	response  *AsyncResponse
}

// NewAsyncRequest creates a new AsyncRequest
func NewAsyncRequest(service *Service) *AsyncRequest {
	req, err := service.Request()
	return &AsyncRequest{
		responder: service.GetResponder(),
		Request:   req,
		Error:     err,
		service:   service,
	}
}

// NewAsyncRequestWithResponder creates a new AsyncRequest with a custom responder
// that overrides the default service Responder.
func NewAsyncRequestWithResponder(service *Service, responder Responder) *AsyncRequest {
	//req, err := service.Request()
	req, err := service.New().Responder(responder).Request()
	return &AsyncRequest{
		responder: responder,
		Request:   req,
		Error:     err,
		service:   service,
	}
}

// GetRequest gets the request.
func (ar *AsyncRequest) GetRequest() *http.Request {
	return ar.Request
}

// GetSuccess gets the success struct.
func (ar *AsyncRequest) GetSuccess() interface{} {
	return ar.responder.GetSuccess()
}

// GetFailure gets the failure struct.
func (ar *AsyncRequest) GetFailure() interface{} {
	return ar.responder.GetFailure()
}

// GetError gets the error.
func (ar *AsyncRequest) GetError() error {
	return ar.Error
}

// Prepare prepares the AsyncRequest object to do the work.
// Implements asyncDoer
func (ar *AsyncRequest) Prepare(index int) {
	resp, err := ar.service.Responder(ar.responder).Do(ar.Request)
	ar.response = &AsyncResponse{
		responder: ar.service.responder,
		Response:  resp,
		Error:     err,
	}
}

// Do does the work of the AsyncRequest.
// Implements asyncDoer
func (ar *AsyncRequest) Do() interface{} {
	return ar.response
}

// ToStop outputs a string that declares whether the channel should stop.
// AsyncRequest should not stop.
// Implements asyncDoer
func (ar *AsyncRequest) ToStop() string {
	return ""
}

// AsyncResponse is the type for asynchronous responses.
// Implements Response interface.
type AsyncResponse struct {
	responder Responder
	Response  *http.Response
	Error error
}

// GetRequest gets the request.
func (ar *AsyncResponse) GetRequest() *http.Response {
	return ar.Response
}

// GetSuccess gets the success struct.
func (ar *AsyncResponse) GetSuccess() interface{} {
	return ar.responder.GetSuccess()
}

// GetFailure gets the failure struct.
func (ar *AsyncResponse) GetFailure() interface{} {
	return ar.responder.GetFailure()
}

// GetError gets the error.
func (ar *AsyncResponse) GetError() error {
	return ar.Error
}

//type async struct {
//	service   *Service
//	requests  []AsyncRequest
//	responses []*AsyncResponse
//	ch        chan *AsyncResponse
//	stopCh    chan struct{}
//	toStop    chan string
//	closed    bool
//	stoppedBy string
//}
//
//func (a *async) do(req AsyncRequest) {
//	select {
//	case <- a.stopCh:
//		return
//	default:
//	}
//
//	resp, err := a.service.Responder(req.responder).Do(req.Request)
//	value := &AsyncResponse{
//		responder: a.service.responder,
//		Response:  resp,
//		//Success:   s.responder.GetSuccess(),
//		//Failure:   s.responder.GetFailure(),
//		Error: err,
//	}
//
//	select {
//	case <- a.stopCh:
//		return
//	case a.ch <- value:
//	}
//}
//
//func (a *async) Do() []*AsyncResponse {
//	for i, req := range a.requests {
//		go a.do(req)
//	}
//
//	for {
//		select {
//		case <- a.stopCh:
//			return a.responses
//		default:
//		}
//
//		select {
//		case <- a.stopCh:
//			return a.responses
//		case r := <-a.ch:
//			a.responses = append(a.responses, r)
//			if len(a.responses) == len(a.requests) {
//				select {
//				case a.toStop <- "stopped":
//				default:
//				}
//				return a.responses
//			}
//		}
//	}
//}
//
//func newAsync(service *Service, reqs []AsyncRequest, length ...int) *async {
//	var l int
//	if len(length) == 0 || (len(length) == 1 && length[0] == 0) {
//		l = len(reqs)
//	} else {
//		l = length[0]
//	}
//
//	a := &async{
//		service:   service,
//		requests:  reqs,
//		responses: make([]*AsyncResponse, 0),
//		ch:        make(chan *AsyncResponse, l),
//		stopCh:    make(chan struct{}),
//		toStop:    make(chan string, 1),
//		closed:    false,
//		length:    l,
//	}
//	go a.Moderator()
//	return a
//}moq -out
