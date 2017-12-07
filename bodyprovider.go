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

func (p bodyProvider) ContentType() string {
	return ""
}

func (p bodyProvider) Body() (io.Reader, error) {
	return p.body, nil
}
