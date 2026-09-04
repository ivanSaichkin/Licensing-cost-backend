package api

import (
	"licensing-cost/internal/app/handler"
	"licensing-cost/internal/app/repository"
	"log"

	"github.com/gin-gonic/gin"
)

func StartServer() {
	log.Println("Server start up")

	repo := repository.NewRepository()
	h := handler.NewHandler(repo)

	r := gin.Default()

	// Подключаем шаблоны и статику
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	// Маршруты (все GET, как требуется в первой лабе)
	r.GET("/feed", h.FeedHandler)
	r.GET("/add", h.AddHandler)
	r.GET("/grid", h.GridHandler)

	// Запуск на :8080
	if err := r.Run(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
	log.Println("Server down")
}
