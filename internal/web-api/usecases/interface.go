package usecases

import "proxy/internal/network"

type WebApiInterface interface {
	SaveRequest(request *network.HTTPRequest) (string, error)
	SaveResponse(response *network.HTTPResponse, id string) error
	GetRequest(id string) (*network.HTTPRequest, error)
	GetResponse(id string) (*network.HTTPResponse, error)
	GetAllRequests() ([]*network.HTTPRequest, error)
}
