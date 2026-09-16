package api

import (
	"log"
	"net/http"

	"licensing-cost/internal/app/handler"
	"licensing-cost/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Starting server...")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Fatal("Ошибка репозитория:", err)
	}
	h := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	// Переименованные маршруты
	r.GET("/licensings", h.LicensingsGrid)
	r.GET("/licensing/:id", h.LicensingFeed)
	r.GET("/licensings/add", h.AddLicensing)

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/licensings")
	})

	if err := r.Run(":8080"); err != nil {
		logrus.Fatal("Ошибка запуска сервера:", err)
	}
}
