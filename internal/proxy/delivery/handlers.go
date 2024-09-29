package delivery

import (
	"crypto/tls"
	"log"
	"net/http"
	"proxy/internal/network"
	"proxy/internal/proxy/certs"
	"proxy/internal/web-api/usecases"
)

type ProxyHandler struct {
	ca      *tls.Certificate
	storage usecases.WebApiInterface
}

func NewProxyHandler(certPath, keyPath string, storage usecases.WebApiInterface) *ProxyHandler {
	proxy := &ProxyHandler{
		storage: storage,
	}

	var err error

	proxy.ca, err = certs.LoadCertificate(certPath, keyPath)
	if err != nil {
		log.Println("Something went wrong while getting CA", err)

		return nil
	}

	return proxy
}

func (proxy *ProxyHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	parsedReq := network.NewHTTPRequest(req)

	var (
		resp *http.Response
		err  error
	)

	if parsedReq.Method == http.MethodConnect {
		err = proxy.handleConnect(writer, parsedReq)

		return
	}

	resp, err = proxy.HandleHTTP(writer, parsedReq)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	id, err := proxy.storage.SaveRequest(parsedReq)
	if err != nil {
		log.Println("Something went wrong while saving request", err)

		return
	}

	parsedResp := network.NewHTTPResponse(resp)

	err = proxy.storage.SaveResponse(parsedResp, id)
	if err != nil {
		log.Println("Something went wrong while saving response", err)

		return
	}

	proxy.SendNewResponse(writer, parsedResp)
}
