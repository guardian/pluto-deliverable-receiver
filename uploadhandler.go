package main

import (
	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
	"gitlab.com/codmill/customer-projects/guardian/deliverable-receiver/helpers"
	"gitlab.com/codmill/customer-projects/guardian/deliverable-receiver/models"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
)

type UploadHandler struct {
	redisClient *redis.Client
	config      *helpers.Config
}

/**
sanitises the given incoming filename and combines it with the base path from config and the slot path
*/
func getTargetFilename(configBase string, slotBase string, requestedFilename string) string {
	fileBase := path.Base(requestedFilename)
	sanitizer := regexp.MustCompile("[^\\w\\d\\s.]")
	sanitizedFileBase := sanitizer.ReplaceAllString(fileBase, "")

	return path.Join(configBase, slotBase, sanitizedFileBase)
}

/**
writes the data from the given reader out to the given filename.
creates the parent directory if necessary
*/
func writeOutData(fullpath string, maybeRange *helpers.RangeHeader, content io.Reader) (int64, error) {
	dirpath := path.Dir(fullpath)
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		log.Printf("INFO Uploadandler.writeOutData target path %s does not exist, creating", dirpath)
		mkDirErr := os.MkdirAll(dirpath, 0775)
		if mkDirErr != nil {
			log.Printf("ERROR UploadHandler.writeOutData could not create directory %s: %s", dirpath, mkDirErr)
			return -1, mkDirErr
		}
	}

	//FIXME: this over-writes an existing file
	openFlags := os.O_WRONLY
	if maybeRange != nil && maybeRange.IsFirst() {
		openFlags |= os.O_CREATE | os.O_TRUNC
	}

	f, openErr := os.OpenFile(fullpath, openFlags, 0664)
	if openErr != nil {
		log.Printf("ERROR UploadHandler.writeOutData could not open %s to write: %s", fullpath, openErr)
		return -1, openErr
	}
	defer f.Close()

	if maybeRange == nil || maybeRange.IsComplete() {
		log.Printf("INFO UploadHandler.writeOutData no range so writing whole file")
		return io.Copy(f, content)
	} else {
		_, seekErr := f.Seek(maybeRange.Start, os.SEEK_SET)
		if seekErr != nil {
			log.Printf("ERROR UploadHandler.writeOutData could not seek '%s': %s", fullpath, seekErr)
			return -1, seekErr
		}
		//FIXME: check if this actually writes correctly
		return io.Copy(f, content)
	}
}

/**
perform an actual file upload.
expects a POST request with two query parameters:
- fileName: filename with no path. If any path parts are present, they are stripped.
- uploadId: uuid of an existing upload created by calling /initiate.  If it does not exist (any more?) a 404 is returned;
the client should try to create a new one and then retry the upload.
*/
func (h UploadHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	if request.Body != nil {
		defer request.Body.Close()
	}

	if !helpers.AssertHttpMethod(request, w, "POST") {
		io.Copy(ioutil.Discard, request.Body) //discard any remaining body
		return
	}

	username, validationErr := helpers.ValidateLogin(request, h.config)
	if validationErr != nil {
		log.Printf("ERROR UploadHandler could not validate request: %s", validationErr)
		response := helpers.GenericErrorResponse{
			Status: "forbidden",
			Detail: validationErr.Error(),
		}
		helpers.WriteJsonContent(response, w, 403)
		return
	}

	values, queryErr := helpers.GetQueryParams(request.RequestURI)
	if queryErr != nil {
		log.Print("ERROR UploadHandler could not parse own url: ", queryErr)
		response := helpers.GenericErrorResponse{
			Status: "server_error",
			Detail: "invalid url",
		}
		helpers.WriteJsonContent(response, w, 400)
		return
	}
	uploadId := values.Get("uploadId")
	fileName := values.Get("fileName")

	if uploadId == "" {
		response := helpers.GenericErrorResponse{
			Status: "invalid_request",
			Detail: "you must specify the uploadId query parameter",
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

	if fileName == "" {
		response := helpers.GenericErrorResponse{
			Status: "invalid_request",
			Detail: "you must specify the fileName query parameter",
		}
		helpers.WriteJsonContent(response, w, 400)
		return
	}

	maybeRangeHeader, rangeErr := helpers.ExtractRange(request)
	if rangeErr != nil {
		log.Printf("ERROR UploadHandler could not parse range parameter '%s': %s", request.Header.Get("Range"), rangeErr)
		response := helpers.GenericErrorResponse{
			Status: "invalid_request",
			Detail: rangeErr.Error(),
		}
		helpers.WriteJsonContent(response, w, 400)
		return
	}

	uploadSlot, slotErr := models.UploadSlotForId(uploadSlotUuid, h.redisClient)
	if slotErr != nil {
		log.Printf("ERROR UploadHandler could not get upload slot for '%s': %s", uploadSlotUuid, slotErr)
		response := helpers.GenericErrorResponse{
			Status: "db_error",
			Detail: "could not get upload slot",
		}
		helpers.WriteJsonContent(response, w, 500)
		return
	}

	if uploadSlot == nil {
		log.Printf("WARNING UploadHandler no upload slot for '%s', this might just mean it's expire", uploadSlotUuid)
		response := helpers.GenericErrorResponse{
			Status: "not_found",
			Detail: "upload slot does not exist",
		}
		helpers.WriteJsonContent(response, w, 404)
		return
	}

	//get a sanitised, absolute path to write the file
	targetFilename := getTargetFilename(h.config.StoragePrefix.LocalPath, uploadSlot.UploadPathRelative, fileName)

	log.Printf("INFO UploadHandler upload request from %s to %s", username, targetFilename)

	bytesWritten, writeErr := writeOutData(targetFilename, maybeRangeHeader, request.Body)
	if writeErr != nil {
		response := helpers.GenericErrorResponse{
			Status: "write_error",
			Detail: writeErr.Error(),
		}
		helpers.WriteJsonContent(response, w, 500)
		return
	}

	response := map[string]interface{}{
		"status":        "ok",
		"bytes_written": bytesWritten,
	}
	helpers.WriteJsonContent(response, w, 200)
}
