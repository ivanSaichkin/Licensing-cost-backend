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
		api.GET("/licensings", h.GetLicensingsAPI)
		api.GET("/licensings/feed", h.GetLicensingFeedAPI)
		api.GET("/licensings/draft", h.GetLicensingDraftAPI)

		// POST / PUT / DELETE
		api.POST("/licensings", h.CreateLicensingAPI)
		api.PUT("/licensings/:id/publish", h.PublishLicensingAPI)
		api.DELETE("/licensings/:id", h.DeleteLicensingAPI)
		api.POST("/licensings/:id/like", h.LikeLicensingAPI)

		// Домен пользователя
		api.POST("/users/register", h.RegisterUserAPI)
		api.POST("/users/login", h.LoginUserAPI)
		api.POST("/users/logout", h.LogoutUserAPI)
	}
}

func (h *Handler) errorHandler(ctx *gin.Context, code int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(code, gin.H{"message": err.Error()})
}

func CurrentUserID() uint {
	return singleton.CurrentUserID()
}
