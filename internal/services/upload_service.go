package services

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"github.com/thanishsid/dingilink-server/internal/model"
)

type UploadService struct {
	S3Client  *s3.Client
	S3Bucket  string
	S3BaseDir string
}

type UploadFileInput struct {
	File              multipart.File        `json:"file"`
	FileHeader        *multipart.FileHeader `json:"fileHeader"`
	GenerateThumbnail *bool                 `json:"generateThumbnail"`
	ThumbnailPosition *float64              `json:"thumbnailPosition"`
}

func (s *UploadService) UploadFile(ctx context.Context, input UploadFileInput) (*model.FileUploadResult, error) {
	contentType := input.FileHeader.Header.Get("Content-Type")

	metadata := model.ObjectMetadata{
		ID:          uuid.New(),
		Filename:    input.FileHeader.Filename,
		ContentType: contentType,
		Size:        input.FileHeader.Size,
	}

	isVideo := strings.HasPrefix(contentType, "video")
	isAudio := strings.HasPrefix(contentType, "audio")
	var localFilePath string

	// if the file is an audio or a video then save it locally for duration extraction and thumbnail generation.
	if isAudio || isVideo {
		localFilePath = path.Join(os.TempDir(), fmt.Sprintf("upload_%s", uuid.NewString()))
		out, err := os.Create(localFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create upload local file: %w", err)
		}
		defer out.Close()
		defer os.Remove(localFilePath)

		_, err = io.Copy(out, input.File)
		if err != nil {
			return nil, fmt.Errorf("failed to copy upload to local file: %w", err)
		}

		// Reset the file pointer before uploading to S3
		if _, err := input.File.Seek(0, 0); err != nil {
			return nil, fmt.Errorf("failed to seek file to start: %w", err)
		}

		duration, err := GetMediaDuration(localFilePath)
		if err != nil {
			return nil, err
		}

		metadata.Duration = &duration

		// Generate thumbnail if file is a video
		if isVideo {
			thumbnailPath := path.Join(os.TempDir(), fmt.Sprintf("thumbnail_%s.jpg", uuid.NewString()))
			thumbnailPosition := (duration / 100) * 5

			defer os.Remove(thumbnailPath)

			if input.ThumbnailPosition != nil {
				thumbnailPosition = *input.ThumbnailPosition
			}

			thumbnail, err := GenerateVideoThumbnail(localFilePath, thumbnailPath, thumbnailPosition)
			if err != nil {
				return nil, err
			}

			// Save thumbnail to a buffer
			var buf bytes.Buffer
			err = jpeg.Encode(&buf, thumbnail, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to encode thumbnail to jepg: %w", err)
			}

			thumbnailID := uuid.New()

			thumbnailMetadata := model.ObjectMetadata{
				ID:          thumbnailID,
				Filename:    fmt.Sprintf("thumbnail_%s.jpg", thumbnailID),
				ContentType: "image/jpeg",
				Size:        int64(buf.Len()),
			}

			thumbnailObjectKey, err := thumbnailMetadata.GenerateObjectKey()
			if err != nil {
				return nil, err
			}

			thumbnailFile := bytes.NewReader(buf.Bytes())

			_, err = s.S3Client.PutObject(ctx, &s3.PutObjectInput{
				Bucket:             aws.String(s.S3Bucket),
				Key:                aws.String(thumbnailObjectKey),
				Body:               thumbnailFile,
				ContentDisposition: aws.String(fmt.Sprintf("inline; filename=%q", thumbnailMetadata.Filename)),
				ContentType:        aws.String(thumbnailMetadata.ContentType),
			})
			if err != nil {
				return nil, err
			}

			metadata.Thumbnail = &thumbnailObjectKey
		}
	}

	objectKey, err := metadata.GenerateObjectKey()
	if err != nil {
		return nil, err
	}

	if !(len(objectKey) <= 1024) {
		return nil, fmt.Errorf("file metadata is too big")
	}

	_, err = s.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:             aws.String(s.S3Bucket),
		Key:                aws.String(objectKey),
		Body:               input.File,
		ContentDisposition: aws.String(fmt.Sprintf("inline; filename=%q", input.FileHeader.Filename)),
		ContentType:        aws.String(contentType),
	})
	if err != nil {
		return nil, err
	}

	return &model.FileUploadResult{
		Key: objectKey,
	}, nil
}

// GetDuration retrieves the duration of the video/audio using ffprobe
func GetMediaDuration(path string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", path)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get media duration: %w", err)
	}

	durationString := strings.TrimSpace(string(output))

	duration, err := strconv.ParseFloat(durationString, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse media duration string: %w", err)
	}

	return duration, nil
}

// GenerateVideoThumbnail generates a thumbnail for the video using ffmpeg
func GenerateVideoThumbnail(videoPath, thumbnailPath string, position float64) (image.Image, error) {
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-ss", fmt.Sprint(position), "-vframes", "1", thumbnailPath)
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	// Open the generated thumbnail
	thumbnailFile, err := os.Open(thumbnailPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open thumbnail: %w", err)
	}
	defer thumbnailFile.Close()

	// Decode image
	img, _, err := image.Decode(thumbnailFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}
