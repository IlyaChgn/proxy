package network

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

type HTTPRequest struct {
	ID         string         `json:"id"`
	Proto      string         `json:"proto"`
	Method     string         `json:"method"`
	Scheme     string         `json:"scheme"`
	Host       string         `json:"host"`
	Port       string         `json:"port"`
	Path       string         `json:"path"`
	Header     http.Header    `json:"headers"`
	PostParams url.Values     `json:"post_params"`
	GetParams  url.Values     `json:"get_params"`
	Cookies    []*http.Cookie `json:"cookies"`
	Body       []byte         `json:"body"`
}

func NewHTTPRequest(req *http.Request) *HTTPRequest {
	parsedRequest := &HTTPRequest{}

	parsedRequest.Proto = req.Proto
	parsedRequest.Method = req.Method

	colonIndex := strings.Index(req.Host, ":")
	if colonIndex == -1 {
		parsedRequest.Host = req.Host
		parsedRequest.Port = "80"
	} else {
		parsedRequest.Host = req.Host[:colonIndex]
		parsedRequest.Port = req.Host[colonIndex+1:]
	}

	parsedRequest.Scheme = req.URL.Scheme
	if parsedRequest.Scheme == "" {
		parsedRequest.Scheme = "http"

		if parsedRequest.Port == "443" {
			parsedRequest.Scheme = "https"
		}
	}

	parsedRequest.Path = req.URL.Path

	parsedRequest.Header = req.Header
	parsedRequest.Header.Del("Proxy-Connection")

	req.ParseForm()
	parsedRequest.PostParams = req.PostForm

	parsedRequest.GetParams = req.URL.Query()
	parsedRequest.Cookies = req.Cookies()

	var body []byte

	if req.ContentLength == -1 {
		body, _ = io.ReadAll(req.Body)
	} else {
		body = make([]byte, req.ContentLength)
		req.Body.Read(body)
	}

	parsedRequest.Body = body

	return parsedRequest
}
