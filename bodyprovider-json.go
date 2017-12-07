package meteor

import (
	"bytes"
	"encoding/json"
	"io"
)

// jsonBodyProvider encodes a JSON tagged struct value as a Body for requests.
// See https://golang.org/pkg/encoding/json/#MarshalIndent for details.
type jsonBodyProvider struct {
	payload interface{}
}

func (p jsonBodyProvider) ContentType() string {
	return jsonContentType
}

func (p jsonBodyProvider) Body() (io.Reader, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(p.payload)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
