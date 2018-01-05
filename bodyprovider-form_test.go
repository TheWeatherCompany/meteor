package meteor

import (
	"io"
	"reflect"
	"testing"
	"strings"
	goquery "github.com/google/go-querystring/query"
)

func Test_formBodyProvider_ContentType(t *testing.T) {
	type StatusUpdateParams struct {
		Status            string  `url:"status,omitempty"`
		InReplyToStatusId int64   `url:"in_reply_to_status_id,omitempty"`
		MediaIds          []int64 `url:"media_ids,omitempty,comma"`
	}

	tests := []struct {
		name    string
		p       formBodyProvider
		want    string
	}{
		{"form", formBodyProvider{&StatusUpdateParams{Status: "writing some Go"}}, formContentType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.ContentType(); got != tt.want {
				t.Errorf("formBodyProvider.ContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_formBodyProvider_Body(t *testing.T) {
	type StatusUpdateParams struct {
		Status            string  `url:"status,omitempty"`
		InReplyToStatusId int64   `url:"in_reply_to_status_id,omitempty"`
		MediaIds          []int64 `url:"media_ids,omitempty,comma"`
	}
	values, _ := goquery.Values(&StatusUpdateParams{Status: "writing some Go"})
	formBody := strings.NewReader(values.Encode())

	tests := []struct {
		name    string
		p       formBodyProvider
		want    io.Reader
		wantErr bool
	}{
		{"form", formBodyProvider{&StatusUpdateParams{Status: "writing some Go"}}, formBody, false},
		{"formErr", formBodyProvider{map[string]string{"status": "writing some Go"}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.Body()
			if (err != nil) != tt.wantErr {
				t.Errorf("formBodyProvider.Body() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("formBodyProvider.Body() = %v, want %v", got, tt.want)
			}
		})
	}
}
