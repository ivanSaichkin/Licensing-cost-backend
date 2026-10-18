package handler

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"

	"licensing-cost/internal/app/ds"
	"licensing-cost/internal/app/repository"
)

// RegisterUser godoc
// @Summary      Регистрация
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request  body      ds.RegisterRequest  true  "Данные"
// @Success      201      {object}  ds.User
// @Failure      400      {string}  string  "Bad Request"
// @Failure      500      {string}  string  "Internal"
// @Router       /users/register [post]
func (h *Handler) RegisterUser(ctx *gin.Context) {
	var req ds.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.CreateUser(req.Login, req.Password)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

// LoginUser godoc
// @Summary      Вход
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request  body      ds.LoginRequest  true  "Логин и пароль"
// @Success      200      {object}  ds.LoginResponse
// @Failure      400      {string}  string  "Bad Request"
// @Failure      403      {string}  string  "Forbidden"
// @Router       /users/login [post]
func (h *Handler) LoginUser(ctx *gin.Context) {
	var req ds.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.GetUserByLogin(req.Login)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if user.Password != repository.HashPassword(req.Password) {
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}

	role := 0
	if user.IsModerator {
		role = 1
	}

	expiresIn := h.Config.JWT.ExpiresIn
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &ds.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiresIn).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "licensing-backend",
		},
		UserID: user.ID,
		Role:   role,
	})

	strToken, err := token.SignedString([]byte(h.Config.JWT.Token))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, ds.LoginResponse{
		ExpiresIn:   int64(expiresIn.Seconds()),
		AccessToken: strToken,
		TokenType:   "Bearer",
	})
}

// LogoutUser godoc
// @Summary      Выход
// @Tags         users
// @Produce      json
// @Security     Bearer
// @Success      200  {string}  string  "OK"
// @Failure      400  {string}  string  "Bad Request"
// @Failure      500  {string}  string  "Internal"
// @Router       /users/logout [post]
func (h *Handler) LogoutUser(ctx *gin.Context) {
	jwtStr := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, jwtBearer) {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	jwtStr = jwtStr[len(jwtBearer):]

	_, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.Config.JWT.Token), nil
	})
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Redis.WriteJWTToBlacklist(ctx.Request.Context(), jwtStr, h.Config.JWT.ExpiresIn)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		log.Println(err)
		return
	}
	ctx.Status(http.StatusOK)
}
