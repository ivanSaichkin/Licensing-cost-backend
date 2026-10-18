package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"licensing-cost/internal/app/config"
	"licensing-cost/internal/app/ds"
	"licensing-cost/internal/app/redis"
	"licensing-cost/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
	Redis      *redis.Client
	Config     *config.Config
}

func NewHandler(r *repository.Repository, rd *redis.Client, cfg *config.Config) *Handler {
	return &Handler{Repository: r, Redis: rd, Config: cfg}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	api := router.Group("/api")

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Регистрация/логин — всегда доступны
	api.POST("/users/register", h.RegisterUser)
	api.POST("/users/login", h.LoginUser)

	// Публичные — с опциональным токеном
	public := api.Group("")
	public.Use(h.WithOptionalAuth())
	{
		public.GET("/licensings", h.GetLicensings)
		public.GET("/licensing/feed", h.GetLicensingReel)
	}

	// Авторизованные
	auth := api.Group("")
	auth.Use(h.WithAuthCheck(0, 1))
	{
		auth.POST("/users/logout", h.LogoutUser)
		auth.GET("/licensings/draft", h.GetLicensingDraft)
		auth.POST("/licensings", h.CreateLicensing)
		auth.POST("/licensings/:id/like", h.LikeLicensing)
	}

	// Только создатель
	creator := api.Group("")
	creator.Use(h.WithAuthCheck(0))
	{
		creator.PUT("/licensings/:id/publish", h.PublishLicensing)
		creator.DELETE("/licensings/:id", h.DeleteLicensing)
	}
}

func (h *Handler) errorHandler(ctx *gin.Context, code int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(code, gin.H{"message": err.Error()})
}

func CurrentUserID(ctx *gin.Context) uint {
	// Сначала пробуем "user_id"
	if v, ok := ctx.Get("user_id"); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	// Потом "claims"
	if v, ok := ctx.Get("claims"); ok {
		if c, ok := v.(*ds.JWTClaims); ok {
			return c.UserID
		}
	}
	return 0
}
