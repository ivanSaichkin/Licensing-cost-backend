package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"licensing-cost/internal/app/config"
	"licensing-cost/internal/app/dsn"
	"licensing-cost/internal/app/handler"
	"licensing-cost/internal/app/repository"
	"licensing-cost/internal/pkg"
)

func main() {
	_ = godotenv.Load()

	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("config error: %v", err)
	}

	rep, errRep := repository.New(&repository.RepositorySettings{
		PostgresDSN:     dsn.FromEnv(),
		MinioEndpoint:   os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey:  os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey:  os.Getenv("MINIO_SECRET_KEY"),
		MinioBucketName: os.Getenv("MINIO_BUCKET_NAME"),
		MinioUseSSL:     os.Getenv("MINIO_USE_SSL") == "true",
	})
	if errRep != nil {
		logrus.Fatalf("repository error: %v", errRep)
	}
	fmt.Println("Server initialized")

	hand := handler.NewHandler(rep)
	pkg.NewApp(conf, router, hand).RunApp()
}
