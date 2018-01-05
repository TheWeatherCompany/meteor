package meteor

import (
	"io"
	"reflect"
	"testing"
	"bytes"
	"encoding/json"
)

func Test_jsonBodyProvider_ContentType(t *testing.T) {
	type IssueRequest struct {
		Title     string   `json:"title,omitempty"`
		Body      string   `json:"body,omitempty"`
		Assignee  string   `json:"assignee,omitempty"`
		Milestone int      `json:"milestone,omitempty"`
		Labels    []string `json:"labels,omitempty"`
	}
	payload := &IssueRequest{
		Title: "Test title",
		Body:  "Some issue",
	}

	tests := []struct {
		name string
		p    jsonBodyProvider
		want string
	}{
		{"json", jsonBodyProvider{payload}, jsonContentType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.ContentType(); got != tt.want {
				t.Errorf("jsonBodyProvider.ContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jsonBodyProvider_Body(t *testing.T) {
	type IssueRequest struct {
		Title     string   `json:"title,omitempty"`
		Body      string   `json:"body,omitempty"`
		Assignee  string   `json:"assignee,omitempty"`
		Milestone int      `json:"milestone,omitempty"`
		Labels    []string `json:"labels,omitempty"`
	}
	payload := &IssueRequest{
		Title: "Test title",
		Body:  "Some issue",
	}
	jsonBody := &bytes.Buffer{}
	json.NewEncoder(jsonBody).Encode(payload)


	tests := []struct {
		name    string
		p       jsonBodyProvider
		want    io.Reader
		wantErr bool
	}{
		{"json", jsonBodyProvider{payload}, jsonBody, false},
		{"error", jsonBodyProvider{func(){}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.Body()
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonBodyProvider.Body() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("jsonBodyProvider.Body() = %v, want %v", got, tt.want)
			}
		})
	}
}
