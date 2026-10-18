package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	_ "licensing-cost/docs"

	"licensing-cost/internal/app/config"
	"licensing-cost/internal/app/dsn"
	"licensing-cost/internal/app/repository"
	"licensing-cost/internal/pkg"
)

// @title           Licensing Backend API
// @version         1.0
// @description     API для работы с моделями лицензирования
// @host            localhost:8080
// @BasePath        /api
// @securityDefinitions.apikey  Bearer
// @in                          header
// @name                        Authorization
// @description                 Введите токен: Bearer {token}
func main() {
	_ = godotenv.Load()

	ctx := context.Background()

	conf, err := config.NewConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	repo, err := repository.New(&repository.RepositorySettings{
		PostgresDSN:     dsn.FromEnv(),
		MinioEndpoint:   os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey:  os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey:  os.Getenv("MINIO_SECRET_KEY"),
		MinioBucketName: os.Getenv("MINIO_BUCKET_NAME"),
		MinioUseSSL:     os.Getenv("MINIO_USE_SSL") == "true",
	})
	if err != nil {
		log.Fatalf("repo error: %v", err)
	}

	router := gin.Default()
	app, err := pkg.NewApp(ctx, conf, router, repo)
	if err != nil {
		log.Fatalf("app error: %v", err)
	}

	logrus.Info("Application start!")
	app.RunApp()
	logrus.Info("Application terminated!")
}
