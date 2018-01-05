package meteor

import (
	"net"
	"net/http"
	"reflect"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

var (
	creds        = map[string]string{"sun": "abc", "dsx": "123"}
	credentials  = Credentials{"sun": "abc", "dsx": "123"}
	customClient = &http.Client{
		Timeout: time.Minute * 5,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}
	defaultClient = http.DefaultClient
)

func TestNewCredentials(t *testing.T) {
	type args struct {
		creds map[string]string
	}
	tests := []struct {
		name string
		args args
		want Credentials
	}{
		{"creds", args{creds}, credentials},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCredentials(tt.args.creds); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMeteor_GetCredBy(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		c    *Meteor
		args args
		want string
	}{
		{"getValidCred", NewMeteor(creds, nil), args{"sun"}, "abc"},
		{"getInvalidCred", NewMeteor(creds, nil), args{"sundsx"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.GetCredBy(tt.args.key); got != tt.want {
				t.Errorf("Meteor.GetCredBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMeteor_GetHTTPClient(t *testing.T) {
	c := customClient
	tests := []struct {
		name string
		c    *Meteor
		want *http.Client
	}{
		{"getDefaultClient", NewMeteor(creds), defaultClient},
		{"getDefaultClientWithNilClient", NewMeteor(creds, nil), defaultClient},
		{"customClient", NewMeteor(creds, c), c},
		{"getDefaultClient", NewMeteor(credentials), defaultClient},
		{"getDefaultClientWithNilClient", NewMeteor(credentials, nil), defaultClient},
		{"customClient", NewMeteor(credentials, c), c},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.GetHTTPClient(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Meteor.GetHTTPClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMeteor(t *testing.T) {
	type args struct {
		credentials Credentials
		httpClient  []*http.Client
	}
	tests := []struct {
		name string
		args args
		want *Meteor
	}{
		// Uses default client
		{"new", args{credentials: credentials}, &Meteor{
			httpClient:  http.DefaultClient,
			credentials: credentials,
			UserAgent:   UserAgent,
			Common:      &Service{
				httpClient:   http.DefaultClient,
				method:       "GET",
				header:       make(http.Header),
				queryStructs: make([]interface{}, 0),
				responder:    GenericResponder(),
			},
		}},

		// uses whatever custom client
		{"newWithClient", args{credentials, []*http.Client{customClient}}, &Meteor{
			httpClient:  customClient,
			credentials: credentials,
			UserAgent:   UserAgent,
			Common:      &Service{
				httpClient:   customClient,
				method:       "GET",
				header:       make(http.Header),
				queryStructs: make([]interface{}, 0),
				responder:    GenericResponder(),
			},
		}},

		// uses only the first client given since multiple clients are not supported
		{"newWithClients", args{credentials, []*http.Client{customClient, http.DefaultClient}}, &Meteor{
			httpClient:  customClient,
			credentials: credentials,
			UserAgent:   UserAgent,
			Common:      &Service{
				httpClient:   customClient,
				method:       "GET",
				header:       make(http.Header),
				queryStructs: make([]interface{}, 0),
				responder:    GenericResponder(),
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMeteor(tt.args.credentials, tt.args.httpClient...); !assert.Equal(t, tt.want, got) {
				t.Errorf("%s NewMeteor() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
