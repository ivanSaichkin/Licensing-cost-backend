package pkg

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"licensing-cost/internal/app/config"
	"licensing-cost/internal/app/handler"
	"licensing-cost/internal/app/redis"
	"licensing-cost/internal/app/repository"
)

type Application struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *handler.Handler
	Redis   *redis.Client
}

func NewApp(
	ctx context.Context,
	cfg *config.Config,
	r *gin.Engine,
	repo *repository.Repository,
) (*Application, error) {
	redisClient, err := redis.New(ctx, cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("cant create redis: %w", err)
	}

	h := handler.NewHandler(repo, redisClient, cfg)

	return &Application{
		Config:  cfg,
		Router:  r,
		Handler: h,
		Redis:   redisClient,
	}, nil
}

func (a *Application) RunApp() {
	logrus.Info("Server start up")
	a.Handler.RegisterHandler(a.Router)

	addr := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(addr); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}
