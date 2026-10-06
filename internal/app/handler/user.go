package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"licensing-cost/internal/app/ds"
)

// POST /api/users/register
func (h *Handler) RegisterUserAPI(ctx *gin.Context) {
	var request ds.RegisterRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.CreateUser(request.Login, request.Password)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

// POST /api/users/login — заглушка для ЛР4
func (h *Handler) LoginUserAPI(ctx *gin.Context) {
	var request ds.LoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "заглушка авторизации"})
}

// POST /api/users/logout — заглушка для ЛР4
func (h *Handler) LogoutUserAPI(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "заглушка деавторизации"})
}
