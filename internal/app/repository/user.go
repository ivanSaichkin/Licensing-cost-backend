package repository

import (
	"crypto/sha256"
	"encoding/hex"

	"licensing-cost/internal/app/ds"
)

func HashPassword(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func (r *Repository) CreateUser(login, password string) (ds.User, error) {
	user := ds.User{
		Login:       login,
		Password:    HashPassword(password),
		IsModerator: false,
	}
	if err := r.db.Create(&user).Error; err != nil {
		return ds.User{}, err
	}
	return user, nil
}

func (r *Repository) GetUserByLogin(login string) (*ds.User, error) {
	var user ds.User
	err := r.db.Where("login = ?", login).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
