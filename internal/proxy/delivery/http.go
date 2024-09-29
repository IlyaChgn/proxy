package delivery

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"proxy/internal/network"
	"proxy/internal/proxy/certs"
	"strings"
)

func (proxy *ProxyHandler) HandleHTTP(writer http.ResponseWriter, req *network.HTTPRequest) (*http.Response, error) {
	var reader io.Reader

	if len(req.PostParams) > 0 {
		reader = strings.NewReader(req.PostParams.Encode())
	} else {
		reader = bytes.NewReader(req.Body)
	}

	newReq, err := http.NewRequest(req.Method,
		fmt.Sprintf("%s://%s:%s%s", req.Scheme, req.Host, req.Port, req.Path), reader)
	newReq.Header = req.Header

	for _, cookie := range req.Cookies {
		newReq.AddCookie(cookie)
	}

	for key, value := range req.GetParams {
		for _, v := range value {
			newReq.URL.Query().Add(key, v)
		}
	}

	var config *certs.TLSConfig
	if req.Scheme == "https" {
		config, err = certs.GetTLSConfig(req, proxy.ca)
		if err != nil {
			log.Println("Error getting TLS config:", err)

			return nil, err
		}
	}

	var tlsCfg *tls.Config
	if config != nil {
		tlsCfg = config.Cfg
	}

	transport := &http.Transport{
		TLSClientConfig: tlsCfg,
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: transport,
	}

	resp, err := client.Do(newReq)
	if err != nil {
		http.Error(writer, "Failed to send request", http.StatusInternalServerError)
		log.Println("Something went wrong while sending request:", err)

		return nil, err
	}

	return resp, nil
}

func (proxy *ProxyHandler) SendNewResponse(writer http.ResponseWriter, resp *network.HTTPResponse) {
	network.CopyHeaders(writer.Header(), resp.Headers)
	writer.WriteHeader(resp.Code)

	writer.Write([]byte(fmt.Sprintf("%s %s\n", resp.Proto, resp.Message)))

	for k, v := range resp.Headers {
		writer.Write([]byte(fmt.Sprintf("%s: %s\n", k, strings.Join(v, ", "))))
	}

	writer.Write([]byte("\n"))

	reader := bytes.NewReader(resp.Body)

	io.Copy(writer, reader)
}
