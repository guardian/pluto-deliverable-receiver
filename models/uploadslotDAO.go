package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
	"log"
	"time"
)

/**
retrieve the upload slot for the given id
 */
func UploadSlotForId(id uuid.UUID, redis *redis.Client) (UploadSlot, error) {
	keyPath := fmt.Sprintf("receiver:upload_slot:%s", id.String())

	content, getErr := redis.Get(keyPath).Result()
	if getErr != nil {
		log.Printf("ERROR models.UploadSlotForId could not get value for '%s': %s", keyPath, getErr)
		return UploadSlot{}, getErr
	}

	var s UploadSlot
	unmarshalErr := json.Unmarshal([]byte(content), &s)
	if unmarshalErr != nil {
		log.Printf("ERROR models.UploadSlotForId could not parse value for '%s': %s. Deleting the corrupted value.", keyPath, unmarshalErr)
		log.Print("ERROR models.UploadSlotForId content was ", content)
		redis.Del(keyPath)
		return UploadSlot{}, unmarshalErr
	}

	return s, nil
}

/**
write the given slot data to the storage layer
 */
func WriteUploadSlot(s *UploadSlot, redis *redis.Client) error {
	keyPath := fmt.Sprintf("receiver:upload_slot:%s", s.Uuid.String())

	encodedData, marshalErr := json.Marshal(s)
	if marshalErr != nil {
		log.Printf("ERROR models.WriteUploadSlot Could not marshal object for '%s': %s", keyPath, marshalErr)
		log.Printf("ERROR models.WriteUploadSlot offending data was %v", *s)
		return marshalErr
	}

	expiry := s.Expiry.Sub(time.Now())
	if expiry < 0 {
		log.Printf("ERROR models.WriteUploadSlot projected expiry time for %s is in the past, can't set value", keyPath)
		return errors.New("expiry time is in the past")
	}

	if writeErr := redis.Set(keyPath, string(encodedData), expiry).Err(); writeErr != nil {
		log.Printf("ERROR models.WriteUploadSlot could not write upload slot to storage: %s", writeErr)
		return errors.New("could not write data")
	}
	return nil
}