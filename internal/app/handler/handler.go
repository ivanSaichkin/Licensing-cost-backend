package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"licensing-cost/internal/app/repository"
	"licensing-cost/internal/app/singleton"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	api := router.Group("/api")
	{
		// 3 GET
		api.GET("/licensings", h.GetLicensings)
		api.GET("/licensings/feed", h.GetLicensingReel)
		api.GET("/licensings/draft", h.GetLicensingDraft)

		// POST / PUT / DELETE
		api.POST("/licensings", h.CreateLicensing)
		api.PUT("/licensings/:id/publish", h.PublishLicensing)
		api.DELETE("/licensings/:id", h.DeleteLicensing)
		api.POST("/licensings/:id/like", h.LikeLicensing)

		// Домен пользователя
		api.POST("/users/register", h.RegisterUser)
		api.POST("/users/login", h.LoginUser)
		api.POST("/users/logout", h.LogoutUser)
	}
}

func (h *Handler) errorHandler(ctx *gin.Context, code int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(code, gin.H{"message": err.Error()})
}

func CurrentUserID() uint {
	return singleton.CurrentUserID()
}
