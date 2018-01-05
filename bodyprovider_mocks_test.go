package meteor

import (
	"io"
	"sync"
)


var (
	lockBodyProviderMockBody        sync.RWMutex
	lockBodyProviderMockContentType sync.RWMutex
)

type BodyProviderMock struct {
	payload interface{}

	// BodyFunc mocks the Body method.
	BodyFunc func() (io.Reader, error)

	// ContentTypeFunc mocks the ContentType method.
	ContentTypeFunc func() string

	// calls tracks calls to the methods.
	calls struct {
		// Body holds details about calls to the Body method.
		Body []struct {
		}
		// ContentType holds details about calls to the ContentType method.
		ContentType []struct {
		}
	}
}

// Body calls BodyFunc.
func (mock *BodyProviderMock) Body() (io.Reader, error) {
	if mock.BodyFunc == nil {
		panic("moq: BodyProviderMock.BodyFunc is nil but BodyProvider.Body was just called")
	}
	callInfo := struct {
	}{}
	lockBodyProviderMockBody.Lock()
	mock.calls.Body = append(mock.calls.Body, callInfo)
	lockBodyProviderMockBody.Unlock()
	return mock.BodyFunc()
}

// BodyCalls gets all the calls that were made to Body.
// Check the length with:
//     len(mockedBodyProvider.BodyCalls())
func (mock *BodyProviderMock) BodyCalls() []struct {
} {
	var calls []struct {
	}
	lockBodyProviderMockBody.RLock()
	calls = mock.calls.Body
	lockBodyProviderMockBody.RUnlock()
	return calls
}

// ContentType calls ContentTypeFunc.
func (mock *BodyProviderMock) ContentType() string {
	if mock.ContentTypeFunc == nil {
		panic("moq: BodyProviderMock.ContentTypeFunc is nil but BodyProvider.ContentType was just called")
	}
	callInfo := struct {
	}{}
	lockBodyProviderMockContentType.Lock()
	mock.calls.ContentType = append(mock.calls.ContentType, callInfo)
	lockBodyProviderMockContentType.Unlock()
	return mock.ContentTypeFunc()
}

// ContentTypeCalls gets all the calls that were made to ContentType.
// Check the length with:
//     len(mockedBodyProvider.ContentTypeCalls())
func (mock *BodyProviderMock) ContentTypeCalls() []struct {
} {
	var calls []struct {
	}
	lockBodyProviderMockContentType.RLock()
	calls = mock.calls.ContentType
	lockBodyProviderMockContentType.RUnlock()
	return calls
}
