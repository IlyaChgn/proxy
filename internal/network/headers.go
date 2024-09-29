package network

import "net/http"

func CopyHeaders(dst, src http.Header) {
	for k, v := range src {
		if k == "Content-Length" {
			continue
		}
		dst[k] = v
	}
}
