package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"proxy/internal/proxy/delivery"
	webapidelivery "proxy/internal/web-api/delivery"
	"proxy/internal/web-api/repository"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

const (
	certPath = "server.crt"
	keyPath  = "server.key"
	Port     = "8000"
)

type Server struct {
	server *http.Server
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading env file", err)

		return
	}

	redisClient := redis.NewClient(
		&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       0,
		},
	)

	storage := repository.NewStorage(redisClient)

	proxyHandler := delivery.NewProxyHandler(certPath, keyPath, storage)
	handler := webapidelivery.NewHandler(storage, proxyHandler)

	router := mux.NewRouter()
	rootRouter := router.PathPrefix("/api").Subrouter()

	rootRouter.HandleFunc("/requests", handler.GetRequestsList)
	rootRouter.HandleFunc("/requests/{id}", handler.GetRequest)
	rootRouter.HandleFunc("/repeat/{id}", handler.RepeatRequest)
	rootRouter.HandleFunc("/scan/{id}", handler.ScanRequest)

	srv := new(Server)

	srv.server = &http.Server{
		Addr:    ":8000",
		Handler: router,
	}

	log.Println("Web-API is running on port", 8000)

	go func() {
		err := srv.server.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Proxy server is running on port 8080")

	err = http.ListenAndServe(":8080", proxyHandler)
	if err != nil {
		log.Fatal(err)
	}
}
