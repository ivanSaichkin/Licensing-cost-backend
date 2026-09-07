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

	r.GET("/grid", h.GridLicenses)
	r.GET("/feed/:id", h.FeedLicense)
	r.GET("/add", h.AddLicense)

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/grid")
	})

	if err := r.Run(":8080"); err != nil {
		logrus.Fatal("Ошибка запуска сервера:", err)
	}
}
