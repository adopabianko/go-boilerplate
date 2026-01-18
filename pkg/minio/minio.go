package minio

import (
	"log"

	"go-boilerplate/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func Connect(cfg config.MinioConfig) *minio.Client {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Fatalf("Failed to connect to Minio: %v", err)
	}

	return minioClient
}
