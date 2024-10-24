package model

import (
	"encoding/base64"
	"encoding/json"

	"github.com/google/uuid"
)

type ObjectMetadata struct {
	ID          uuid.UUID `json:"id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"contentType"`
	Size        int64     `json:"size"`
	Duration    *float64  `json:"duration,omitempty"`
	Thumbnail   *string   `json:"thumbnail,omitempty"`
}

func (o ObjectMetadata) GenerateObjectKey() (string, error) {
	jsn, err := json.Marshal(o)
	if err != nil {
		return "", err
	}

	objectKey := base64.URLEncoding.EncodeToString(jsn)

	return objectKey, nil
}

func DecodeObjectMetadata(objectKey string) (*ObjectMetadata, error) {
	jsn, err := base64.URLEncoding.DecodeString(objectKey)
	if err != nil {
		return nil, err
	}

	var metadata ObjectMetadata

	if err := json.Unmarshal(jsn, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

type FileUploadResult struct {
	Key string
}
