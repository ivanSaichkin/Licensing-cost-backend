package repository

import (
	"errors"
	"fmt"
	"time"

	"licensing-cost/internal/app/ds"

	"gorm.io/gorm"
)

// GetAllPublishedLicensings — все опубликованные, не удалённые
func (r *Repository) GetAllPublishedLicensings() ([]ds.LicensingModel, error) {
	var licensings []ds.LicensingModel
	err := r.db.
		Where("status = ?", "published").
		Order("id ASC").
		Find(&licensings).Error
	if err != nil {
		return nil, err
	}
	if len(licensings) == 0 {
		return nil, fmt.Errorf("опубликованных моделей нет")
	}
	return licensings, nil
}

// FilterLicensingsByCommission — фильтр по комиссии
func (r *Repository) FilterLicensingsByCommission(maxCommission float64) ([]ds.LicensingModel, error) {
	var licensings []ds.LicensingModel
	err := r.db.
		Where("status = ? AND commission_per_unit <= ?", "published", maxCommission).
		Order("id ASC").
		Find(&licensings).Error
	if err != nil {
		return nil, err
	}
	return licensings, nil
}

// GetLicensingByID — одна модель, не удалённая
func (r *Repository) GetLicensingByID(id int) (*ds.LicensingModel, error) {
	var licensing ds.LicensingModel
	err := r.db.
		Where("id = ? AND status != ?", id, "deleted").
		First(&licensing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &licensing, nil
}

// GetNextPublishedLicensing — следующая по ID, только published
func (r *Repository) GetNextPublishedLicensing(afterID int) (*ds.LicensingModel, error) {
	var licensing ds.LicensingModel
	err := r.db.
		Where("id > ? AND status = ?", afterID, "published").
		Order("id ASC").
		First(&licensing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &licensing, nil
}

// GetDraft — черновик текущего пользователя
func (r *Repository) GetDraft() (*ds.LicensingModel, error) {
	var licensing ds.LicensingModel
	err := r.db.
		Where("creator_id = ? AND status = ?", 1, "draft").
		First(&licensing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &licensing, nil
}

// CreateDraft — создание через ORM
func (r *Repository) CreateDraft(title, description, licenseType, imageURL, videoURL string, commission float64, minForCalc int) (*ds.LicensingModel, error) {
	licensing := &ds.LicensingModel{
		Title:             title,
		Description:       description,
		LicenseType:       licenseType,
		CommissionPerUnit: commission,
		MinForCalc:        minForCalc,
		ImageURL:          imageURL,
		VideoURL:          videoURL,
		Status:            "draft",
		CreatorID:         1,
		CreatedAt:         time.Now(),
	}
	if err := r.db.Create(licensing).Error; err != nil {
		return nil, err
	}
	return licensing, nil
}

// PublishLicensing — публикация через ORM
func (r *Repository) PublishLicensing(id uint) error {
	return r.db.Model(&ds.LicensingModel{}).
		Where("id = ? AND status = ?", id, "draft").
		Updates(map[string]interface{}{
			"status":       "published",
			"published_at": time.Now(),
		}).Error
}

// DeleteLicensing — удаление через SQL UPDATE (без ORM): статус → deleted
func (r *Repository) DeleteLicensing(id uint) error {
	query := "UPDATE licensing_models SET status = 'deleted' WHERE id = $1"
	return r.db.Exec(query, id).Error
}

// GetLikesCount — количество лайков
func (r *Repository) GetLikesCount(licensingID uint) int64 {
	var count int64
	r.db.Model(&ds.Like{}).Where("licensing_id = ?", licensingID).Count(&count)
	return count
}
