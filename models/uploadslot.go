package models

import (
	"errors"
	"github.com/google/uuid"
	"gitlab.com/codmill/customer-projects/guardian/deliverable-receiver/helpers"
	"log"
	"path"
	"strings"
	"time"
)

type UploadSlot struct {
	Uuid uuid.UUID	`json:"uuid"`
	UploadPathRelative string 	`json:"upload_path_relative"`
	ProjectId int `json:"project_id"`
	Expiry time.Time `json:"expiry"`
}

/**
take in data from the configuration to determine the actual output path
 */
func (s UploadSlot) GetFullPath(config *helpers.Config) (string, error) {
	if config.StoragePrefix.LocalPath == "" {
		return "", errors.New("No storage prefix set in configuration")
	}

	targetPath := path.Join(config.StoragePrefix.LocalPath, s.UploadPathRelative)
	if !strings.HasPrefix(targetPath, "/") {
		return "", errors.New("Upload path is not absolute, please check storage_prefix.localpath in the settings")
	}

	return targetPath, nil
}

func NewUploadSlot(projectId int, uploadPathRelative string, ttl time.Duration) (UploadSlot, error) {
	uid, uidErr := uuid.NewRandom()
	if uidErr != nil {
		log.Print("ERROR models.NewUploadSlot could not generate uuid: ", uidErr)
		return UploadSlot{}, errors.New("could not generate uuid")
	}
	return UploadSlot{
		uid,
		uploadPathRelative,
		projectId,
		time.Now().Add(ttl),
	}, nil
}