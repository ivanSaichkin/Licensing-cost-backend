package repository

import (
	"licensing-cost/internal/app/ds"
)

func (r *Repository) CreateUser(login, password string) (ds.User, error) {
	user := ds.User{Login: login, Password: password, IsModerator: false}
	if err := r.db.Create(&user).Error; err != nil {
		return ds.User{}, err
	}
	return user, nil
}
