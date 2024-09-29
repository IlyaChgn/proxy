package network

import (
	"compress/gzip"
	"io"
	"net/http"
)

type HTTPResponse struct {
	ID         string      `json:"id"`
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Proto      string      `json:"proto"`
	Headers    http.Header `json:"headers"`
	Body       []byte      `json:"-"`
	StringBody string      `json:"body"`
}

func NewHTTPResponse(resp *http.Response) *HTTPResponse {
	var reader io.ReadCloser

	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, _ = gzip.NewReader(resp.Body)
	} else {
		reader = resp.Body
	}

	body, _ := io.ReadAll(reader)

	defer reader.Close()

	parsedResponse := &HTTPResponse{
		Code:       resp.StatusCode,
		Message:    resp.Status,
		Proto:      resp.Proto,
		Headers:    resp.Header,
		Body:       body,
		StringBody: string(body),
	}

	return parsedResponse
}
