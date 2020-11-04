package main

import (
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v7"
	"gitlab.com/codmill/customer-projects/guardian/deliverable-receiver/helpers"
	"gitlab.com/codmill/customer-projects/guardian/deliverable-receiver/models"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type InitiateHandler struct {
	redisClient *redis.Client
	config *helpers.Config
}

type InitiateRequest struct {
	ProjectId int `json:"project_id"`
	DropFolderPath string `json:"drop_folder"`
}

func makeRelativePath(inputPath string, config *helpers.Config) (string, error) {
	if ! strings.HasPrefix(inputPath, config.StoragePrefix.LocalPath) {
		return "", errors.New("invalid destination path, please check the settings")
	}

	return inputPath[len(config.StoragePrefix.LocalPath):], nil
}

func (h InitiateHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	bodyContent, readErr := ioutil.ReadAll(request.Body)

	if readErr != nil {
		log.Printf("ERROR InitiateHandler could not read request from client: %s", readErr)
		return
	}

	if ! helpers.AssertHttpMethod(request, w, "POST") {
		return	//error has already been written
	}

	var requestContent InitiateRequest
	unmarshalErr := json.Unmarshal(bodyContent, &requestContent)

	if unmarshalErr != nil {
		log.Printf("ERROR InitiateHandler could not parse request body: %s", unmarshalErr)
		response := helpers.GenericErrorResponse{
			Status: "error",
			Detail: "could not parse request, see logs",
		}
		helpers.WriteJsonContent(response, w, 400)
		return
	}

	relPath, relPathErr := makeRelativePath(requestContent.DropFolderPath, h.config)
	if relPathErr != nil {
		response := helpers.GenericErrorResponse{
			Status: "error",
			Detail: "invalid dropfolder path",
		}
		helpers.WriteJsonContent(response, w, 400)
		return
	}

	ttl, configErr := h.config.UploadSlotTTLDuration()
	if configErr != nil {
		log.Print("ERROR InitiateHandler could not determine upload slot TTL: ", configErr)
		response := helpers.GenericErrorResponse{
			Status: "config_error",
			Detail: "could not determine upload slot TTL",
		}
		helpers.WriteJsonContent(response, w, 500)
		return
	}

	newSlot, newSlotErr := models.NewUploadSlot(requestContent.ProjectId, relPath, ttl)

}