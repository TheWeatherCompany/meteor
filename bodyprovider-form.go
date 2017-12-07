package meteor

import (
	"io"
	"strings"

	goquery "github.com/google/go-querystring/query"
)

// formBodyProvider encodes a url tagged struct value as Body for requests.
// See https://godoc.org/github.com/google/go-querystring/query for details.
type formBodyProvider struct {
	payload interface{}
}

func (p formBodyProvider) ContentType() string {
	return formContentType
}

func (p formBodyProvider) Body() (io.Reader, error) {
	values, err := goquery.Values(p.payload)
	if err != nil {
		return nil, err
	}
	return strings.NewReader(values.Encode()), nil
}