package storage

import (
	"context"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var Client *minio.Client
var Bucket = "bundlr-artifacts"

func InitMinIO() {
	var err error
	Client, err = minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("minio", "minio123", ""),
		Secure: false,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func GeneratePresignedUpload(fileKey string) (string, error) {
	url, err := Client.PresignedPutObject(context.Background(), Bucket, fileKey, time.Minute*10)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}
