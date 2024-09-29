package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"proxy/internal/network"
	"strconv"
)

type Storage struct {
	client *redis.Client
}

func NewStorage(client *redis.Client) *Storage {
	client.SetNX(context.Background(), "next_key", 0, 0)
	return &Storage{client: client}
}

func (storage *Storage) nextKey() (string, error) {
	id, err := storage.client.Incr(context.Background(), "next_key").Result()
	if err != nil {
		log.Println("storage nextKey error:", err)

		return "", err
	}

	idStr := strconv.Itoa(int(id))

	return idStr, nil
}

func (storage *Storage) SaveRequest(request *network.HTTPRequest) (string, error) {
	var err error

	request.ID, err = storage.nextKey()
	if err != nil {
		return "", err
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Println("error serializing request ", err)

		return "", err
	}

	err = storage.client.Set(context.Background(), fmt.Sprintf("request_%s", request.ID), jsonData, 0).
		Err()
	if err != nil {
		log.Println("error saving request", err)

		return "", err
	}

	return request.ID, nil
}

func (storage *Storage) SaveResponse(response *network.HTTPResponse, id string) error {
	response.ID = id

	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Println("error serializing response ", err)

		return err
	}

	err = storage.client.Set(context.Background(), fmt.Sprintf("response_%s", id), jsonData, 0).Err()
	if err != nil {
		log.Println("error saving response", err)

		return err
	}

	return nil
}

func (storage *Storage) GetRequest(id string) (*network.HTTPRequest, error) {
	request, _ := storage.client.Get(context.Background(), fmt.Sprintf("request_%s", id)).Bytes()

	var parsedReq network.HTTPRequest

	err := json.Unmarshal(request, &parsedReq)
	if err != nil {
		log.Println("error deserializing request ", err)

		return nil, err
	}

	return &parsedReq, nil
}

func (storage *Storage) GetResponse(id string) (*network.HTTPResponse, error) {
	response, _ := storage.client.Get(context.Background(), fmt.Sprintf("response_%s", id)).Bytes()

	var parsedResp network.HTTPResponse

	err := json.Unmarshal(response, &parsedResp)
	if err != nil {
		log.Println("error deserializing response ", err)

		return nil, err
	}

	return &parsedResp, nil
}

func (storage *Storage) GetAllRequests() ([]*network.HTTPRequest, error) {
	lastID, _ := storage.client.Get(context.Background(), "next_key").Result()
	lastIDInt, _ := strconv.Atoi(lastID)

	requests := make([]*network.HTTPRequest, 0)

	for id := 1; id <= lastIDInt; id++ {
		request, _ := storage.client.Get(context.Background(), fmt.Sprintf("request_%s", strconv.Itoa(id))).
			Bytes()

		var parsedReq network.HTTPRequest

		err := json.Unmarshal(request, &parsedReq)
		if err != nil {
			log.Println("error deserializing request ", err)

			return nil, err
		}

		requests = append(requests, &parsedReq)
	}

	return requests, nil
}
