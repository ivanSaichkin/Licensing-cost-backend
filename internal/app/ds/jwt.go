package ds

import (
	"github.com/golang-jwt/jwt"
)

// JWTClaims — полезная нагрузка токена
type JWTClaims struct {
	jwt.StandardClaims
	UserID uint `json:"user_id"`
	Role   int  `json:"role"` // 0 — создатель, 1 — модератор
}
