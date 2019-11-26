package meteor

import (
	"net/http"
)

const (
	libraryVersion string = "1"

	// UserAgent is the user agent making the requests
	UserAgent string = "go-wx/" + libraryVersion
)

// Credentials holds the keys used for the API
type Credentials map[string]string

// NewCredentials returns new credentials map.
func NewCredentials(creds map[string]string) Credentials {
	return creds
}

// Meteor manages communication with any API.
type Meteor struct {
	// HTTP client used to communicate with the API.
	httpClient *http.Client

	// Credentials holder
	credentials Credentials

	// HTTP Requests holder
	requests []*http.Request

	// Reuse a single struct instead of allocating one for each service on the heap.
	Common *Service

	// User agent used when communicating with the API.
	UserAgent string
}

// GetCredBy gets a credential by key.
func (c *Meteor) GetCredBy(key string) string {
	if v, ok := c.credentials[key]; ok {
		return v
	}
	return ""
}

// GetHTTPClient gets the HTTP Client.
func (c *Meteor) GetHTTPClient() *http.Client {
	return c.httpClient
}

// NewClient returns a new API client. If a nil httpClient is
// provided, http.DefaultClient will be used. To use API methods which require
// authentication, provide an http.Client that will perform the authentication
// for you (such as that provided by the golang.org/x/oauth2 library).
func NewMeteor(credentials Credentials, httpClient ...*http.Client) *Meteor {
	var theClient *http.Client
	if len(httpClient) == 0 || (len(httpClient) == 1 && httpClient[0] == nil) {
		theClient = GetDefaultClient()
	} else {
		theClient = httpClient[0]
	}
	c := &Meteor{
		httpClient:  theClient,
		credentials: credentials,
		UserAgent:   UserAgent,
		Common:      New().Client(theClient),
	}

	return c
}

// NewClient returns a new API client. If a nil httpClient is
// provided, http.DefaultClient will be used. To use API methods which require
// authentication, provide an http.Client that will perform the authentication
// for you (such as that provided by the golang.org/x/oauth2 library).
func NewSimpleMeteor(httpClient ...*http.Client) *Meteor {
	var theClient *http.Client
	if len(httpClient) == 0 || (len(httpClient) == 1 && httpClient[0] == nil) {
		theClient = GetDefaultClient()
	} else {
		theClient = httpClient[0]
	}
	c := &Meteor{
		httpClient: theClient,
		UserAgent:  UserAgent,
		Common:     New().Client(theClient),
	}

	return c
}
