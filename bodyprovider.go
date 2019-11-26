package meteor

//go:generate moq -out bodyprovider_mocks_test.go . BodyProvider

import (
	"errors"
	"io"
)

const (
	jpegContentType = "image/jpeg"
	pngContentType  = "image/png"
	gifContentType  = "image/gif"
	textContentType = "text/plain"
	jsonContentType = "application/json"
	formContentType = "application/x-www-form-urlencoded"
)

// BodyProvider provides Body content for http.Request attachment.
type BodyProvider interface {
	// ContentType returns the Content-Type of the body.
	ContentType() string
	// Body returns the io.Reader body.
	Body() (io.Reader, error)
}

// bodyProvider provides the wrapped body value as a Body for requests.
type bodyProvider struct {
	body io.Reader
}

// ContentType gets the content type of the body.
// Implements BodyProvider interface
func (p bodyProvider) ContentType() string {
	return ""
}

// Body returns the body of the provider
// Implements BodyProvider interface
func (p bodyProvider) Body() (io.Reader, error) {
	if p.body == nil {
		return nil, errors.New("no body reader")
	}
	return p.body, nil
}
