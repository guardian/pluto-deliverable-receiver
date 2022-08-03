package main

import (
	"github.com/go-redis/redis/v8"
	"gitlab.com/codmill/customer-projects/guardian/deliverable-receiver/helpers"
	"log"
	"net/http"
	"context"
)

type HealthcheckHandler struct {
	redisClient *redis.Client
}

func (h HealthcheckHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	_, err := h.redisClient.Ping(context.TODO()).Result()

	if err == nil {
		w.WriteHeader(200)
	} else {
		log.Printf("HEALTHCHECK FAILED: %s connecting to Redis", err)
		response := helpers.GenericErrorResponse{
			Status: "error",
			Detail: "could not contact redis db",
		}
		helpers.WriteJsonContent(response, w, 500)
	}
}
