# Meteor

* TODO Add TravisCI Build Status
* TODO Add GoLang Docs
* TODO Add a logo here.

Meteor is a Go HTTP client library for creating and sending API requests.

Meteor stores HTTP Request properties to simplify sending requests and decoding responses. Check [usage](#usage) or the [examples](examples) to learn how to compose a Meteor into your API client.

### Features

* Method Setters: Get/Post/Put/Patch/Delete/Head
* Add or Set Request Headers
* Base/Path: Extend a Meteor for different endpoints
* Encode structs into URL query parameters
* Encode a form or JSON into the Request Body
* Receive JSON success or failure responses

## Install

    go get github.ibm.com/TheWeatherCompany/Meteor

## Documentation

Coming soon...

## Usage

Use a Meteor to set path, method, header, query, or body properties and create an `http.Request`.

```go
type Params struct {
    Count int `url:"count,omitempty"`
}

req, err := Meteor.New().Get("https://example.com").QueryStruct(&Params{Count: 5}).Request()
```

### Path

Use `Path` to set or extend the URL for created Requests. Extension means the path will be resolved relative to the existing URL.

```go
// creates a GET request to https://example.com/foo/bar
req, err := Meteor.New().Base("https://example.com/").Path("foo/").Path("bar").Request()
```

Use `Get`, `Post`, `Put`, `Patch`, `Delete`, or `Head` which are exactly the same as `Path` except they set the HTTP method too.

```go
req, err := Meteor.New().Post("http://upload.com/gophers")
```

### Headers

`Add` or `Set` headers for requests created by a Meteor.

```go
s := Meteor.New().Base(baseUrl).Set("User-Agent", "Gophergram API Client")
req, err := s.New().Get("gophergram/list").Request()
```

## Motivation

Many client libraries follow the lead of [google/go-github](https://github.com/google/go-github) (our inspiration!), but do so by reimplementing logic common to all clients.

[Sling](https://github.com/dghubble/sling) borrowed and abstracted those ideas into an agnostic component any API client can use for creating and sending requests. Due to massive refactoring and additional requirements, we decided to fork Sling and create a new custom Go HTTP client library for The Weather Company, an IBM Business.

## Contributing

See the [Contributing Guide](CONTRIBUTING.md).

## License

[MIT License](LICENSE)