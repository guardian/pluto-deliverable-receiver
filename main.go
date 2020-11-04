package main

import (
	"github.com/go-redis/redis/v7"
	"gitlab.com/codmill/customer-projects/guardian/deliverable-receiver/helpers"
	"log"
	"net/http"
)

type MyHttpApp struct {
	index IndexHandler
	healthcheck HealthcheckHandler
	initiate InitiateHandler
	uploadFile UploadHandler
}

func SetupRedis(config *helpers.Config) (*redis.Client, error) {
	log.Printf("Connecting to Redis on %s", config.Redis.Address)
	client := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Address,
		Password: config.Redis.Password,
		DB:       config.Redis.DBNum,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Printf("Could not contact Redis: %s", err)
		return nil, err
	}
	log.Printf("Done.")
	return client, nil
}

func main() {
	var app MyHttpApp
	/*
		read in config and establish connection to persistence layer
	*/
	log.Printf("Reading config from serverconfig.yaml")
	config, configReadErr := helpers.ReadConfig("config/serverconfig.yaml")
	log.Print("Done.")

	if configReadErr != nil {
		log.Fatal("No configuration, can't continue")
	}

	redisClient, redisErr := SetupRedis(config)
	if redisErr != nil {
		log.Fatal("Could not connect to redis")
	}

	app.healthcheck.redisClient = redisClient
	app.uploadFile.redisClient = redisClient
	app.initiate.redisClient = redisClient
	app.initiate.config = config

	http.Handle("/", app.index)
	http.Handle("/healthcheck", app.healthcheck)
	http.Handle("/initiate", app.initiate)
	http.Handle("/upload", app.uploadFile)

	log.Printf("Starting server on port 9000")
	startServerErr := http.ListenAndServe(":9000", nil)

	if startServerErr != nil {
		log.Fatal(startServerErr)
	}
}
