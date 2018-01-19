package meteor

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

type FakeParams struct {
	KindName string `url:"kind_name"`
	Count    int    `url:"count"`
}

// Url-tagged query struct
var paramsA = struct {
	Limit int `url:"limit"`
}{
	30,
}
var paramsB = FakeParams{KindName: "recent", Count: 25}

// Json-tagged model struct
type FakeModel struct {
	Text          string  `json:"text,omitempty"`
	FavoriteCount int64   `json:"favorite_count,omitempty"`
	Temperature   float64 `json:"temperature,omitempty"`
}

// Generic Body
var genericBody = strings.NewReader("raw body")

// JSON Body
type IssueRequest struct {
	Title     string   `json:"title,omitempty"`
	Body      string   `json:"body,omitempty"`
	Assignee  string   `json:"assignee,omitempty"`
	Milestone int      `json:"milestone,omitempty"`
	Labels    []string `json:"labels,omitempty"`
}

var jsonBody = &IssueRequest{
	Title: "Test title",
	Body:  "Some issue",
}

// Form Body
type StatusUpdateParams struct {
	Status            string  `url:"status,omitempty"`
	InReplyToStatusId int64   `url:"in_reply_to_status_id,omitempty"`
	MediaIds          []int64 `url:"media_ids,omitempty,comma"`
}

var formBody = &StatusUpdateParams{Status: "writing some Go"}

var formatVars = []interface{}{"33", "-84"}

const (
	baseURL       = "https://example.com"
	geocodeFormat = "geocode/%v/%v"
	geocodePath   = "/geocode/33/-84"
)

func TestNew(t *testing.T) {
	got := New()

	if got.httpClient != GetDefaultClient() {
		t.Errorf("expected %v, got %v", GetDefaultClient(), got.httpClient)
	}
	if got.header == nil {
		t.Errorf("Header map not initialized with make")
	}
	if got.queryStructs == nil {
		t.Errorf("queryStructs not initialized with make")
	}
}

// TODO Comment tests
func TestService_New(t *testing.T) {
	fakeBodyProvider := jsonBodyProvider{FakeModel{}}

	tests := []struct {
		name   string
		parent *Service
	}{
		{"new", New()},
		{"GETMethod", &Service{httpClient: &http.Client{}, method: "GET", rawURL: baseURL}},
		{"nilHttpClient", &Service{httpClient: nil, method: "", rawURL: baseURL}},
		{"emptyQueryStructMap", &Service{queryStructs: make([]interface{}, 0)}},
		{"paramsA", &Service{queryStructs: []interface{}{paramsA}}},
		{"paramsA+paramsB", &Service{queryStructs: []interface{}{paramsA, paramsB}}},
		{"fakeBodyProvider", &Service{bodyProvider: fakeBodyProvider}},
		{"nilBodyProvider", &Service{bodyProvider: nil}},
		{"newHeaders+ContentType", New().Add("Content-Type", "application/json")},
		{"newHeaders+ContentType+ContentType", New().Add("Content-Type", "application/json").Add("Content-Type", "application/json")},
		{"newHeadersA+B", New().Add("A", "B").Add("a", "c").New()},
		{"newNewHeadera", New().Add("A", "B").New().Add("a", "c")},
		{"newBodyFormParamsB", New().BodyForm(paramsB)},
		{"newBodyFormParamsBNew", New().BodyForm(paramsB).New()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.parent.New()
			if got.httpClient != tt.parent.httpClient {
				t.Errorf("expected %v, got %v", tt.parent.httpClient, got.httpClient)
			}
			if got.method != tt.parent.method {
				t.Errorf("expected %s, got %s", tt.parent.method, got.method)
			}
			if got.rawURL != tt.parent.rawURL {
				t.Errorf("expected %s, got %s", tt.parent.rawURL, got.rawURL)
			}
			// Header should be a copy of parent header. For example, calling
			// base.Add("k","v") should not mutate previously created chilren
			if tt.parent.header != nil {
				// struct literal cases don't init Header in usual way, skip header check
				if !assert.Equal(t, got.header, tt.parent.header) {
					t.Errorf("not DeepEqual: expected %v, got %v", tt.parent.header, got.header)
				}
				tt.parent.header.Add("K", "V")
				if got.header.Get("K") != "" {
					t.Errorf("child.header was a reference to original map, should be copy")
				}
			}
			// queryStruct slice should be a new slice with a copy of the contents
			if len(tt.parent.queryStructs) > 0 {
				// mutating one slice should not mutate the other
				got.queryStructs[0] = nil
				if tt.parent.queryStructs[0] == nil {
					t.Errorf("child.queryStructs was a re-slice, expected slice with copied contents")
				}
			}
			// body should be copied
			if got.bodyProvider != tt.parent.bodyProvider {
				t.Errorf("expected %v, got %v", tt.parent.bodyProvider, got.bodyProvider)
			}
		})
	}
}

func TestService_Reset(t *testing.T) {
	svc := New()
	tests := []struct {
		name string
		s    *Service
	}{
		{"new", svc},
		{"newHeaders+ContentType", svc.Add("Content-Type", "application/json").New()},
		{"newHeaders+ContentType+ContentType", svc.Add("Content-Type", "application/json").Add("Content-Type", "application/json").New()},
		{"newHeadersA+B", svc.Add("A", "B").Add("a", "c").New()},
		{"newNewHeadera", svc.Add("A", "B").New().Add("a", "c").New()},
		{"newBodyFormParamsB", svc.BodyForm(paramsB).New()},
		{"newBodyFormParamsBNew", svc.BodyForm(paramsB).New()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Reset()
			if !assert.Equal(t, GetDefaultClient(), got.httpClient) {
				t.Errorf("%v Service.Reset() = %v, want %v", tt.name, got.httpClient, GetDefaultClient())
			}
			if !assert.Equal(t, "GET", got.method) {
				t.Errorf("%v Service.Reset() = %v, want %v", tt.name, got.method, "GET")
			}
			if !assert.Equal(t, "", got.rawURL) {
				t.Errorf("%v Service.Reset() = %v, want %v", tt.name, got.rawURL, "")
			}
			if !assert.Equal(t, nil, got.bodyProvider) {
				t.Errorf("%v Service.Reset() = %v, want %v", tt.name, got.bodyProvider, nil)
			}
			if !assert.Equal(t, make(http.Header), got.header) {
				t.Errorf("%v Service.Reset() = %v, want %v", tt.name, got.header, make(http.Header))
			}
			if !assert.Equal(t, make([]interface{}, 0), got.queryStructs) {
				t.Errorf("%v Service.Reset() = %v, want %v", tt.name, got.queryStructs, make([]interface{}, 0))
			}
			if !assert.Equal(t, GenericResponder(), got.responder) {
				t.Errorf("%v Service.Reset() = %v, want %v", tt.name, got.responder, GenericResponder())
			}
		})
	}
}

func TestService_Client(t *testing.T) {
	type args struct {
		httpClient *http.Client
	}
	tests := []struct {
		name string
		s    *Service
		args args
		want *http.Client
	}{
		// nil should set our default client
		{"nil", New(), args{nil}, GetDefaultClient()},
		// empty should set the GoLang default client
		{"empty", New(), args{&http.Client{}}, GetDefaultClient()},
		// custom clients should be set
		{"default", New(), args{GetDefaultClient()}, GetDefaultClient()},
		{"custom", New(), args{customClient}, customClient},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Client(tt.args.httpClient); !assert.Equal(t, tt.want, got.httpClient) {
				t.Errorf("%v Service.Client() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestService_Doer(t *testing.T) {
	type args struct {
		doer Doer
	}
	tests := []struct {
		name string
		s    *Service
		args args
		want Doer
	}{
		// nil should set our default client
		{"nil", New(), args{nil}, GetDefaultClient()},
		// empty should set the GoLang default client
		{"empty", New(), args{&http.Client{}}, GetDefaultClient()},
		// custom clients should be set
		{"default", New(), args{GetDefaultClient()}, GetDefaultClient()},
		{"custom", New(), args{customClient}, customClient},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Doer(tt.args.doer); !assert.Equal(t, tt.want, got.httpClient) {
				t.Errorf("%v Service.Doer() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestService_Method(t *testing.T) {
	type args struct {
		method  string
		pathURL []string
	}
	tests := []struct {
		name       string
		s          *Service
		args       args
		wantMethod string
		wantPath   string
	}{
		{"GET-noPath", New(), args{"GET", nil}, "GET", ""},
		{"GET-withPath", New(), args{"GET", []string{"https://example.com/test"}}, "GET", "https://example.com/test"},
		{"HEAD-noPath", New(), args{"HEAD", nil}, "HEAD", ""},
		{"HEAD-withPath", New(), args{"HEAD", []string{"https://example.com/test"}}, "HEAD", "https://example.com/test"},
		{"POST-noPath", New(), args{"POST", nil}, "POST", ""},
		{"POST-withPath", New(), args{"POST", []string{"https://example.com/test"}}, "POST", "https://example.com/test"},
		{"PUT-noPath", New(), args{"PUT", nil}, "PUT", ""},
		{"PUT-withPath", New(), args{"PUT", []string{"https://example.com/test"}}, "PUT", "https://example.com/test"},
		{"PATCH-noPath", New(), args{"PATCH", nil}, "PATCH", ""},
		{"PATCH-withPath", New(), args{"PATCH", []string{"https://example.com/test"}}, "PATCH", "https://example.com/test"},
		{"DELETE-noPath", New(), args{"DELETE", nil}, "DELETE", ""},
		{"DELETE-withPath", New(), args{"DELETE", []string{"https://example.com/test"}}, "DELETE", "https://example.com/test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Method(tt.args.method, tt.args.pathURL...)
			if got.method != tt.wantMethod {
				t.Errorf("Service.Method(%v) = %v, want %v", tt.name, got.method, tt.wantMethod)
			}
			if got.rawURL != tt.wantPath {
				t.Errorf("Service.Method(%v) = %v, want %v", tt.name, got.rawURL, tt.wantPath)
			}
		})
	}
}

func TestService_Methodf(t *testing.T) {
	type args struct {
		method string
		format string
		a      []interface{}
	}
	tests := []struct {
		name       string
		s          *Service
		args       args
		wantMethod string
		wantPath   string
	}{
		{"GET", New(), args{"GET", geocodeFormat, formatVars}, "GET", geocodePath},
		{"HEAD", New(), args{"HEAD", geocodeFormat, formatVars}, "HEAD", geocodePath},
		{"POST", New(), args{"POST", geocodeFormat, formatVars}, "POST", geocodePath},
		{"PUT", New(), args{"PUT", geocodeFormat, formatVars}, "PUT", geocodePath},
		{"PATCH", New(), args{"PATCH", geocodeFormat, formatVars}, "PATCH", geocodePath},
		{"DELETE", New(), args{"DELETE", geocodeFormat, formatVars}, "DELETE", geocodePath},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Methodf(tt.args.method, tt.args.format, tt.args.a...)
			if got.method != tt.wantMethod {
				t.Errorf("Service.Methodf(%v) = %v, want %v", tt.name, got.method, tt.wantMethod)
			}
			if got.rawURL != tt.wantPath {
				t.Errorf("Service.Methodf(%v) = %v, want %v", tt.name, got.rawURL, tt.wantPath)
			}
		})
	}
}

func TestService_Head(t *testing.T) {
	type args struct {
		pathURL []string
	}
	tests := []struct {
		name       string
		s          *Service
		args       args
		wantMethod string
		wantPath   string
	}{
		{"nil", New(), args{}, "HEAD", ""},
		{"fullPath", New(), args{[]string{"https://example.com/foo/"}}, "HEAD", "https://example.com/foo/"},
		{"multiPaths", New(), args{[]string{"https://example.com/", "foo/"}}, "HEAD", "https://example.com/foo/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Head(tt.args.pathURL...)
			if got.method != tt.wantMethod {
				t.Errorf("Service.Head(%v) = %v, want %v", tt.name, got.method, tt.wantMethod)
			}
			if got.rawURL != tt.wantPath {
				t.Errorf("Service.Head(%v) = %v, want %v", tt.name, got.rawURL, tt.wantPath)
			}
		})
	}
}

func TestService_Headf(t *testing.T) {
	type args struct {
		format string
		a      []interface{}
	}
	tests := []struct {
		name       string
		s          *Service
		args       args
		wantMethod string
		wantPath   string
	}{
		{"head", New(), args{geocodeFormat, formatVars}, "HEAD", geocodePath},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Headf(tt.args.format, tt.args.a...)
			if got.method != tt.wantMethod {
				t.Errorf("Service.Headf(%v) = %v, want %v", tt.name, got.method, tt.wantMethod)
			}
			if got.rawURL != tt.wantPath {
				t.Errorf("Service.Headf(%v) = %v, want %v", tt.name, got.rawURL, tt.wantPath)
			}
		})
	}
}

func TestService_Get(t *testing.T) {
	type args struct {
		pathURL []string
	}
	tests := []struct {
		name       string
		s          *Service
		args       args
		wantMethod string
		wantPath   string
	}{
		{"nil", New(), args{}, "GET", ""},
		{"fullPath", New(), args{[]string{"https://example.com/foo/"}}, "GET", "https://example.com/foo/"},
		{"multiPaths", New(), args{[]string{"https://example.com/", "foo/"}}, "GET", "https://example.com/foo/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Get(tt.args.pathURL...)
			if got.method != tt.wantMethod {
				t.Errorf("Service.Get(%v) = %v, want %v", tt.name, got.method, tt.wantMethod)
			}
			if got.rawURL != tt.wantPath {
				t.Errorf("Service.Get(%v) = %v, want %v", tt.name, got.rawURL, tt.wantPath)
			}
		})
	}
}

func TestService_Getf(t *testing.T) {
	type args struct {
		format string
		a      []interface{}
	}
	tests := []struct {
		name       string
		s          *Service
		args       args
		wantMethod string
		wantPath   string
	}{
		{"get", New(), args{geocodeFormat, formatVars}, "GET", geocodePath},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Getf(tt.args.format, tt.args.a...)
			if got.method != tt.wantMethod {
				t.Errorf("Service.Getf(%v) = %v, want %v", tt.name, got.method, tt.wantMethod)
			}
			if got.rawURL != tt.wantPath {
				t.Errorf("Service.Getf(%v) = %v, want %v", tt.name, got.rawURL, tt.wantPath)
			}
		})
	}
}

func TestService_Post(t *testing.T) {
	type args struct {
		pathURL []string
	}
	tests := []struct {
		name       string
		s          *Service
		args       args
		wantMethod string
		wantPath   string
	}{
		{"nil", New(), args{}, "POST", ""},
		{"fullPath", New(), args{[]string{"https://example.com/foo/"}}, "POST", "https://example.com/foo/"},
		{"multiPaths", New(), args{[]string{"https://example.com/", "foo/"}}, "POST", "https://example.com/foo/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Post(tt.args.pathURL...)
			if got.method != tt.wantMethod {
				t.Errorf("Service.Post(%v) = %v, want %v", tt.name, got.method, tt.wantMethod)
			}
			if got.rawURL != tt.wantPath {
				t.Errorf("Service.Post(%v) = %v, want %v", tt.name, got.rawURL, tt.wantPath)
			}
		})
	}
}

func TestService_Postf(t *testing.T) {
	type args struct {
		format string
		a      []interface{}
	}
	tests := []struct {
		name       string
		s          *Service
		args       args
		wantMethod string
		wantPath   string
	}{
		{"post", New(), args{geocodeFormat, formatVars}, "POST", geocodePath},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Postf(tt.args.format, tt.args.a...)
			if got.method != tt.wantMethod {
				t.Errorf("Service.Postf(%v) = %v, want %v", tt.name, got.method, tt.wantMethod)
			}
			if got.rawURL != tt.wantPath {
				t.Errorf("Service.Postf(%v) = %v, want %v", tt.name, got.rawURL, tt.wantPath)
			}
		})
	}
}

func TestService_Put(t *testing.T) {
	type args struct {
		pathURL []string
	}
	tests := []struct {
		name       string
		s          *Service
		args       args
		wantMethod string
		wantPath   string
	}{
		{"nil", New(), args{}, "PUT", ""},
		{"fullPath", New(), args{[]string{"https://example.com/foo/"}}, "PUT", "https://example.com/foo/"},
		{"multiPaths", New(), args{[]string{"https://example.com/", "foo/"}}, "PUT", "https://example.com/foo/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Put(tt.args.pathURL...)
			if got.method != tt.wantMethod {
				t.Errorf("Service.Put(%v) = %v, want %v", tt.name, got.method, tt.wantMethod)
			}
			if got.rawURL != tt.wantPath {
				t.Errorf("Service.Put(%v) = %v, want %v", tt.name, got.rawURL, tt.wantPath)
			}
		})
	}
}

func TestService_Putf(t *testing.T) {
	type args struct {
		format string
		a      []interface{}
	}
	tests := []struct {
		name       string
		s          *Service
		args       args
		wantMethod string
		wantPath   string
	}{
		{"put", New(), args{geocodeFormat, formatVars}, "PUT", geocodePath},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Putf(tt.args.format, tt.args.a...)
			if got.method != tt.wantMethod {
				t.Errorf("Service.Putf(%v) = %v, want %v", tt.name, got.method, tt.wantMethod)
			}
			if got.rawURL != tt.wantPath {
				t.Errorf("Service.Putf(%v) = %v, want %v", tt.name, got.rawURL, tt.wantPath)
			}
		})
	}
}

func TestService_Patch(t *testing.T) {
	type args struct {
		pathURL []string
	}
	tests := []struct {
		name       string
		s          *Service
		args       args
		wantMethod string
		wantPath   string
	}{
		{"nil", New(), args{}, "PATCH", ""},
		{"fullPath", New(), args{[]string{"https://example.com/foo/"}}, "PATCH", "https://example.com/foo/"},
		{"multiPaths", New(), args{[]string{"https://example.com/", "foo/"}}, "PATCH", "https://example.com/foo/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Patch(tt.args.pathURL...)
			if got.method != tt.wantMethod {
				t.Errorf("Service.Patch(%v) = %v, want %v", tt.name, got.method, tt.wantMethod)
			}
			if got.rawURL != tt.wantPath {
				t.Errorf("Service.Patch(%v) = %v, want %v", tt.name, got.rawURL, tt.wantPath)
			}
		})
	}
}

func TestService_Patchf(t *testing.T) {
	type args struct {
		format string
		a      []interface{}
	}
	tests := []struct {
		name       string
		s          *Service
		args       args
		wantMethod string
		wantPath   string
	}{
		{"patch", New(), args{geocodeFormat, formatVars}, "PATCH", geocodePath},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Patchf(tt.args.format, tt.args.a...)
			if got.method != tt.wantMethod {
				t.Errorf("Service.Patchf(%v) = %v, want %v", tt.name, got.method, tt.wantMethod)
			}
			if got.rawURL != tt.wantPath {
				t.Errorf("Service.Patchf(%v) = %v, want %v", tt.name, got.rawURL, tt.wantPath)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	type args struct {
		pathURL []string
	}
	tests := []struct {
		name       string
		s          *Service
		args       args
		wantMethod string
		wantPath   string
	}{
		{"nil", New(), args{}, "DELETE", ""},
		{"fullPath", New(), args{[]string{"https://example.com/foo/"}}, "DELETE", "https://example.com/foo/"},
		{"multiPaths", New(), args{[]string{"https://example.com/", "foo/"}}, "DELETE", "https://example.com/foo/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Delete(tt.args.pathURL...)
			if got.method != tt.wantMethod {
				t.Errorf("Service.Delete(%v) = %v, want %v", tt.name, got.method, tt.wantMethod)
			}
			if got.rawURL != tt.wantPath {
				t.Errorf("Service.Delete(%v) = %v, want %v", tt.name, got.rawURL, tt.wantPath)
			}
		})
	}
}

func TestService_Deletef(t *testing.T) {
	type args struct {
		format string
		a      []interface{}
	}
	tests := []struct {
		name       string
		s          *Service
		args       args
		wantMethod string
		wantPath   string
	}{
		{"delete", New(), args{geocodeFormat, formatVars}, "DELETE", geocodePath},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Deletef(tt.args.format, tt.args.a...)
			if got.method != tt.wantMethod {
				t.Errorf("Service.Deletef(%v) = %v, want %v", tt.name, got.method, tt.wantMethod)
			}
			if got.rawURL != tt.wantPath {
				t.Errorf("Service.Deletef(%v) = %v, want %v", tt.name, got.rawURL, tt.wantPath)
			}
		})
	}
}

func TestService_Add(t *testing.T) {
	svc := New()
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		s    *Service
		args args
		want http.Header
	}{
		{"nil", New(), args{}, make(http.Header)},
		// Add a new header
		{"single", svc, args{"Content-Type", "text/plain"}, http.Header{"Content-Type": []string{"text/plain"}}},
		// Add to an already existing header
		{"double", svc, args{"Content-Type", "text/plain"}, http.Header{"Content-Type": []string{"text/plain", "text/plain"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Add(tt.args.key, tt.args.value)
			if !assert.Equal(t, tt.want, got.header) {
				t.Errorf("%v Service.Add() = %#v, want %#v", tt.name, got, tt.want)
			}
		})
	}
}

func TestService_Set(t *testing.T) {
	svc := New()
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		s    *Service
		args args
		want http.Header
	}{
		{"nil", New(), args{}, make(http.Header)},
		// Set new Header
		{"single", svc, args{"Content-Type", "text/plain"}, http.Header{"Content-Type": []string{"text/plain"}}},
		// Set an already set header
		{"double", svc, args{"Content-Type", "application/json"}, http.Header{"Content-Type": []string{"application/json"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Set(tt.args.key, tt.args.value)
			if !assert.Equal(t, tt.want, got.header) {
				t.Errorf("%v Service.Set() = %#v, want %#v", tt.name, got, tt.want)
			}
		})
	}
}

func TestService_SetBasicAuth(t *testing.T) {
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name string
		s    *Service
		args args
		want string
	}{
		{"nil", New(), args{}, "Basic "},
		{"usernameOnly", New(), args{"my_user", ""}, "Basic bXlfdXNlcjo="},
		{"passwordOnly", New(), args{"", "password"}, "Basic OnBhc3N3b3Jk"},
		{"basic", New(), args{"my_user", "password"}, "Basic bXlfdXNlcjpwYXNzd29yZA=="},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.SetBasicAuth(tt.args.username, tt.args.password); !assert.Equal(t, tt.want, got.header.Get("Authorization")) {
				t.Errorf("%v Service.SetBasicAuth() = %v, want %v", tt.name, got.header.Get("Authorization"), tt.want)
			}
		})
	}
}

func Test_basicAuth(t *testing.T) {
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"nil", args{}, ""},
		{"usernameOnly", args{"my_user", ""}, "bXlfdXNlcjo="},
		{"passwordOnly", args{"", "password"}, "OnBhc3N3b3Jk"},
		{"basic", args{"my_user", "password"}, "bXlfdXNlcjpwYXNzd29yZA=="},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := basicAuth(tt.args.username, tt.args.password); got != tt.want {
				t.Errorf("%v basicAuth() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestService_RawBase(t *testing.T) {
	type args struct {
		rawURL string
	}
	tests := []struct {
		name string
		s    *Service
		args args
		want string
	}{
		// RawBase should only reflect what was given
		{"nil", New(), args{}, ""},
		{"noTrailingSlash", New(), args{baseURL}, baseURL},
		{"withTrailingSlash", New(), args{"https://example.com/"}, "https://example.com/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.RawBase(tt.args.rawURL)
			if !assert.Equal(t, tt.want, got.rawURL) {
				t.Errorf("%v Service.RawBase(%v) = %v, want %v", tt.name, tt.args.rawURL, got.rawURL, tt.want)
			}
		})
	}
}

func TestService_Base(t *testing.T) {
	type args struct {
		rawURL string
	}
	tests := []struct {
		name string
		s    *Service
		args args
		want string
	}{
		// RawBase should always append a trailing slash
		{"nil", New(), args{}, "/"},
		{"noTrailingSlash", New(), args{baseURL}, "https://example.com/"},
		{"withTrailingSlash", New(), args{"https://example.com/"}, "https://example.com/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Base(tt.args.rawURL)
			if !assert.Equal(t, tt.want, got.rawURL) {
				t.Errorf("%v Service.Base() = %v, want %v", tt.name, got.rawURL, tt.want)
			}
		})
	}
}

func TestService_slashIt(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		s    *Service
		args args
		want string
	}{
		{"nil", New(), args{}, "/"},
		{"noTrailingSlash", New(), args{"foo"}, "foo/"},
		{"withTrailingSlash", New(), args{"foo/"}, "foo/"},
		{"withSlashPrefix", New(), args{"/foo"}, "/foo/"},
		{"withSlashPrefixAndTrailingSlash", New(), args{"/foo/"}, "/foo/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.slashIt(tt.args.str); got != tt.want {
				t.Errorf("%v Service.slashIt() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestService_Path(t *testing.T) {
	svc := New().RawBase(baseURL)
	type args struct {
		path string
	}
	tests := []struct {
		name     string
		s        *Service
		args     args
		wantPath string
	}{
		// ON NEW SERVICE
		// nothing happens
		{"nil", svc.New(), args{}, baseURL},
		// appended
		{"normal", svc.New(), args{"foo"}, "https://example.com/foo"},
		// appended to base
		{"prefixSlash", svc.New(), args{"/foobar"}, "https://example.com/foobar"},
		// ON EXISTING SERVICE
		// appended
		{"trailingSlash", svc, args{"bar/"}, "https://example.com/bar/"},
		// appended
		{"noSlash", svc, args{"foo"}, "https://example.com/bar/foo"},
		// adds trailing slash to previous noSlash path and appended
		{"multiTrailingSlash", svc, args{"foobar/"}, "https://example.com/bar/foo/foobar/"},
		// attached to the base
		{"multiPrefixSlash", svc, args{"/foo"}, "https://example.com/foo"},
		// WITH EXISTING PATHS
		{"multipleSinglePaths", svc.New().Path("foo").Path("bar"), args{"foobar"}, "https://example.com/foo/bar/foobar"},
		{"multipleSinglePathsWithSlashPrefix", svc.New().Path("foo").Path("bar"), args{"/foobar"}, "https://example.com/foobar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Path(tt.args.path)
			if !assert.EqualValues(t, tt.wantPath, got.rawURL) {
				t.Errorf("%v Service.Path(%v) = %v, want %v", tt.name, tt.args.path, got, tt.wantPath)
			}
		})
	}
}

func TestService_Pathf(t *testing.T) {
	svc := New().RawBase(baseURL)
	type args struct {
		format string
		a      []interface{}
	}
	tests := []struct {
		name     string
		s        *Service
		args     args
		wantPath string
	}{
		{"nil", svc.New(), args{}, baseURL},
		{"normal", svc.New(), args{geocodeFormat, formatVars}, baseURL + geocodePath},
		{"prefixSlash", svc.New(), args{"/" + geocodeFormat, formatVars}, baseURL + geocodePath},
		{"trailingSlash", svc.New(), args{geocodeFormat + "/", formatVars}, baseURL + geocodePath + "/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Pathf(tt.args.format, tt.args.a...)
			if !assert.Equal(t, tt.wantPath, got.rawURL) {
				t.Errorf("%v Service.Path(%v) = %#v, want %#v", tt.name, tt.args.format, got.rawURL, tt.wantPath)
			}
		})
	}
}

func TestService_ResetPath(t *testing.T) {
	svc := New().Base(baseURL)
	tests := []struct {
		name     string
		s        *Service
		wantPath string
	}{

		{"nil", svc.New(), "https://example.com/"},
		{"multipath", svc.New().Path("foo/bar"), "https://example.com/"},
		{"path", svc.New().Path("foo"), "https://example.com/"},
		{"partialPathAndPath", svc.New().Path("foo").Path("bar"), "https://example.com/"},
		{"partialPath", svc.Path("foo"), "https://example.com/"},
		{"multiPartialPath", svc.Path("bar"), "https://example.com/"},
		{"multiPartialPath", svc.Path("foobar"), "https://example.com/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.ResetPath(); !assert.Equal(t, tt.wantPath, got.rawURL) {
				t.Errorf("%v Service.ResetPath() = %v, want %v", tt.name, got.rawURL, tt.wantPath)
			}
		})
	}
}

func TestService_Extension(t *testing.T) {
	svc := New().Base(baseURL).Path("foo/bar")
	type args struct {
		ext string
	}
	tests := []struct {
		name     string
		s        *Service
		args     args
		wantPath string
	}{
		{"nil", svc.New(), args{}, "https://example.com/foo/bar"},
		// prefix `.`
		{"noDot", svc.New(), args{"json"}, "https://example.com/foo/bar.json"},
		// prefix `.`
		{"comma", svc.New(), args{",json"}, "https://example.com/foo/bar.,json"},
		// appends
		{"dot", svc.New(), args{".json"}, "https://example.com/foo/bar.json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Extension(tt.args.ext); !assert.Equal(t, tt.wantPath, got.rawURL) {
				t.Errorf("%v Service.Extension() = %v, want %v", tt.name, got.rawURL, tt.wantPath)
			}
		})
	}
}

func TestService_QueryStruct(t *testing.T) {
	svc := New().Base(baseURL).Path("foo/bar")
	type args struct {
		queryStruct interface{}
	}
	tests := []struct {
		name   string
		s      *Service
		args   args
		wantQS []interface{}
	}{
		{"nil", svc.New(), args{}, []interface{}{}},
		// append
		{"paramsA", svc, args{paramsA}, []interface{}{paramsA}},
		// append
		{"paramsB", svc, args{paramsB}, []interface{}{paramsA, paramsB}},
		// QueryStructs do not support maps, so does nothing
		{"map", svc, args{map[string]string{"foo": "bar"}}, []interface{}{paramsA, paramsB}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.QueryStruct(tt.args.queryStruct); !assert.Equal(t, tt.wantQS, got.queryStructs) {
				t.Errorf("%v Service.QueryStruct() = %v, want %v", tt.name, got.queryStructs, tt.wantQS)
			}
		})
	}
}

func TestService_Body(t *testing.T) {
	type args struct {
		body io.Reader
	}
	tests := []struct {
		name string
		s    *Service
		args args
		want io.Reader
	}{
		{"generic", New(), args{genericBody}, genericBody},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Body(tt.args.body)
			b, _ := got.bodyProvider.Body()
			if !assert.Equal(t, tt.want, b) {
				t.Errorf("%v Service.Body() = %v, want %v", tt.name, b, tt.want)
			}
		})
	}

}

func TestService_ContentType(t *testing.T) {
	type args struct {
		ct string
	}
	tests := []struct {
		name string
		s    *Service
		args args
		want string
	}{
		{"nil", New(), args{}, ""},
		{textContentType, New(), args{textContentType}, textContentType},
		{jsonContentType + " Overwriting " + textContentType, New().ContentType(textContentType), args{jsonContentType}, jsonContentType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.ContentType(tt.args.ct); !assert.Equal(t, tt.want, got.header.Get(contentType)) {
				t.Errorf("%v Service.ContentType() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestService_PlainText(t *testing.T) {
	tests := []struct {
		name string
		s    *Service
		want string
	}{
		{"nil", New(), textContentType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.PlainText(); !assert.Equal(t, tt.want, got.header.Get(contentType)) {
				t.Errorf("%v Service.PlainText() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestService_JSON(t *testing.T) {
	tests := []struct {
		name string
		s    *Service
		want string
	}{
		{"nil", New(), jsonContentType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.JSON(); !assert.Equal(t, tt.want, got.header.Get(contentType)) {
				t.Errorf("%v Service.JSON() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestService_JPEG(t *testing.T) {
	tests := []struct {
		name string
		s    *Service
		want string
	}{
		{"nil", New(), jpegContentType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.JPEG(); !assert.Equal(t, tt.want, got.header.Get(contentType)) {
				t.Errorf("%v Service.JPEG() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestService_GIF(t *testing.T) {
	tests := []struct {
		name string
		s    *Service
		want string
	}{
		{"nil", New(), gifContentType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.GIF(); !assert.Equal(t, tt.want, got.header.Get(contentType)) {
				t.Errorf("%v Service.GIF() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestService_PNG(t *testing.T) {
	tests := []struct {
		name string
		s    *Service
		want string
	}{
		{"nil", New(), pngContentType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.PNG(); !assert.Equal(t, tt.want, got.header.Get(contentType)) {
				t.Errorf("%v Service.PNG() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestService_Form(t *testing.T) {
	tests := []struct {
		name string
		s    *Service
		want string
	}{
		{"nil", New(), formContentType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Form(); !assert.Equal(t, tt.want, got.header.Get(contentType)) {
				t.Errorf("%v Service.Form() = %v, want %v", tt.name, got.header.Get(contentType), tt.want)
			}
		})
	}
}

func TestService_BodyProvider(t *testing.T) {
	type args struct {
		body BodyProvider
	}
	type want struct {
		bodyProvider BodyProvider
		contentType  string
		shouldPanic  bool
	}

	tests := []struct {
		name string
		s    *Service
		args args
		want want
	}{
		{"nil", New(), args{}, want{nil, "", true}},
		{"generic", New(), args{bodyProvider{genericBody}}, want{bodyProvider{genericBody}, "", false}},
		{"json", New(), args{jsonBodyProvider{payload: jsonBody}}, want{jsonBodyProvider{payload: jsonBody}, jsonContentType, false}},
		{"form", New(), args{formBodyProvider{payload: formBody}}, want{formBodyProvider{payload: formBody}, formContentType, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.BodyProvider(tt.args.body)
			if !assert.Equal(t, tt.want.bodyProvider, got.bodyProvider) {
				t.Errorf("%v Service.BodyProvider() = %v, want %v", tt.name, got.bodyProvider, tt.want.bodyProvider)
			}
			if tt.want.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("%v Service.BodyProvider() should have panicked!", tt.name)
					}
				}()
				if !assert.Equal(t, tt.want.contentType, got.bodyProvider.ContentType()) {
					t.Errorf("%v Service.BodyProvider() = %v, want %v", tt.name, got.bodyProvider, tt.want.contentType)
				}
			}
		})
	}
}

func TestService_BodyJSON(t *testing.T) {
	type args struct {
		bodyJSON interface{}
	}
	type want struct {
		bodyProvider BodyProvider
		contentType  string
		shouldPanic  bool
	}
	tests := []struct {
		name string
		s    *Service
		args args
		want want
	}{
		{"json", New(), args{jsonBody}, want{jsonBodyProvider{payload: jsonBody}, jsonContentType, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.BodyJSON(tt.args.bodyJSON)
			if !assert.Equal(t, tt.want.bodyProvider, got.bodyProvider) {
				t.Errorf("%v Service.BodyJSON() = %v, want %v", tt.name, got.bodyProvider, tt.want.bodyProvider)
			}
			if tt.want.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("%v Service.BodyJSON() should have panicked!", tt.name)
					}
				}()
				if !assert.Equal(t, tt.want.contentType, got.bodyProvider.ContentType()) {
					t.Errorf("%v Service.BodyJSON() = %v, want %v", tt.name, got.bodyProvider, tt.want.contentType)
				}
			}
		})
	}
}

func TestService_BodyForm(t *testing.T) {
	type args struct {
		bodyForm interface{}
	}
	type want struct {
		bodyProvider BodyProvider
		contentType  string
		shouldPanic  bool
	}
	tests := []struct {
		name string
		s    *Service
		args args
		want want
	}{
		{"form", New(), args{formBody}, want{formBodyProvider{payload: formBody}, formContentType, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.BodyForm(tt.args.bodyForm)
			if !assert.Equal(t, tt.want.bodyProvider, got.bodyProvider) {
				t.Errorf("%v Service.BodyForm() = %v, want %v", tt.name, got.bodyProvider, tt.want.bodyProvider)
			}
			if tt.want.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("%v Service.BodyForm() should have panicked!", tt.name)
					}
				}()
				if !assert.Equal(t, tt.want.contentType, got.bodyProvider.ContentType()) {
					t.Errorf("%v Service.BodyForm() = %v, want %v", tt.name, got.bodyProvider, tt.want.contentType)
				}
			}
		})
	}
}

func TestService_Responder(t *testing.T) {
	type args struct {
		responder Responder
	}
	tests := []struct {
		name string
		s    *Service
		args args
		want Responder
	}{
		{"nil", New(), args{}, GenericResponder()},
		{"json", New(), args{JSONResponder(nil, nil)}, JSONResponder(nil, nil)},
		{"jsonSuccess", New(), args{JSONSuccessResponder(nil)}, JSONSuccessResponder(nil)},
		{"binary", New(), args{BinaryResponder(nil)}, BinaryResponder(nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Responder(tt.args.responder); !assert.Equal(t, tt.want, got.responder) {
				t.Errorf("%v Service.Responder() = %v, want %v", tt.name, got.responder, tt.want)
			}
		})
	}
}

type wxResponse struct {
	A string `json:"a"`
	B string `json:"B"`
	C time.Time `json:"C"`
}
type e struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type m struct {
	StatusCode    int    `json:"status_code"`
	TransactionID string `json:"transaction_id"`
}
type wxErr struct {
	Errors   []e  `json:"errors"`
	Metadata m    `json:"metadata"`
	Success  bool `json:"success"`
}

var (
	t1, _ = time.Parse(
		time.RFC3339,
		"2017-11-01T22:08:41+00:00")
	wantedSuccess = &wxResponse{"a success", "another success", t1}
	wantedFailure = &wxErr{Errors: []e{{Code: "EAE:INV-0001", Message: "Invalid request"}}, Metadata: m{StatusCode: 400, TransactionID: "1429140092945:1801695336"}, Success: false}
)

func newSuccess() *wxResponse {
	return &wxResponse{}
}
func newFail() *wxErr {
	return &wxErr{}
}

func TestService_JSONResponder(t *testing.T) {
	type args struct {
		success interface{}
		failure interface{}
	}
	type want struct {
		responder Responder
		err       bool
		success   interface{}
		failure   interface{}
	}

	successBody := `{"a":"a success", "b": "another success", "c": "2017-11-01T22:08:41+00:00"}`
	successHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(successBody))
	}
	//successBadHandler := func(w http.ResponseWriter, r *http.Request) {
	//	w.WriteHeader(http.StatusOK)
	//	w.Write([]byte(`{"a":"a success", "b": "another success", "c": "2017-11-01T22:08:41Z"}`))
	//}
	success204Handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
	failBody := `{"errors": [{"code": "EAE:INV-0001","message": "Invalid request"}],"metadata": {"status_code": 400,"transaction_id": "1429140092945:1801695336"},"success": false}`
	failureHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(failBody))
	}
	//failureBadHandler := func(w http.ResponseWriter, r *http.Request) {
	//	w.WriteHeader(http.StatusBadRequest)
	//	w.Write([]byte(`//{"errors": [{"code": "EAE:INV-0001","message": "Invalid request"}],"metadata": {"status_code": 400,"transaction_id": "1429140092945:1801695336"},"success": false}//`))
	//}

	tests := []struct {
		name string
		s    *Service
		h func(w http.ResponseWriter, r *http.Request)
		args args
		want want
	}{
		{"nil-success", New(), successHandler,args{}, want{JSONResponder(nil, nil), false, nil, nil}},
		{"nil-success204", New(), success204Handler,args{}, want{JSONResponder(nil, nil), false, nil, nil}},
		{"nil-failure", New(), failureHandler,args{}, want{JSONResponder(nil, nil), false, nil, nil}},
		//{"nil-successBad", New(), successBadHandler,args{}, want{JSONResponder(nil, nil), false, nil, nil}},
		//{"nil-failureBad", New(), failureBadHandler,args{}, want{JSONResponder(nil, nil), false, nil, nil}},

		{"SuccessStructOnly-success", New(), successHandler,args{newSuccess(), nil}, want{JSONResponder(newSuccess(), nil), false, wantedSuccess, nil}},
		{"SuccessStructOnly-success204", New(), success204Handler,args{newSuccess(), nil}, want{JSONResponder(newSuccess(), nil), false, newSuccess(), nil}},
		{"SuccessStructOnly-failure", New(), failureHandler, args{newSuccess(), nil}, want{JSONResponder(newSuccess(), nil), false, newSuccess(), nil}},
		//{"SuccessStructOnly-badSuccess", New(), successBadHandler, args{newSuccess(), nil}, want{JSONResponder(newSuccess(), nil), true, newSuccess(), nil}},
		//{"SuccessStructOnly-badFailure", New(), failureBadHandler, args{newSuccess(), nil}, want{JSONResponder(newSuccess(), nil), true, newSuccess(), nil}},

		{"FailureStructOnly-success", New(), successHandler, args{nil, newFail()}, want{JSONResponder(nil, newFail()), false, nil, wantedFailure}},
		{"FailureStructOnly-success204", New(), success204Handler, args{nil, newFail()}, want{JSONResponder(nil, newFail()), false, nil, wantedFailure}},
		{"FailureStructOnly-failure", New(), failureHandler, args{nil, newFail()}, want{JSONResponder(nil, newFail()), false, nil, wantedFailure}},

		{"both-success", New(), successHandler, args{newSuccess(), newFail()}, want{JSONResponder(newSuccess(), newFail()), false, wantedSuccess, wantedFailure}},
		{"both-success204", New(), success204Handler, args{newSuccess(), newFail()}, want{JSONResponder(newSuccess(), newFail()), false, newSuccess(), newFail()}},
		{"both-failure", New(), failureHandler, args{newSuccess(), newFail()}, want{JSONResponder(newSuccess(), newFail()), false, newSuccess(), wantedFailure}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.JSONResponder(tt.args.success, tt.args.failure); !assert.Equal(t, tt.want.responder, got.responder) {
				t.Errorf("%v Service.JSONResponder() = %v, want %v", tt.name, got.responder, tt.want.responder)
			}

			var err error

			req, _ := tt.s.New().Base(baseURL).Path("success").Request()
			recorder := httptest.NewRecorder()
			tt.h(recorder,req)
			tt.s.responder.Respond(req, recorder.Result(), nil)
			_, err = tt.s.responder.DoResponse()

			if !assert.Equal(t, tt.want.err, err != nil) {
				t.Errorf("%v.Error Service.JSONResponder() error = %v, wantErr %v", tt.name, err, tt.want.err)
				return
			}
			if success := tt.s.responder.GetSuccess(); !assert.Equal(t, (interface{})(tt.want.success), success) {
				t.Errorf("%v.Success Service.JSONResponder() = %v, want %v", tt.name, success, tt.want.success)
			}

			//// Test Bad Success
			//sbw := httptest.NewRecorder()
			//successHandler(sbw, req)
			//tt.s.responder.Respond(req, sbw.Result(), nil)
			//_, err = tt.s.responder.DoResponse()
			//
			//if !assert.Equal(t, tt.want.err, err != nil) {
			//	t.Errorf("%v.Error Service.JSONResponder() error = %v, wantErr %v", tt.name, err, tt.want.err)
			//	return
			//}
			//if success := tt.s.responder.GetSuccess(); !assert.Equal(t, (interface{})(tt.want.success), success) {
			//	t.Errorf("%v.Success Service.JSONResponder() = %v, want %v", tt.name, success, tt.want.success)
			//}
			//
			//// Test Failure
			//fw := httptest.NewRecorder()
			//failureHandler(fw, req)
			//tt.s.responder.Respond(req, fw.Result(), nil)
			//_, err = tt.s.responder.DoResponse()
			//
			//if !assert.Equal(t, tt.want.err, err != nil) {
			//	t.Errorf("%v.Error Service.JSONResponder() error = %v, wantErr %v", tt.name, err, tt.want.err)
			//	return
			//}
			//if fail := tt.s.responder.GetFailure(); !assert.Equal(t, (interface{})(tt.want.failure), fail) {
			//	t.Errorf("%v.Failure Service.JSONResponder() = %v, want %v", tt.name, fail, tt.want.failure)
			//}
		})
	}
}

func TestService_JSONSuccessResponder(t *testing.T) {
	type args struct {
		success interface{}
	}
	type want struct {
		responder Responder
		err       bool
		success   interface{}
		failure   interface{}
	}
	tests := []struct {
		name string
		s    *Service
		args args
		want want
	}{
		{"nil", New(), args{}, want{JSONResponder(nil, nil), false, nil, nil}},
		{"SuccessStructOnly", New(), args{newSuccess()}, want{JSONResponder(newSuccess(), nil), false, wantedSuccess, nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.JSONSuccessResponder(tt.args.success); !assert.Equal(t, tt.want.responder, got.responder) {
				t.Errorf("%v Service.JSONSuccessResponder() = %v, want %v", tt.name, got.responder, tt.want.responder)
			}

			req, _ := tt.s.New().Base(baseURL).Path("success").Request()
			successHandler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"a":"a success", "b": "another success", "c": "2017-11-01T22:08:41+00:00"}`))
			}
			failureHandler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"errors": [{"code": "EAE:INV-0001","message": "Invalid request"}],"metadata": {"status_code": 400,"transaction_id": "1429140092945:1801695336"},"success": false}`))
			}

			// Test Success
			sw := httptest.NewRecorder()
			successHandler(sw, req)
			tt.s.responder.Respond(req, sw.Result(), nil)
			_, err := tt.s.responder.DoResponse()

			if !assert.Equal(t, tt.want.err, err != nil) {
				t.Errorf("%v.Error Service.JSONResponder() error = %v, wantErr %v", tt.name, err, tt.want.err)
				return
			}
			if success := tt.s.responder.GetSuccess(); !assert.Equal(t, (interface{})(tt.want.success), success) {
				t.Errorf("%v.Success Service.JSONResponder() = %v, want %v", tt.name, success, tt.want.success)
			}

			// Test Failure
			fw := httptest.NewRecorder()
			failureHandler(fw, req)
			tt.s.responder.Respond(req, fw.Result(), nil)
			_, err = tt.s.responder.DoResponse()

			if !assert.Equal(t, tt.want.err, err != nil) {
				t.Errorf("%v.Error Service.JSONSuccessResponder() error = %v, wantErr %v", tt.name, err, tt.want.err)
				return
			}
			if fail := tt.s.responder.GetFailure(); !assert.Equal(t, (interface{})(tt.want.failure), fail) {
				t.Errorf("%v.Failure Service.JSONSuccessResponder() = %v, want %v", tt.name, fail, tt.want.failure)
			}
		})
	}
}

func TestService_BinaryResponder(t *testing.T) {
	type args struct {
		failure interface{}
	}
	type want struct {
		responder Responder
		err       bool
		success   interface{}
		failure   interface{}
	}

	tests := []struct {
		name string
		s    *Service
		args args
		want want
	}{
		{"nil", New(), args{}, want{BinaryResponder(nil), false, []byte(`{"a":"a success", "b": "another success", "c": "2017-11-01T22:08:41+00:00"}`), nil}},
		{"withFailure", New(), args{newFail()}, want{BinaryResponder(newFail()), false, []byte(`{"a":"a success", "b": "another success", "c": "2017-11-01T22:08:41+00:00"}`), wantedFailure}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.BinaryResponder(tt.args.failure); !assert.Equal(t, tt.want.responder, got.responder) {
				t.Errorf("%v Service.BinaryResponder() = %v, want %v", tt.name, got.responder, tt.want.responder)
			}

			req, _ := tt.s.New().Base(baseURL).Path("success").Request()
			successHandler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"a":"a success", "b": "another success", "c": "2017-11-01T22:08:41+00:00"}`))
			}
			failureHandler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"errors": [{"code": "EAE:INV-0001","message": "Invalid request"}],"metadata": {"status_code": 400,"transaction_id": "1429140092945:1801695336"},"success": false}`))
			}

			var err error

			// Test Success
			sw := httptest.NewRecorder()
			successHandler(sw, req)
			tt.s.responder.Respond(req, sw.Result(), nil)
			_, err = tt.s.responder.DoResponse()

			if !assert.Equal(t, tt.want.err, err != nil) {
				t.Errorf("%v.Error Service.BinaryResponder() error = %v, wantErr %v", tt.name, err, tt.want.err)
				return
			}
			if success := tt.s.responder.GetSuccess(); !assert.Equal(t, (interface{})(tt.want.success), success) {
				t.Errorf("%v.Success Service.BinaryResponder() = %v, want %v", tt.name, success, tt.want.success)
			}

			// Test Failure
			fw := httptest.NewRecorder()
			failureHandler(fw, req)
			tt.s.responder.Respond(req, fw.Result(), nil)
			_, err = tt.s.responder.DoResponse()

			if !assert.Equal(t, tt.want.err, err != nil) {
				t.Errorf("%v.Error Service.BinaryResponder() error = %v, wantErr %v", tt.name, err, tt.want.err)
				return
			}
			if fail := tt.s.responder.GetFailure(); !assert.Equal(t, (interface{})(tt.want.failure), fail) {
				t.Errorf("%v.Failure Service.JSONResponder() = %v, want %v", tt.name, fail, tt.want.failure)
			}
		})
	}
}

func TestService_BinarySuccessResponder(t *testing.T) {
	type want struct {
		responder Responder
		err       bool
		success   interface{}
		failure   interface{}
	}
	tests := []struct {
		name string
		s    *Service
		want want
	}{
		{"success", New(), want{BinarySuccessResponder(), false, []byte(`{"a":"a success", "b": "another success", "c": "2017-11-01T22:08:41+00:00"}`), nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.BinarySuccessResponder(); !assert.Equal(t, tt.want.responder, got.responder) {
				t.Errorf("%v Service.BinarySuccessResponder() = %v, want %v", tt.name, got.responder, tt.want.responder)
			}

			req, _ := tt.s.New().Base(baseURL).Path("success").Request()
			successHandler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"a":"a success", "b": "another success", "c": "2017-11-01T22:08:41+00:00"}`))
			}
			failureHandler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"errors": [{"code": "EAE:INV-0001","message": "Invalid request"}],"metadata": {"status_code": 400,"transaction_id": "1429140092945:1801695336"},"success": false}`))
			}

			// Test Success
			sw := httptest.NewRecorder()
			successHandler(sw, req)
			tt.s.responder.Respond(req, sw.Result(), nil)
			_, err := tt.s.responder.DoResponse()

			if !assert.Equal(t, tt.want.err, err != nil) {
				t.Errorf("%v.Error Service.BinarySuccessResponder() error = %v, wantErr %v", tt.name, err, tt.want.err)
				return
			}
			if success := tt.s.responder.GetSuccess(); !assert.Equal(t, (interface{})(tt.want.success), success) {
				t.Errorf("%v.Success Service.BinarySuccessResponder() = %v, want %v", tt.name, success, tt.want.success)
			}

			// Test Failure
			fw := httptest.NewRecorder()
			failureHandler(fw, req)
			tt.s.responder.Respond(req, fw.Result(), nil)
			_, err = tt.s.responder.DoResponse()

			if !assert.Equal(t, tt.want.err, err != nil) {
				t.Errorf("%v.Error Service.BinarySuccessResponder() error = %v, wantErr %v", tt.name, err, tt.want.err)
				return
			}
			if fail := tt.s.responder.GetFailure(); !assert.Equal(t, (interface{})(tt.want.failure), fail) {
				t.Errorf("%v.Failure Service.BinarySuccessResponder() = %v, want %v", tt.name, fail, tt.want.failure)
			}
		})
	}
}

func TestService_Request(t *testing.T) {
	svc := New()
	type want struct {
		req *http.Request
		err bool
	}
	tests := []struct {
		name string
		s    *Service
		want want
	}{
		{"nil", svc.New(), func() want {
			req, err := http.NewRequest("GET", "", nil)
			return want{req, err != nil}
		}()},
		{"err", svc.New().BodyJSON(http.Request{}), want{nil, true}},
		{"rawBase", svc.RawBase(baseURL).New(), func() want {
			req, err := http.NewRequest("GET", baseURL, nil)
			return want{req, err != nil}
		}()},
		{"base", svc.Base(baseURL).New(), func() want {
			req, err := http.NewRequest("GET", "https://example.com/", nil)
			return want{req, err != nil}
		}()},
		{"path", svc.Path("foo").New(), func() want {
			req, err := http.NewRequest("GET", "https://example.com/foo", nil)
			return want{req, err != nil}
		}()},
		{"query", svc.QueryStruct(paramsA).New(), func() want {
			req, err := http.NewRequest("GET", "https://example.com/foo?limit=30", nil)
			return want{req, err != nil}
		}()},
		{"body", svc.Body(genericBody).New(), func() want {
			req, err := http.NewRequest("GET", "https://example.com/foo?limit=30", genericBody)
			return want{req, err != nil}
		}()},
		{"resetPath", svc.ResetPath().New(), func() want {
			req, err := http.NewRequest("GET", "https://example.com/?limit=30", genericBody)
			return want{req, err != nil}
		}()},
		{"reset", svc.Reset().New(), func() want {
			req, err := http.NewRequest("GET", "", nil)
			return want{req, err != nil}
		}()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Request()
			if (err != nil) != tt.want.err {
				t.Errorf("%v Service.Request() error = %v, wantErr %v", tt.name, err, tt.want.err)
				return
			}
			//if tt.want.err && (err != nil) {
			//	return
			//}
			if !assert.Equal(t, tt.want.req.URL, got.URL) {
				t.Errorf("%v.URL Service.Request() = %v, want %v", tt.name, got.URL, tt.want.req.URL)
			}
			if !assert.Equal(t, tt.want.req.Body, got.Body) {
				t.Errorf("%v.Body Service.Request() = %v, want %v", tt.name, got.Body, tt.want.req.Body)
			}
			if !assert.Equal(t, tt.want.req.ContentLength, got.ContentLength) {
				t.Errorf("%v.ContentLength Service.Request() = %v, want %v", tt.name, got.ContentLength, tt.want.req.ContentLength)
			}
			if !assert.Equal(t, tt.want.req.Header, got.Header) {
				t.Errorf("%v.Header Service.Request() = %v, want %v", tt.name, got.Header, tt.want.req.Header)
			}
			if !assert.Equal(t, tt.want.req.Host, got.Host) {
				t.Errorf("%v.Host Service.Request() = %v, want %v", tt.name, got.Host, tt.want.req.Host)
			}
			if !assert.Equal(t, tt.want.req.Form, got.Form) {
				t.Errorf("%v.Form Service.Request() = %v, want %v", tt.name, got.Form, tt.want.req.Form)
			}
			if !assert.Equal(t, tt.want.req.PostForm, got.PostForm) {
				t.Errorf("%v.PostForm Service.Request() = %v, want %v", tt.name, got.PostForm, tt.want.req.PostForm)
			}
		})
	}
}

func TestService_AsyncRequest(t *testing.T) {
	svc := New()
	type args struct {
		responder Responder
	}
	type want struct {
		req       *http.Request
		err       bool
		responder Responder
	}
	tests := []struct {
		name string
		s    *Service
		args args
		want want
	}{
		{"nil", svc.New(), args{}, func() want {
			req, err := http.NewRequest("GET", "", nil)
			return want{req, err != nil, GenericResponder()}
		}()},
		{"rawBase", svc.RawBase(baseURL).New(), args{GenericResponder()}, func() want {
			req, err := http.NewRequest("GET", baseURL, nil)
			return want{req, err != nil, GenericResponder()}
		}()},
		{"base", svc.Base(baseURL).New(), args{GenericResponder()}, func() want {
			req, err := http.NewRequest("GET", "https://example.com/", nil)
			return want{req, err != nil, GenericResponder()}
		}()},
		{"path", svc.Path("foo").New(), args{GenericResponder()}, func() want {
			req, err := http.NewRequest("GET", "https://example.com/foo", nil)
			return want{req, err != nil, GenericResponder()}
		}()},
		{"query", svc.QueryStruct(paramsA).New(), args{GenericResponder()}, func() want {
			req, err := http.NewRequest("GET", "https://example.com/foo?limit=30", nil)
			return want{req, err != nil, GenericResponder()}
		}()},
		{"body", svc.Body(genericBody).New(), args{GenericResponder()}, func() want {
			req, err := http.NewRequest("GET", "https://example.com/foo?limit=30", genericBody)
			return want{req, err != nil, GenericResponder()}
		}()},
		{"resetPath", svc.ResetPath().New(), args{GenericResponder()}, func() want {
			req, err := http.NewRequest("GET", "https://example.com/?limit=30", genericBody)
			return want{req, err != nil, GenericResponder()}
		}()},
		{"reset", svc.Reset().New(), args{GenericResponder()}, func() want {
			req, err := http.NewRequest("GET", "", nil)
			return want{req, err != nil, GenericResponder()}
		}()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.AsyncRequest(tt.args.responder)
			if (got.Error != nil) != tt.want.err {
				t.Errorf("%v Service.Request() error = %v, wantErr %v", tt.name, got.Error, tt.want.err)
				return
			}
			if !assert.Equal(t, tt.want.responder, got.responder) {
				t.Errorf("%v.URL Service.AsyncRequest() = %v, want %v", tt.name, got.responder, tt.want.responder)
			}
			if !assert.Equal(t, tt.want.req.URL, got.Request.URL) {
				t.Errorf("%v.URL Service.AsyncRequest() = %v, want %v", tt.name, got.Request.URL, tt.want.req.URL)
			}
			if !assert.Equal(t, tt.want.req.Body, got.Request.Body) {
				t.Errorf("%v.Body Service.AsyncRequest() = %v, want %v", tt.name, got.Request.Body, tt.want.req.Body)
			}
			if !assert.Equal(t, tt.want.req.ContentLength, got.Request.ContentLength) {
				t.Errorf("%v.ContentLength Service.AsyncRequest() = %v, want %v", tt.name, got.Request.ContentLength, tt.want.req.ContentLength)
			}
			if !assert.Equal(t, tt.want.req.Header, got.Request.Header) {
				t.Errorf("%v.Header Service.AsyncRequest() = %v, want %v", tt.name, got.Request.Header, tt.want.req.Header)
			}
			if !assert.Equal(t, tt.want.req.Host, got.Request.Host) {
				t.Errorf("%v.Host Service.AsyncRequest() = %v, want %v", tt.name, got.Request.Host, tt.want.req.Host)
			}
			if !assert.Equal(t, tt.want.req.Form, got.Request.Form) {
				t.Errorf("%v.Form Service.AsyncRequest() = %v, want %v", tt.name, got.Request.Form, tt.want.req.Form)
			}
			if !assert.Equal(t, tt.want.req.PostForm, got.Request.PostForm) {
				t.Errorf("%v.PostForm Service.AsyncRequest() = %v, want %v", tt.name, got.Request.PostForm, tt.want.req.PostForm)
			}
		})
	}
}

func TestService_addQueryStructs(t *testing.T) {
	type args struct {
		reqURL       *url.URL
		queryStructs []interface{}
	}
	u, _ := url.Parse(baseURL)
	tests := []struct {
		name    string
		s       *Service
		args    args
		want    string
		wantErr bool
	}{
		{"nil", New(), args{}, "", true},
		{"nil", New(), args{u, []interface{}{paramsA, paramsB}}, "count=25&kind_name=recent&limit=30", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("%v Service.addQueryStructs() should have panicked!", tt.name)
					}
				}()
			}

			tt.s.addQueryStructs(tt.args.reqURL, tt.args.queryStructs)
			if u.RawQuery != tt.want {
				t.Errorf("%v Service.addQueryStructs() req.URL.RawQuery = %v, want %v", tt.name, u.RawQuery, tt.want)
			}
		})
	}
}

func Test_addHeaders(t *testing.T) {
	type args struct {
		req    *http.Request
		header http.Header
	}
	req, _ := New().Base(baseURL).Request()
	tests := []struct {
		name string
		args args
		len  int
	}{
		{"nil", args{req, nil}, 0},
		{"single", args{req, http.Header{"Content-Type": []string{"application/json"}}}, 1},
		{"multiple", args{req, http.Header{"Content-Type": []string{"application/json"}, "TWC-Viking": []string{"Be a Weather Viking"}}}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addHeaders(tt.args.req, tt.args.header)
			if got := len(tt.args.req.Header); got != tt.len {
				t.Errorf("%v Service.addHeaders() len = %v, want %v", tt.name, got, tt.len)
			}
		})
	}
}

func TestService_ReceiveSuccess(t *testing.T) {
	svcSuccess := New().Base(baseURL).Path("success")
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	successBody := `{"id": 1234567890, "name": "Meteor Rocks!"}]`
	httpmock.RegisterResponder("GET", baseURL+"/success",
		httpmock.NewStringResponder(http.StatusOK, successBody))

	var successV *wxResponse

	type args struct {
		successV interface{}
	}
	type want struct {
		response *http.Response
		err      bool
		success  interface{}
	}
	tests := []struct {
		name string
		s    *Service
		args args
		want want
	}{
		{"success-nil", svcSuccess.New(), args{}, want{&http.Response{Status: "200", StatusCode: http.StatusOK, Header: http.Header{}, Body: httpmock.NewRespBodyFromString(successBody)}, false, nil}},
		{"success", svcSuccess.New(), args{successV}, want{&http.Response{Status: "200", StatusCode: http.StatusOK, Header: http.Header{}, Body: httpmock.NewRespBodyFromString(successBody)}, false, wantedSuccess}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.ReceiveSuccess(tt.args.successV)
			if (err != nil) != tt.want.err {
				t.Errorf("%v Service.ReceiveSuccess() error = %v, wantErr %v", tt.name, err, tt.want.err)
				return
			}
			if !assert.Equal(t, tt.want.response, got) {
				t.Errorf("%v Service.ReceiveSuccess() = %v, want %v", tt.name, got, tt.want.response)
			}
			if tt.args.successV != nil && !assert.Equal(t, tt.args.successV, successV) {
				t.Errorf("%v Service.ReceiveSuccess() = %v, want %v", tt.name, (interface{})(successV), tt.args.successV)
			}
		})
	}
}

func TestService_Receive(t *testing.T) {
	svcSuccess := New().Base(baseURL).Path("success")
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	successBody := `{"id": 1234567890, "name": "Meteor Rocks!"}]`
	httpmock.RegisterResponder("GET", baseURL+"/success",
		httpmock.NewStringResponder(http.StatusOK, successBody))

	type args struct {
		successV interface{}
		failureV interface{}
	}
	type want struct {
		response *http.Response
		err      bool
		success  interface{}
		failure  interface{}
	}
	type test struct {
		name string
		s    *Service
		args args
		want want
	}
	tests := []test{
		{"success-nil", svcSuccess.New(), args{}, want{&http.Response{Status: "200", StatusCode: http.StatusOK, Header: http.Header{}, Body: httpmock.NewRespBodyFromString(successBody)}, false, nil, nil}},
		{"success", svcSuccess.New(), args{newSuccess(), newFail()}, want{&http.Response{Status: "200", StatusCode: http.StatusOK, Header: http.Header{}, Body: httpmock.NewRespBodyFromString(successBody)}, false, wantedSuccess, nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Receive(tt.args.successV, tt.args.failureV)
			if (err != nil) != tt.want.err {
				t.Errorf("%v Service.Receive() error = %v, wantErr %v", tt.name, err, tt.want.err)
				return
			}
			if !assert.Equal(t, tt.want.response, got) {
				t.Errorf("%v Service.Receive() = %v, want %v", tt.name, got, tt.want.response)
			}
		})
	}

	svcFail := New().Base(baseURL).Path("fail")
	failBody := `{"errors": [{"code": "EAE:INV-0001","message": "Invalid request"}],"metadata": {"status_code": 400,"transaction_id": "1429140092945:1801695336"},"success": false}`
	httpmock.RegisterResponder("GET", baseURL+"/fail",
		httpmock.NewStringResponder(http.StatusBadRequest, failBody))

	tests = []test{
		{"fail-nil", svcFail.New(), args{}, want{&http.Response{Status: "400", StatusCode: http.StatusBadRequest, Header: http.Header{}, Body: httpmock.NewRespBodyFromString(failBody)}, false, nil, nil}},
		{"fail", svcFail.New(), args{newSuccess(), newFail()}, want{&http.Response{Status: "400", StatusCode: http.StatusBadRequest, Header: http.Header{}, Body: httpmock.NewRespBodyFromString(failBody)}, false, nil, wantedFailure}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Receive(tt.args.successV, tt.args.failureV)
			if (err != nil) != tt.want.err {
				t.Errorf("%v Service.Receive() error = %v, wantErr %v", tt.name, err, tt.want.err)
				return
			}
			if !assert.Equal(t, tt.want.response, got) {
				t.Errorf("%v Service.Receive() = %v, want %v", tt.name, got, tt.want.response)
			}
		})
	}
}

func TestService_GetResponder(t *testing.T) {
	success := newSuccess()
	fail := newFail()
	svc := New().Base(baseURL)
	tests := []struct {
		name string
		s    *Service
		want Responder
	}{
		{"generic", svc.New(), GenericResponder()},
		{"json", svc.New().JSONResponder(success, fail), JSONResponder(success, fail)},
		{"jsonSuccess", svc.New().JSONSuccessResponder(success), JSONSuccessResponder(success)},
		{"binary", svc.New().BinaryResponder(fail), BinaryResponder(fail)},
		{"binarySuccess", svc.New().BinarySuccessResponder(), BinarySuccessResponder()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.GetResponder(); !assert.Equal(t, tt.want, got) {
				t.Errorf("Service.GetResponder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_GetSuccess(t *testing.T) {
	success := newSuccess()
	fail := newFail()
	svc := New().Base(baseURL).JSONResponder(success, fail)
	tests := []struct {
		name string
		s    *Service
		want interface{}
	}{
		{"nil", svc, (interface{})(success)},
		{"json", svc.New().JSONResponder(success, fail), (interface{})(success)},
		{"jsonSuccess", svc.New().JSONSuccessResponder(success), (interface{})(success)},
		{"binary", svc.New().BinaryResponder(fail), (interface{})(&[]byte{})},
		{"binarySuccess", svc.New().BinarySuccessResponder(), (interface{})(&[]byte{})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.GetSuccess(); !assert.Equal(t, tt.want, got) {
				t.Errorf("%v Service.GetSuccess() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestService_GetFailure(t *testing.T) {
	success := newSuccess()
	fail := newFail()
	svc := New().Base(baseURL).JSONResponder(success, fail)
	tests := []struct {
		name string
		s    *Service
		want interface{}
	}{
		{"nil", svc, (interface{})(fail)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.GetFailure(); !assert.Equal(t, tt.want, got) {
				t.Errorf("Service.GetFailure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_Do(t *testing.T) {
	svc := New().Base(baseURL).Path("path")
	req, _ := svc.Request()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	body := `{"id": 1234567890, "name": "Meteor Rocks!"}]`
	httpmock.RegisterResponder("GET", baseURL+"/path",
		httpmock.NewStringResponder(200, body))

	type args struct {
		req *http.Request
	}
	tests := []struct {
		name    string
		s       *Service
		args    args
		want    *http.Response
		wantErr bool
	}{
		{"basic", svc, args{req}, &http.Response{Status: "200", StatusCode: 200, Header: http.Header{}, Body: httpmock.NewRespBodyFromString(body)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Do(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("%v Service.Do() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if !assert.Equal(t, tt.want, got) {
				t.Errorf("%v Service.Do() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestService_DoAsync(t *testing.T) {
	svc := New().Base(baseURL).Path("something")

	success := newSuccess()
	fail := newFail()
	respondr := JSONResponder(success, fail)
	svc.Responder(respondr)

	r, e := svc.Request()
	reqs := NewAsyncDoers(&AsyncRequest{responder: respondr, Request: r, Error: e, service: svc}, &AsyncRequest{responder: respondr, Request: r, Error: e, service: svc})
	//reqs := NewAsyncDoers([]AsyncRequest{&AsyncRequest{responder: respondr, Request: r, Error: e, service: svc}, &AsyncRequest{responder: respondr, Request: r, Error: e, service: svc}})
	//reqs := []AsyncDoer{&AsyncRequest{responder: respondr, Request: r, Error: e, service: svc}, &AsyncRequest{responder: respondr, Request: r, Error: e, service: svc}}
	//reqs := []AsyncRequest{{responder: respondr, Request: r, Error: e}, {responder: respondr, Request: r, Error: e}}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	//body := `{"a": "a success", "b": "another success"}`
	body := `{"a":"a success", "b": "another success", "c": "2017-11-01T22:08:41+00:00"}`
	httpmock.RegisterResponder("GET", baseURL+"/something",
		httpmock.NewStringResponder(http.StatusOK, body))

	type args struct {
		//reqs []AsyncRequest
		reqs []AsyncDoer
	}
	tests := []struct {
		name string
		s    *Service
		args args
		want []*AsyncResponse
	}{
		{"default", svc, args{reqs}, []*AsyncResponse{
			//{respondr, &http.Response{Status: "200", StatusCode: http.StatusOK, Header: http.Header{}, Body: httpmock.NewRespBodyFromString(body)}, wantedSuccess, &wxErr{}, nil},
			{respondr, &http.Response{Status: "200", StatusCode: http.StatusOK, Header: http.Header{}, Body: httpmock.NewRespBodyFromString(body)}, nil},
			{respondr, &http.Response{Status: "200", StatusCode: http.StatusOK, Header: http.Header{}, Body: httpmock.NewRespBodyFromString(body)}, nil},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.DoAsync(tt.args.reqs); !assert.Equal(t, tt.want, got) {
				//if got := tt.s.DoAsync(tt.args.reqs); !reflect.DeepEqual(tt.want, got) {
				t.Errorf("%v Service.DoAsync() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_isOk(t *testing.T) {
	type args struct {
		statusCode int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// Good ones
		{"StatusOK", args{http.StatusOK}, true},
		{"StatusCreated", args{http.StatusCreated}, true},
		{"StatusAccepted", args{http.StatusAccepted}, true},
		{"StatusNonAuthoritativeInfo", args{http.StatusNonAuthoritativeInfo}, true},
		{"StatusNoContent", args{http.StatusNoContent}, true},
		{"StatusResetContent", args{http.StatusResetContent}, true},
		{"StatusPartialContent", args{http.StatusPartialContent}, true},
		{"StatusMultiStatus", args{http.StatusMultiStatus}, true},
		{"StatusAlreadyReported", args{http.StatusAlreadyReported}, true},
		{"StatusIMUsed", args{http.StatusIMUsed}, true},

		// Bad ones
		{"StatusBadRequest", args{http.StatusBadRequest}, false},
		{"StatusInternalServerError", args{http.StatusInternalServerError}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOk(tt.args.statusCode); got != tt.want {
				t.Errorf("%v isOk() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
