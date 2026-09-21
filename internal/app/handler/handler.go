package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"licensing-cost/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/licensings", h.LicensingsGrid)
	router.GET("/licensing/:id", h.LicensingFeed)
	router.GET("/licensings/add", h.AddLicensing)
	router.POST("/licensings", h.CreateLicensing)
	router.POST("/licensings/publish", h.PublishLicensing)
	router.POST("/licensings/delete", h.DeleteLicensing)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

func (h *Handler) errorHandler(ctx *gin.Context, code int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(code, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
