package meteor

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"io"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
	"github.com/jarcoal/httpmock"
)

func TestJSONSuccessResponder(t *testing.T) {
	type args struct {
		success interface{}
	}
	tests := []struct {
		name string
		args args
		want *jsonResponder
	}{
		{"nil", args{}, &jsonResponder{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JSONSuccessResponder(tt.args.success); !assert.Equal(t, tt.want, got) {
				t.Errorf("JSONSuccessResponder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONResponder(t *testing.T) {
	type args struct {
		success interface{}
		failure interface{}
	}
	tests := []struct {
		name string
		args args
		want *jsonResponder
	}{
		{"nil", args{}, &jsonResponder{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JSONResponder(tt.args.success, tt.args.failure); !assert.Equal(t, tt.want, got) {
				t.Errorf("JSONResponder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jsonResponder_Respond(t *testing.T) {
	var err error
	svc := New().Base(baseURL).Path("something")
	req, err := svc.Request()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	successBody := `{"id": 1234567890, "name": "Meteor Rocks!"}]`
	httpmock.RegisterResponder("GET", baseURL+"/success",
		httpmock.NewStringResponder(http.StatusOK, successBody))

	resp, err := svc.Do()

	type args struct {
		req  *http.Request
		resp *http.Response
		err  error
	}
	tests := []struct {
		name string
		r    *jsonResponder
		args args
		want Responder
	}{
		{"default", &jsonResponder{}, args{req, resp, err}, &jsonResponder{
			Request:  req,
			Response: resp,
			Error:    err,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Respond(tt.args.req, tt.args.resp, tt.args.err); !assert.Equal(t, tt.want, got) {
				t.Errorf("jsonResponder.Respond() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jsonResponder_DoResponse(t *testing.T) {
	svc := New().Base(baseURL).Path("something")
	req, _ := svc.Request()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	successBody := `{"id": 1234567890, "name": "Meteor Rocks!"}]`
	httpmock.RegisterResponder("GET", baseURL+"/something",
		httpmock.NewStringResponder(http.StatusOK, successBody))

	tests := []struct {
		name    string
		s       *Service
		req     *http.Request
		want    *http.Response
		wantErr bool
	}{
		{"default", svc.New(), req, &http.Response{Status: "200", StatusCode: http.StatusOK, Header: http.Header{}, Body: httpmock.NewRespBodyFromString(successBody)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.s.httpClient.Do(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonResponder.DoResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			defer func() {
				if resp.Header.Get("Accept-Ranges") != "bytes" {
					io.CopyN(ioutil.Discard, resp.Body, 512)
				}
				resp.Body.Close()
			}()

			r := tt.s.responder.Respond(req, resp, err)
			got, err := r.DoResponse()
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonResponder.DoResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, tt.want, got) {
				t.Errorf("jsonResponder.DoResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_jsonResponder_GetSuccess(t *testing.T) {
	svc := New().Base(baseURL)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	successBody := `{"a":"a success", "b": "another success", "c": "2017-11-01T22:08:41+00:00"}`
	successHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(successBody))
	}

	success204Handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte(``))
	}

	failBody := `{"errors": [{"code": "EAE:INV-0001","message": "Invalid request"}],"metadata": {"status_code": 400,"transaction_id": "1429140092945:1801695336"},"success": false}`
	failureHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(failBody))
	}

	tests := []struct {
		name string
		s *Service
		status int
		body string
		h func(w http.ResponseWriter, r *http.Request)
		want interface{}
	}{
		{"success", svc.New().Path("success"), http.StatusOK, successBody, successHandler, wantedSuccess},
		{"success204", svc.New().Path("success204"), http.StatusNoContent, ``, success204Handler, newSuccess()},
		{"fail", svc.New().Path("fail"), http.StatusBadRequest, failBody, failureHandler, newSuccess()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.RegisterResponder("GET", baseURL+"/"+tt.name,
				httpmock.NewStringResponder(tt.status, tt.body))

			req, _ := tt.s.JSONSuccessResponder(newSuccess()).Request()
			recorder := httptest.NewRecorder()
			tt.h(recorder,req)
			tt.s.responder.Respond(req, recorder.Result(), nil)
			tt.s.responder.DoResponse()

			tt.s.Do(req)
			resp := tt.s.GetResponder()
			if got := resp.GetSuccess(); !assert.Equal(t, tt.want, got) {
				t.Errorf("%v jsonResponder.GetSuccess() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_jsonResponder_GetFailure(t *testing.T) {
	svc := New().Base(baseURL)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	successBody := `{"a":"a success", "b": "another success", "c": "2017-11-01T22:08:41+00:00"}`
	successHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(successBody))
	}

	success204Handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte(``))
	}

	failBody := `{"errors": [{"code": "EAE:INV-0001","message": "Invalid request"}],"metadata": {"status_code": 400,"transaction_id": "1429140092945:1801695336"},"success": false}`
	failureHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(failBody))
	}
	tests := []struct {
		name string
		s *Service
		status int
		body string
		h func(w http.ResponseWriter, r *http.Request)
		want interface{}
	}{
		{"success", svc.New().Path("success"), http.StatusOK, successBody, successHandler, newFail()},
		{"success204", svc.New().Path("success204"), http.StatusNoContent, ``, success204Handler, newFail()},
		{"fail", svc.New().Path("fail"), http.StatusBadRequest, failBody, failureHandler, wantedFailure},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.RegisterResponder("GET", baseURL+"/"+tt.name,
				httpmock.NewStringResponder(tt.status, tt.body))

			req, _ := tt.s.JSONResponder(newSuccess(), newFail()).Request()
			recorder := httptest.NewRecorder()
			tt.h(recorder,req)
			tt.s.responder.Respond(req, recorder.Result(), nil)
			tt.s.responder.DoResponse()

			tt.s.Do(req)
			resp := tt.s.GetResponder()
			if got := resp.GetFailure(); !assert.Equal(t, tt.want, got) {
				t.Errorf("%v jsonResponder.GetFailure() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_decodeResponseJSON(t *testing.T) {
	type args struct {
		resp     *http.Response
		successV interface{}
		failureV interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := decodeResponseJSON(tt.args.resp, tt.args.successV, tt.args.failureV); (err != nil) != tt.wantErr {
				t.Errorf("decodeResponseJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_decodeResponseBodyJSON(t *testing.T) {
	type args struct {
		resp *http.Response
		v    interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := decodeResponseBodyJSON(tt.args.resp, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("decodeResponseBodyJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
