package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"gitlab.com/codmill/customer-projects/guardian/deliverable-receiver/helpers"
	"gitlab.com/codmill/customer-projects/guardian/deliverable-receiver/models"
	"io"
	"log"
	"net/http"
	"os"
)

type ValidateChecksumHandler struct {
	redisClient *redis.Client
	config      *helpers.Config
}

func GetLocalSHA(filename string) (string, error) {
	f, openErr := os.Open(filename)
	if openErr != nil {
		log.Printf("ERROR GetLocalSHA could not open %s: %s", filename, openErr)
		return "", openErr
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Printf("ERROR GetLocalSHA could not read local file %s: %s", filename, err)
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

func (h ValidateChecksumHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	if !helpers.AssertHttpMethod(request, w, "GET") {
		return
	}

	username, validationErr := helpers.ValidateLogin(request, h.config)
	if validationErr != nil {
		log.Printf("ERROR ValidateChecksumHandler could not validate request: %s", validationErr)
		response := helpers.GenericErrorResponse{
			Status: "forbidden",
			Detail: validationErr.Error(),
		}
		helpers.WriteJsonContent(response, w, 403)
		return
	}

	params, getParamsErr := helpers.GetQueryParams(request.RequestURI)
	if getParamsErr != nil {
		log.Printf("ERROR ValidateChecksumHandler could not get request params from '%s': %s", request.RequestURI, getParamsErr)
		response := helpers.GenericErrorResponse{
			Status: "bad_request",
			Detail: "invalid request params",
		}
		helpers.WriteJsonContent(response, w, 400)
		return
	}

	fileName := params.Get("fileName")
	uploadId := params.Get("uploadId")
	clientChecksum := params.Get("sum")

	if fileName == "" || uploadId == "" || clientChecksum == "" {
		response := helpers.GenericErrorResponse{
			Status: "bad_request",
			Detail: "You must provide fileName, uploadId and sum as query parameters",
		}
		helpers.WriteJsonContent(response, w, 400)
		return
	}

	uploadSlotUuid, uuidErr := uuid.Parse(uploadId)
	if uuidErr != nil {
		response := helpers.GenericErrorResponse{
			Status: "invalid_request",
			Detail: "uploadId must be a uuid",
		}
		helpers.WriteJsonContent(response, w, 400)
		return
	}

	uploadSlot, slotErr := models.UploadSlotForId(uploadSlotUuid, h.redisClient)
	if slotErr != nil {
		log.Printf("ERROR ValidateChecksumHandler could not get upload slot for '%s': %s", uploadSlotUuid, slotErr)
		response := helpers.GenericErrorResponse{
			Status: "db_error",
			Detail: "could not get upload slot",
		}
		helpers.WriteJsonContent(response, w, 500)
		return
	}

	if uploadSlot == nil {
		log.Printf("WARNING ValidateChecksumHandler no upload slot for '%s', this might just mean it's expire", uploadSlotUuid)
		response := helpers.GenericErrorResponse{
			Status: "not_found",
			Detail: "upload slot does not exist",
		}
		helpers.WriteJsonContent(response, w, 404)
		return
	}

	//get a sanitised, absolute path to write the file
	targetFilename := getTargetFilename(h.config.StoragePrefix.LocalPath, uploadSlot.UploadPathRelative, fileName)

	log.Printf("INFO ValidateChecksumHandler checksum request from %s to %s", username, targetFilename)
	sha, shaErr := GetLocalSHA(targetFilename)
	if shaErr != nil {
		log.Printf("ERROR ValidateChecksumHandler could not calculate local SHA: %s", shaErr)
		response := helpers.GenericErrorResponse{
			Status: "error",
			Detail: shaErr.Error(),
		}
		helpers.WriteJsonContent(response, w, 500)
		return
	}

	log.Printf("INFO ValidateChecksumHandler local SHA is %s, client SHA is %s", sha, clientChecksum)
	if sha == clientChecksum {
		response := helpers.GenericErrorResponse{
			Status: "ok",
			Detail: "checksums matched",
		}
		helpers.WriteJsonContent(response, w, 200)
		return
	} else {
		response := helpers.GenericErrorResponse{
			Status: "conflict",
			Detail: fmt.Sprintf("local sha is %s but client is %s", sha, clientChecksum),
		}
		helpers.WriteJsonContent(response, w, 409)
		return
	}
}
