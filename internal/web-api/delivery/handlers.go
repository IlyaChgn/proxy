package delivery

import (
	"github.com/gorilla/mux"
	"net/http"
	"proxy/internal/network"
	"proxy/internal/proxy/delivery"
	"proxy/internal/web-api/usecases"
)

type Handler struct {
	storage usecases.WebApiInterface
	proxy   *delivery.ProxyHandler
}

func NewHandler(storage usecases.WebApiInterface, proxy *delivery.ProxyHandler) *Handler {
	return &Handler{
		storage: storage,
		proxy:   proxy,
	}
}

func (h *Handler) GetRequestsList(writer http.ResponseWriter, request *http.Request) {
	reqs, err := h.storage.GetAllRequests()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	SendOkResponse(writer, reqs)
}

func (h *Handler) GetRequest(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	req, err := h.storage.GetRequest(vars["id"])
	if err != nil || req == nil {
		http.Error(writer, "request not found", http.StatusInternalServerError)

		return
	}

	SendOkResponse(writer, req)
}

func (h *Handler) RepeatRequest(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	req, err := h.storage.GetRequest(vars["id"])
	if err != nil || req == nil {
		http.Error(writer, "request not found", http.StatusInternalServerError)

		return
	}

	resp, err := h.proxy.HandleHTTP(writer, req)
	if err != nil {
		http.Error(writer, "something went wrong while repeating request", http.StatusInternalServerError)

		return
	}

	parsedResp := network.NewHTTPResponse(resp)
	h.proxy.SendNewResponse(writer, parsedResp)
}

func (h *Handler) ScanRequest(writer http.ResponseWriter, request *http.Request) {}
