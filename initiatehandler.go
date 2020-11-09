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
	"regexp"
	"strings"
)

type InitiateHandler struct {
	redisClient *redis.Client
	config      *helpers.Config
}

type InitiateRequest struct {
	ProjectId      int    `json:"project_id"`
	DropFolderPath string `json:"drop_folder"`
}

func makeRelativePath(inputPath string, config *helpers.Config) (string, error) {
	serverSidePathConverter := regexp.MustCompile("^/Volumes")
	potentialPaths := []string{inputPath, serverSidePathConverter.ReplaceAllString(inputPath, "/srv")}

	for _, path := range potentialPaths {
		if strings.HasPrefix(path, config.StoragePrefix.LocalPath) {
			return path[len(config.StoragePrefix.LocalPath):], nil
		}
	}
	return "", errors.New("invalid destination path, please check the settings")

}

func (h InitiateHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	bodyContent, readErr := ioutil.ReadAll(request.Body)

	if readErr != nil {
		log.Printf("ERROR InitiateHandler could not read request from client: %s", readErr)
		return
	}

	if !helpers.AssertHttpMethod(request, w, "POST") {
		return //error has already been written
	}

	username, validationErr := helpers.ValidateLogin(request, h.config)
	if validationErr != nil {
		log.Printf("ERROR InitiateHandler could not validate request: %s", validationErr)
		response := helpers.GenericErrorResponse{
			Status: "forbidden",
			Detail: validationErr.Error(),
		}
		helpers.WriteJsonContent(response, w, 403)
		return
	}

	log.Printf("INFO InitiateHandler upload initiate request from %s", username)

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
	if newSlotErr != nil {
		log.Print("ERROR InitiateHandler could not create new upload slot: ", newSlotErr)
		response := helpers.GenericErrorResponse{
			Status: "server_error",
			Detail: newSlotErr.Error(),
		}
		helpers.WriteJsonContent(response, w, 500)
		return
	}

	writErr := models.WriteUploadSlot(&newSlot, h.redisClient)
	if writErr != nil {
		log.Print("ERROR InitiateHandler could not write upload slot to storage: ", writErr)
		response := helpers.GenericErrorResponse{
			Status: "db_error",
			Detail: writErr.Error(),
		}
		helpers.WriteJsonContent(response, w, 500)
		return
	}

	response := map[string]interface{}{
		"status": "ok",
		"result": newSlot,
	}
	helpers.WriteJsonContent(response, w, 200)
}
