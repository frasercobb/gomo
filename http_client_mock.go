package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

type MockHTTPClient struct {
	returnResponse *http.Response
	returnError    error
	calls          []*http.Request
}

func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		returnResponse: nil,
		returnError:    fmt.Errorf("Client.Do was called unexpectedly"),
		calls:          []*http.Request{},
	}
}

func (c *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.calls = append(c.calls, req)
	return c.returnResponse, c.returnError
}

func (c *MockHTTPClient) GetCalls() []*http.Request {
	return c.calls
}

func (c *MockHTTPClient) GivenErrorIsReturned(err error) {
	c.returnResponse = nil
	c.returnError = err
}

func (c *MockHTTPClient) GivenResponseIsReturned(statusCode int, body string, header http.Header) {
	bodyContent := ioutil.NopCloser(bytes.NewReader([]byte(body)))
	response := &http.Response{Body: bodyContent, StatusCode: statusCode, Header: header}
	c.returnResponse = response
	c.returnError = nil
}
