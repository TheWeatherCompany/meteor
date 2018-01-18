# Meteor

Meteor is a Go HTTP client library for creating and sending API requests.

Meteor stores HTTP Request properties to simplify sending requests and decoding responses. Check [usage](#usage) or the [examples](examples) to learn how to compose a Meteor into your API client.

### Features

* Method Setters: Get/Post/Put/Patch/Delete/Head
* Add or Set Request Headers
* Manipulate the URL string quickly & easily
* Encode structs into URL query parameters
* Use a Body Provider for Request body manipuation:
  * Encode a raw string
  * Encode a form
  * Encode JSON
  * Create your own body provider!
* Use a response providers (Responder) for response manipulation:
  * Receive JSON success and/or failure responses
  * Receive Binary success responses (optionally with JSON failure responses)
  * Create your own!
* Make the requests _*asynchronously*_.
* Reuses the connection for faster subsequent calls.

## Install

    go get github.ibm.com/TheWeatherCompany/Meteor

## Usage

Use a Meteor to set path, method, header, query, or body properties and create an `http.Request`.

```go
type Params struct {
    APIKey string `url:"apiKey"`
    Format string `url:"format,omitempty"`
}

params := &Params{APIKey: "mykey"}
req, err := Meteor.New().Get("https://example.com").QueryStruct(params).Request()
```

### Base

Use `Base` to create the base subdomain.domain.tld of the URL. Base will always append a trailing slash. If you would like to create the base without the trailing slash, use `RawBase`. To use `Path`, you must use either `Base` or `RawBase` with the URL passed containing a trailing slash.

#### `Base`

```go
// These create a GET request to https://example.com/
req, err := Meteor.New().Base("https://example.com")
req, err := Meteor.New().Base("https://example.com/")
req, err := Meteor.New().BaseRaw("https://example.com/")
```

#### `RawBase`

```go
// creates a GET request to https://example.com
req, err := Meteor.New().RawBase("https://example.com")
```

### Path

Path methods are used to set or extend the URL for created Requests

#### `Path`/`Pathf`
Use `Path` or `Pathf` to append full paths or partial paths (with a trailing slash). The path requires a *trailing slash*, not a slash prefix. Adding a slash prefix will reset the path.

```go
// creates a GET request to https://example.com/foo/bar
req, err := Meteor.New().Base("https://example.com/").Path("foo/").Path("bar").Request()

// creates a GET request to https://example.com/bar
req, err := Meteor.New().Base("https://example.com/").Path("foo/").Path("/bar").Request()
req, err := Meteor.New().Base("https://example.com/").Path("foo").Path("/bar").Request()

// creates a GET request to https://example.com/foo/{dynamicPath1}/{dynamicPath2}
req, err := Meteor.New().Base("https://example.com/").Pathf("foo/%v/%v", dynamicPath1, dynamicPath2).Request()
```

If the path does not have a trailing slash and another `Path` or `Pathf` is added, it will over-write the last part of the path.

```go
// creates a GET request to https://example.com/foo/foobar
req, err := Meteor.New().Base("https://example.com/").Path("foo/").Path("bar").Path("foobar").Request()
```

#### `ResetPath`
Use `ResetPath` to reset the path.

```go
// both create a GET request to https://example.com/
req, err := Meteor.New().Base("https://example.com/").PartialPath("foo").PartialPath("bar").ResetPath().Request()
```

#### `Extension`
Use `Extension` to add an extention to the path.

```go
// both create a GET request to https://example.com/foo/bar.json
req, err := Meteor.New().Base("https://example.com/").PartialPath("foo").PartialPath("bar").Extension("json").Request()
req, err := Meteor.New().Base("https://example.com/").PartialPath("foo").PartialPath("bar").Extension(".json").Request()
```

All path methods can be used together.

```go
// creates a GET request to https://example.com/foo/bar.json
req, err := Meteor.New().Base("https://example.com/").PartialPath("foo").Path("bar.json").Request()
```

### Method

Use `Get`, `Post`, `Put`, `Patch`, `Delete`, or `Head` sets the appropriate HTTP method. `Method` allows you to set the HTTP Method. All allow you to set the path (like `Path`) _optionally_.

```go
req, err := Meteor.New().Post("http://example.com/foo/bar")
req, err := Meteor.New().Method("POST", "http://example.com/foo/bar")
req, err := Meteor.New().Path("http://example.com/foo/bar").Post()
req, err := Meteor.New().Base("http://example.com/").Path("foo/bar").Post()
```

You can also use `Getf`, `Postf`, `Putf`, `Patchf`, `Deletef`, `Headf`, or `Methodf` to resolve a dynamic method.

```go
// creates a GET request to https://example.com/foo/{dynamicPath1}/{dynamicPath2}
req, err := Meteor.New().Base("https://example.com/").Postf("foo/%v/%v", dynamicPath1, dynamicPath2).Request()
```

### Headers

`Add` or `Set` headers for requests created by a Meteor.

```go
s := Meteor.New().Base(baseUrl).Set("User-Agent", "Gophergram API Client")
req, err := s.New().Get("gophergram/list").Request()
```

### Query

#### QueryStruct

Define [url tagged structs](https://godoc.org/github.com/google/go-querystring/query). Use `QueryStruct` to encode a struct as query parameters on requests.

```go
// Github Issue Parameters
type IssueParams struct {
    Filter    string `url:"filter,omitempty"`
    State     string `url:"state,omitempty"`
    Labels    string `url:"labels,omitempty"`
    Sort      string `url:"sort,omitempty"`
    Direction string `url:"direction,omitempty"`
    Since     string `url:"since,omitempty"`
}
```

```go
githubBase := meteor.New().Base("https://api.github.com/").Client(httpClient)

params := &IssueParams{Sort: "updated", State: "open"}
req, err := githubBase.New().Getf("repos/%s/%s/issues", owner, repo).QueryStruct(params).Request()
```

### Body

#### JSON Body with Github

Define [JSON tagged structs](https://golang.org/pkg/encoding/json/). Use `BodyJSON` to JSON encode a struct as the Body on requests.

```go
type IssueRequest struct {
    Title     string   `json:"title,omitempty"`
    Body      string   `json:"body,omitempty"`
    Assignee  string   `json:"assignee,omitempty"`
    Milestone int      `json:"milestone,omitempty"`
    Labels    []string `json:"labels,omitempty"`
}

wxb := meteor.New().Base("https://api.github.com/").Client(httpClient)
body := &IssueRequest{
    Title: "Test title",
    Body:  "Some issue",
}
req, err := githubBase.New().Postf("repos/%s/%s/issues", owner, repo).BodyJSON(body).Request()

// or alternatively
req, err := githubBase.New().Path("repos/%s/%s/issues", owner, repo).Post().BodyJSON(body).Request()
```

Requests will include an `application/json` Content-Type header.

#### Form Body with Twitter

Define [url tagged structs](https://godoc.org/github.com/google/go-querystring/query). Use `BodyForm` to form url encode a struct as the Body on requests.

```go
type StatusUpdateParams struct {
    Status             string   `url:"status,omitempty"`
    InReplyToStatusId  int64    `url:"in_reply_to_status_id,omitempty"`
    MediaIds           []int64  `url:"media_ids,omitempty,comma"`
}

tweetParams := &StatusUpdateParams{Status: "writing some Go"}
req, err := twitterBase.New().Post(path).BodyForm(tweetParams).Request()
```

Requests will include an `application/x-www-form-urlencoded` Content-Type header.

#### Plain Body

Use `Body` to set a plain `io.Reader` on requests created by a Meteor.

```go
body := strings.NewReader("raw body")
req, err := meteor.New().Base("https://example.com").Body(body).Request()
```

Set a content type header, if desired (e.g. `Set("Content-Type", "text/plain")`).

### Extend a Meteor Service

Each Meteor creates a standard `http.Request` (e.g. with some path and query
params) each time `Request()` is called. You may wish to extend an existing Meteor to minimize duplication (e.g. a common client or base url).

Each Meteor instance provides a `New()` method which creates an independent copy, so setting properties on the child won't mutate the parent Meteor.

```go
const sunV1API = "https://api.weather.com/v1/"
base := meteor.New().Base(sunV1API).Client(nil)

// Daily Forecast, https://api.weather.com/v1/forecast/hourly/3day.json
days := 3
dailyForecastMeteor := base.New().Pathf("forecast/daily/%vday.json", days).Get().QueryStruct(params)
req, err := dailyForecastMeteor.Request()

// Hourly Forecast, https://api.weather.com/v1/forecast/hourly/6hour.json
hours := 6
hourlyForecastMeteor := base.New().Pathf("forecast/hourly/%vhour.json", hours)
req, err := hourlyForecastMeteor.Request()
```

Without the calls to `base.New()`, `dailyForecastMeteor` and `hourlyForecastMeteor` would reference the base Meteor and make the second request to
"https://api.weather.com/forecast/daily/forecast/hourly/6hour.json", which
is undesired.

Recap: If you wish to *extend* a Meteor service, create a new child copy with `New()`.

### Receiving

#### Receive

Define a JSON struct to decode a type from 2XX success responses. Use `ReceiveSuccess(successV interface{})` to send a new Request and decode the response body into `successV` if it succeeds.

```go
// Github Issue (abbreviated)
type Issue struct {
    Title  string `json:"title"`
    Body   string `json:"body"`
}
```

```go
issues := new([]Issue)
resp, err := githubBase.New().Get(path).QueryStruct(params).ReceiveSuccess(issues)
fmt.Println(issues, resp, err)
```

Most APIs return failure responses with JSON error details. To decode these, define success and failure JSON structs. Use `Receive(successV, failureV interface{})` to send a new Request that will automatically decode the response into the `successV` for 2XX responses or into `failureV` for non-2XX responses.

```go
type GithubError struct {
    Message string `json:"message"`
    Errors  []struct {
        Resource string `json:"resource"`
        Field    string `json:"field"`
        Code     string `json:"code"`
    } `json:"errors"`
    DocumentationURL string `json:"documentation_url"`
}
```

```go
issues := new([]Issue)
githubError := new(GithubError)
resp, err := githubBase.New().Get(path).QueryStruct(params).Receive(issues, githubError)
fmt.Println(issues, githubError, resp, err)
```

Pass a nil `successV` or `failureV` argument to skip JSON decoding into that value.

### Sending

#### Do

Meteor has two Do methods: `Do` and `DoAsync`. `Do`/`DoAsync` sends an HTTP request and returns the response. After the receiving the response, it calls the appropriate Responder and return a raw Response and error.

`Do` can be called with or without a request parameter. For example, considering this setup:
```go
var sResp dailyforecast.DailyForecastResponse
v1Base := GetService().New().Base(sunV1API).Client(nil)

dailyForecastMeteor := v1Base.New().Pathf("geocode/%v/%v", "34.063", "-84.217").Pathf("forecast/daily/%vday.json", 3).Get().QueryStruct(&Params{
		Key: meteorService.GetCredBy("sun"),
		Language: "en-US",
		Units: "e",
	}).Responder(meteor.JSONSuccessResponder(&sRespe))
```

You could call `Do` with `Request()` (if you would like to further modify the request before doing the API or do some additional error handling):
```go
req, err := dailyForecastMeteor.Request()
if err != nil {
	// Do something
}
dailyForecastMeteor.Do(req)
```

Or, solo:
```go
dailyForecastMeteor.Do()
```

##### DoAsync

Additionally, you could call various APIs using `DoAsync`. For example, say you have an API that returns various IDs for an additional API. For example, you could do something like this:

```go
import "github.ibm.com/TheWeatherCompany/cumulus"

```


### Modify a Request

Meteor provides the raw http.Request so modifications can be made using standard net/http features. For example, in Go 1.7+ , add HTTP tracing to a request with a context:

```go
req, err := meteor.New().Get("https://example.com").QueryStruct(params).Request()
// handle error

trace := &httptrace.ClientTrace{
   DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
      fmt.Printf("DNS Info: %+v\n", dnsInfo)
   },
   GotConn: func(connInfo httptrace.GotConnInfo) {
      fmt.Printf("Got Conn: %+v\n", connInfo)
   },
}

req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
client.Do(req)
```

### Build an API

APIs typically define an endpoint (also called a service) for each type of resource. For example, here is a tiny Github IssueService which [lists](https://developer.github.com/v3/issues/#list-issues-for-a-repository) repository issues.

```go
const baseURL = "https://api.github.com/"

type IssueService struct {
    meteor *meteor.Meteor
}

func NewIssueService(httpClient *http.Client) *IssueService {
    return &IssueService{
        meteor: meteor.New().Client(httpClient).Base(baseURL),
    }
}

func (s *IssueService) ListByRepo(owner, repo string, params *IssueListParams) ([]Issue, *http.Response, error) {
    issues := new([]Issue)
    githubError := new(GithubError)
    resp, err := s.meteor.New().Getf("repos/%s/%s/issues", owner, repo).QueryStruct(params).Receive(issues, githubError)
    if err == nil {
        err = githubError
    }
    return *issues, resp, err
}
```

## Documentation

Coming soon to an IBM place near you... 

## TODO

* TODO Add TravisCI Build Status.
* TODO Add GoLang Docs. Source code for [GoDoc](https://github.com/golang/gddo).
* TODO Add a logo here.

## Motivation

Many client libraries follow the lead of [google/go-github](https://github.com/google/go-github) (our inspiration!), but do so by reimplementing logic common to all clients.

[Meteor](https://github.com/dghubble/meteor) borrowed and abstracted those ideas into an agnostic component any API client can use for creating and sending requests. Due to massive refactoring and additional requirements, we decided to fork [Sling](https://github.com/dghubble/sling) and create a new custom Go HTTP client library for The Weather Company, an IBM Business.

## Contributing

See the [Contributing Guide](CONTRIBUTING.md).

## License

[MIT License](LICENSE)