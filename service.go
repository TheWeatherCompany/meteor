package meteor

import (
	"encoding/base64"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"fmt"

	goquery "github.com/google/go-querystring/query"
)

const (
	contentType = "Content-Type"
)

// Doer executes http requests.  It is implemented by *http.Client.
// You can wrap *http.Client with layers of Doers to form a stack
// of client-side middleware.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// ResponseChecker is a function to check responses to determine whether
// the response should return without processing the body. Returning true
// will short-curcuit the BodyProvider.
type ResponseChecker func(*http.Response, error) (bool, error)

// Service is an HTTP Request builder and sender.
type Service struct {
	// httpClient for doing requests
	httpClient Doer
	// HTTP method (GET, POST, etc.)
	method string
	// raw url string for requests
	rawURL string
	// stores key-values pairs to add to request's Headers
	header http.Header
	// url tagged query structs
	queryStructs []interface{}
	// body provider
	bodyProvider BodyProvider
	// responder
	responder Responder
}

// New returns a new Service with an http DefaultClient.
func New() *Service {
	return &Service{
		httpClient:   http.DefaultClient,
		method:       "GET",
		header:       make(http.Header),
		queryStructs: make([]interface{}, 0),
		responder:    GenericResponder(),
	}
}

// New returns a copy of a Service for creating a new Service with properties
// from a parent Service. For example,
//
// 	parentSling := Service.New().Service(client).Base("https://api.io/")
// 	fooSling := parentSling.New().Get("foo/")
// 	barSling := parentSling.New().Get("bar/")
//
// fooSling and barSling will both use the same client, but send requests to
// https://api.io/foo/ and https://api.io/bar/ respectively.
//
// Note that query and body values are copied so if pointer values are used,
// mutating the original value will mutate the value within the child Service.
func (s *Service) New() *Service {
	// copy Headers pairs into new Header map
	headerCopy := make(http.Header)
	for k, v := range s.header {
		headerCopy[k] = v
	}
	return &Service{
		httpClient: s.httpClient,
		method:       s.method,
		rawURL:       s.rawURL,
		header:       headerCopy,
		queryStructs: append([]interface{}{}, s.queryStructs...),
		bodyProvider: s.bodyProvider,
		responder:    s.responder,
	}
}

// Http Client

// Service sets the http Service used to do requests. If a nil client is given,
// the http.DefaultClient will be used.
func (s *Service) Client(httpClient *http.Client) *Service {
	if httpClient == nil {
		return s.Doer(http.DefaultClient)
	}
	return s.Doer(httpClient)
}

// Doer sets the custom Doer implementation used to do requests.
// If a nil client is given, the http.DefaultClient will be used.
func (s *Service) Doer(doer Doer) *Service {
	if doer == nil {
		s.httpClient = http.DefaultClient
	} else {
		s.httpClient = doer
	}
	return s
}

// Method

// Method sets the Service method and the path to the given pathURL
func (s *Service) Method(method string, pathURL ...string) *Service {
	if method != "" {
		s.method = method
	}
	return s.Path(strings.Join(pathURL, "/"))
}

// Methodf sets the Service method and the path to the resolved path format.
func (s *Service) Methodf(method string, format string, a ...interface{}) *Service {
	return s.Method(method, fmt.Sprintf(format, a...))
}

// Head sets the Service method to HEAD and the path to the given pathURL.
func (s *Service) Head(pathURL ...string) *Service {
	return s.Method("HEAD", pathURL...)
}

// Headf sets the Service method to HEAD and the path to the resolved path format.
func (s *Service) Headf(format string, a ...interface{}) *Service {
	return s.Head(fmt.Sprintf(format, a...))
}

// Get sets the Service method to GET and the path to the given pathURL.
func (s *Service) Get(pathURL ...string) *Service {
	return s.Method("GET", pathURL...)
}

// Getf sets the Service method to GET and the path to the resolved path format.
func (s *Service) Getf(format string, a ...interface{}) *Service {
	return s.Get(fmt.Sprintf(format, a...))
}

// Post sets the Service method to POST and the path to the given pathURL.
func (s *Service) Post(pathURL ...string) *Service {
	return s.Method("POST", pathURL...)
}

// Postf sets the Service method to POST and the path to the resolved path format.
func (s *Service) Postf(format string, a ...interface{}) *Service {
	return s.Post(fmt.Sprintf(format, a...))
}

// Put sets the Service method to PUT and the path to the given pathURL.
func (s *Service) Put(pathURL ...string) *Service {
	return s.Method("PUT", pathURL...)
}

// Putf sets the Service method to PUT and the path to the resolved path format.
func (s *Service) Putf(format string, a ...interface{}) *Service {
	return s.Put(fmt.Sprintf(format, a...))
}

// Patch sets the Service method to PATCH and the path to the given pathURL.
func (s *Service) Patch(pathURL ...string) *Service {
	return s.Method("PATCH", pathURL...)
}

// Patchf sets the Service method to PATCH and the path to the resolved path format.
func (s *Service) Patchf(format string, a ...interface{}) *Service {
	return s.Patch(fmt.Sprintf(format, a...))
}

// Delete sets the Service method to DELETE and the path to the given pathURL.
func (s *Service) Delete(pathURL ...string) *Service {
	return s.Method("DELETE", pathURL...)
}

// Deletef sets the Service method to DELETE and the path to the resolved path format.
func (s *Service) Deletef(format string, a ...interface{}) *Service {
	return s.Delete(fmt.Sprintf(format, a...))
}

// Header

// Add adds the key, value pair in Headers, appending values for existing keys
// to the key's values. Header keys are canonicalized.
func (s *Service) Add(key, value string) *Service {
	s.header.Add(key, value)
	return s
}

// Set sets the key, value pair in Headers, replacing existing values
// associated with key. Header keys are canonicalized.
func (s *Service) Set(key, value string) *Service {
	s.header.Set(key, value)
	return s
}

// SetBasicAuth sets the Authorization header to use HTTP Basic Authentication
// with the provided username and password. With HTTP Basic Authentication
// the provided username and password are not encrypted.
func (s *Service) SetBasicAuth(username, password string) *Service {
	return s.Set("Authorization", "Basic "+basicAuth(username, password))
}

// basicAuth returns the base64 encoded username:password for basic auth copied
// from net/http.
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// Url

// RawBase sets the rawURL. If you intend to extend the url with Path,
// baseUrl should be specified with a trailing slash, or just use Base.
func (s *Service) RawBase(rawURL string) *Service {
	s.rawURL = rawURL
	return s
}

// Base sets the rawURL.
func (s *Service) Base(rawURL string) *Service {
	s.rawURL = s.slashIt(rawURL)
	return s
}

// slashIt adds a trailing slash to the string ensuring there are no double slashes.
func (s *Service) slashIt(str string) string {
	return strings.TrimSuffix(str, "/") + "/"
}

// Path extends the rawURL with the given path by resolving the reference to
// an absolute URL. If parsing errors occur, the rawURL is left unmodified.
func (s *Service) Path(path string) *Service {
	baseURL, baseErr := url.Parse(s.rawURL)
	pathURL, pathErr := url.Parse(path)
	if baseErr == nil && pathErr == nil {
		s.rawURL = baseURL.ResolveReference(pathURL).String()
		return s
	}
	return s
}

// QueryStruct appends the queryStruct to the Service's queryStructs. The value
// pointed to by each queryStruct will be encoded as url query parameters on
// new requests (see Request()).
// The queryStruct argument should be a pointer to a url tagged struct. See
// https://godoc.org/github.com/google/go-querystring/query for details.
func (s *Service) QueryStruct(queryStruct interface{}) *Service {
	if queryStruct != nil {
		s.queryStructs = append(s.queryStructs, queryStruct)
	}
	return s
}

// Body

// Body sets the Service's body. The body value will be set as the Body on new
// requests (see Request()).
// If the provided body is also an io.Closer, the request Body will be closed
// by http.Service methods.
func (s *Service) Body(body io.Reader) *Service {
	if body == nil {
		return s
	}
	return s.BodyProvider(bodyProvider{body: body})
}

// BodyProvider sets the Service's body provider.
func (s *Service) BodyProvider(body BodyProvider) *Service {
	if body == nil {
		return s
	}
	s.bodyProvider = body

	ct := body.ContentType()
	if ct != "" {
		s.Set(contentType, ct)
	}

	return s
}

// BodyJSON sets the Service's bodyJSON. The value pointed to by the bodyJSON
// will be JSON encoded as the Body on new requests (see Request()).
// The bodyJSON argument should be a pointer to a JSON tagged struct. See
// https://golang.org/pkg/encoding/json/#MarshalIndent for details.
func (s *Service) BodyJSON(bodyJSON interface{}) *Service {
	if bodyJSON == nil {
		return s
	}
	return s.BodyProvider(jsonBodyProvider{payload: bodyJSON})
}

// BodyForm sets the Service's bodyForm. The value pointed to by the bodyForm
// will be url encoded as the Body on new requests (see Request()).
// The bodyForm argument should be a pointer to a url tagged struct. See
// https://godoc.org/github.com/google/go-querystring/query for details.
func (s *Service) BodyForm(bodyForm interface{}) *Service {
	if bodyForm == nil {
		return s
	}
	return s.BodyProvider(formBodyProvider{payload: bodyForm})
}

// Responder sets the Service's responder.
func (s *Service) Responder(responder Responder) *Service {
	if responder == nil {
		return s
	}
	s.responder = responder
	return s
}

// JSONResponder sets the Service's responder to do JSON.
func (s *Service) JSONResponder(success, failure interface{}) *Service {
	s.responder = JSONResponder(success, failure)
	return s
}

// JSONResponder sets the Service's responder to do JSON for success only.
func (s *Service) SuccessJSONResponder(success interface{}) *Service {
	s.responder = JSONResponder(success, nil)
	return s
}

// Requests

// Request returns a new http.Request created with the Service properties.
// Returns any errors parsing the rawURL, encoding query structs, encoding
// the body, or creating the http.Request.
func (s *Service) Request() (*http.Request, error) {
	reqURL, err := url.Parse(s.rawURL)
	if err != nil {
		return nil, err
	}

	err = s.addQueryStructs(reqURL, s.queryStructs)
	if err != nil {
		return nil, err
	}

	var body io.Reader
	if s.bodyProvider != nil {
		body, err = s.bodyProvider.Body()
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(s.method, reqURL.String(), body)
	if err != nil {
		return nil, err
	}
	addHeaders(req, s.header)
	return req, err
}

// AsyncRequest returns a new AsyncRequest created with the Service properties.
// Returns any errors parsing the rawURL, encoding query structs, encoding
// the body, or creating the AsyncRequest.
func (s *Service) AsyncRequest(responder Responder) *AsyncRequest {
	if responder == nil {
		responder = s.GetResponder()
	}
	return NewAsyncRequestWithResponder(s, responder)
}

// addQueryStructs parses url tagged query structs using go-querystring to
// encode them to url.Values and format them onto the url.RawQuery. Any
// query parsing or encoding errors are returned.
// TODO Optionally decode the query params.
func (s *Service) addQueryStructs(reqURL *url.URL, queryStructs []interface{}) error {
	urlValues, err := url.ParseQuery(reqURL.RawQuery)
	if err != nil {
		return err
	}
	// encodes query structs into a url.Values map and merges maps
	for _, queryStruct := range queryStructs {
		queryValues, err := goquery.Values(queryStruct)
		if err != nil {
			return err
		}
		for key, values := range queryValues {
			for _, value := range values {
				urlValues.Add(key, value)
			}
		}
	}
	// url.Values format to a sorted "url encoded" string, e.g. "key=val&foo=bar"
	reqURL.RawQuery = urlValues.Encode()
	reqURL.RawQuery, _ = url.QueryUnescape(reqURL.RawQuery)

	return nil
}

// addHeaders adds the key, value pairs from the given http.Header to the
// request. Values for existing keys are appended to the keys values.
func addHeaders(req *http.Request, header http.Header) {
	for key, values := range header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
}

// Sending

// ReceiveSuccess creates a new HTTP request and returns the response. Success
// responses (2XX) are JSON decoded into the value pointed to by successV.
// Any error creating the request, sending it, or decoding a 2XX response
// is returned.
func (s *Service) ReceiveSuccess(successV interface{}) (*http.Response, error) {
	return s.Receive(successV, nil)
}

// Receive creates a new HTTP request and returns the response. Success
// responses (2XX) are JSON decoded into the value pointed to by successV and
// other responses are JSON decoded into the value pointed to by failureV.
// Any error creating the request, sending it, or decoding the response is
// returned.
// Receive is shorthand for calling Request and Do.
func (s *Service) Receive(successV, failureV interface{}) (*http.Response, error) {
	req, err := s.Request()
	if err != nil {
		return nil, err
	}

	resp, err := s.Do(req)
	successV = s.responder.GetSuccess()
	failureV = s.responder.GetFailure()
	return resp, err
}

// Do sends an HTTP request and returns the response. Success responses (2XX)
// are JSON decoded into the value pointed to by successV and other responses
// are JSON decoded into the value pointed to by failureV.
// Any error sending the request or decoding the response is returned.
func (s *Service) Do(req *http.Request) (*http.Response, error) {
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return resp, err
	}

	defer func() {
		if resp.Header.Get("Accept-Ranges") != "bytes" {
			io.CopyN(ioutil.Discard, resp.Body, 512)
		}
		resp.Body.Close()
	}()


	// Do correct Response
	r := s.responder.Respond(req, resp, err)
	return r.DoResponse()
}

// doAsync helps DoAsync by performing the actual request returning
// the response on the proper channel.
func (s *Service) doAsync(req AsyncRequest, ch chan<- *AsyncResponse) {
	//resp, err := s.Do(req.Request, req.Success, req.Failure)
	//ch <- &AsyncResponse{resp, req.Success, req.Failure, err}
	resp, err := s.Responder(req.Responder).Do(req.Request)
	ch <- &AsyncResponse{
		responder: s.responder,
		Response:  resp,
		Success:   s.responder.GetSuccess(),
		Failure:   s.responder.GetFailure(),
		Error:     err,
	}
}

// DoAsync performs the requests in an asychronous pattern.
func (s *Service) DoAsync(reqs []AsyncRequest) []*AsyncResponse {
	l := len(reqs)
	ch := make(chan *AsyncResponse, l)
	responses := []*AsyncResponse{}
	for _, req := range reqs {
		go s.doAsync(req, ch)
	}

	for {
		select {
		case r := <-ch:
			responses = append(responses, r)
			if len(responses) == l {
				return responses
			}
		}
	}

	return responses
}

// DecodeResponse decodes the JSON response.
func (s *Service) DecodeResponse(resp *http.Response, v interface{}) (err error) {
	return decodeResponseBodyJSON(resp, v)
}

// isOk determines whether the HTTP Status Code is an OK Code (200-299)
func isOk(statusCode int) bool {
	return (http.StatusOK <= statusCode && statusCode <= 299)
}
