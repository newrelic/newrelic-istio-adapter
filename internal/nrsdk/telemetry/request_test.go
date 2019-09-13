package telemetry

import (
	"errors"
	"testing"
)

type testRequestBuilder struct {
	requests []func() (request, error)
}

func (ts testRequestBuilder) newRequest(licenseKey, urlOverride string) (request, error) {
	return ts.requests[0]()
}

func (ts testRequestBuilder) split() []requestsBuilder {
	reqs := ts.requests[1:]
	if len(reqs) == 0 {
		return nil
	}
	return []requestsBuilder{
		testRequestBuilder{requests: reqs},
		testRequestBuilder{requests: reqs},
	}
}

func TestNewRequestsSplitSuccess(t *testing.T) {
	ts := testRequestBuilder{
		requests: []func() (request, error){
			func() (request, error) { return request{compressedBodyLength: 20}, nil },
			func() (request, error) { return request{compressedBodyLength: 15}, nil },
			func() (request, error) { return request{compressedBodyLength: 11}, nil },
			func() (request, error) { return request{compressedBodyLength: 9}, nil },
		},
	}
	reqs, err := newRequests(ts, "", "", 10)
	if err != nil {
		t.Error(err)
	}
	if len(reqs) != 8 {
		t.Error(len(reqs))
	}
}

func TestNewRequestsCantMakeRequest(t *testing.T) {
	reqError := errors.New("cant make request")
	ts := testRequestBuilder{
		requests: []func() (request, error){
			func() (request, error) { return request{compressedBodyLength: 20}, nil },
			func() (request, error) { return request{compressedBodyLength: 15}, nil },
			func() (request, error) { return request{compressedBodyLength: 11}, nil },
			func() (request, error) { return request{}, reqError },
		},
	}
	reqs, err := newRequests(ts, "", "", 10)
	if err != reqError {
		t.Error(err)
	}
	if len(reqs) != 0 {
		t.Error(len(reqs))
	}
}

func TestNewRequestsCantSplit(t *testing.T) {
	ts := testRequestBuilder{
		requests: []func() (request, error){
			func() (request, error) { return request{compressedBodyLength: 20}, nil },
			func() (request, error) { return request{compressedBodyLength: 15}, nil },
			func() (request, error) { return request{compressedBodyLength: 11}, nil },
		},
	}
	reqs, err := newRequests(ts, "", "", 10)
	if err != errUnableToSplit {
		t.Error(err)
	}
	if len(reqs) != 0 {
		t.Error(len(reqs))
	}
}
