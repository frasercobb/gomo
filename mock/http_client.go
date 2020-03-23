package mock

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

type HTTPClient struct {
	returnResponse *http.Response
	returnError    error
	calls          []*http.Request
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		returnResponse: nil,
		returnError:    fmt.Errorf("Client.Do was called unexpectedly"),
		calls:          []*http.Request{},
	}
}

func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.calls = append(c.calls, req)
	return c.returnResponse, c.returnError
}

func (c *HTTPClient) GetCalls() []*http.Request {
	return c.calls
}

func (c *HTTPClient) GivenErrorIsReturned(err error) {
	c.returnResponse = nil
	c.returnError = err
}

func (c *HTTPClient) GivenResponseIsReturned(statusCode int, body string, header http.Header) {
	bodyContent := ioutil.NopCloser(bytes.NewReader([]byte(body)))
	response := &http.Response{Body: bodyContent, StatusCode: statusCode, Header: header}
	c.returnResponse = response
	c.returnError = nil
}
