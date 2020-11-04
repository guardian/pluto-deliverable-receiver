package main

import (
	"github.com/go-redis/redis/v7"
	"net/http"
)

type UploadHandler struct {
	redisClient *redis.Client
}

func (h UploadHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
}
