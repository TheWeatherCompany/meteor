package meteor

import (
	"io"
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
	return p.body, nil
}
