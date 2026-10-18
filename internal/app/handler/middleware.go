package handler

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"

	"licensing-cost/internal/app/ds"
)

const jwtBearer = "Bearer "

// WithOptionalAuth — парсит токен, если он есть, но не блокирует запрос.
// Используется на публичных эндпоинтах, чтобы получить claims (is_creator / is_liked).
func (h *Handler) WithOptionalAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		jwtStr := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(jwtStr, jwtBearer) {
			ctx.Next()
			return
		}
		jwtStr = jwtStr[len(jwtBearer):]

		err := h.Redis.CheckJWTInBlacklist(ctx.Request.Context(), jwtStr)
		if err == nil {
			ctx.Next()
			return
		}
		if !errors.Is(err, redis.Nil) {
			log.Println("redis error:", err)
			ctx.Next()
			return
		}

		token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.Config.JWT.Token), nil
		})
		if err != nil {
			ctx.Next()
			return
		}

		claims, ok := token.Claims.(*ds.JWTClaims)
		if ok && token.Valid {
			ctx.Set("claims", claims)
			ctx.Set("user_id", claims.UserID)
		}
		ctx.Next()
	}
}

// WithAuthCheck — требует токен и проверяет роль. Блокирует запрос при ошибке.
func (h *Handler) WithAuthCheck(assignedRoles ...int) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		jwtStr := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(jwtStr, jwtBearer) {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		jwtStr = jwtStr[len(jwtBearer):]

		// 1. Проверка blacklist
		err := h.Redis.CheckJWTInBlacklist(ctx.Request.Context(), jwtStr)
		if err == nil {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		if !errors.Is(err, redis.Nil) {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// 2. Парсинг JWT
		token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.Config.JWT.Token), nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusForbidden)
			log.Println("jwt parse error:", err)
			return
		}

		claims, ok := token.Claims.(*ds.JWTClaims)
		if !ok || !token.Valid {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		// 3. Проверка роли
		for _, role := range assignedRoles {
			if claims.Role == role {
				// 4. ВАЖНО: сохраняем claims в контекст
				ctx.Set("claims", claims)
				ctx.Set("user_id", claims.UserID) // ← дублируем для надёжности
				ctx.Next()
				return
			}
		}

		ctx.AbortWithStatus(http.StatusForbidden)
	}
}
