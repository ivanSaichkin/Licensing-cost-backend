package repository

import (
	"errors"
	"fmt"
	"time"

	"licensing-cost/internal/app/ds"

	"gorm.io/gorm"
)

func (r *Repository) GetAllPublishedLicensings() ([]ds.Licensing, error) {
	var licensings []ds.Licensing
	err := r.db.Where("status = ?", "published").Order("id ASC").Find(&licensings).Error
	if err != nil {
		return nil, err
	}
	if len(licensings) == 0 {
		return nil, fmt.Errorf("опубликованных моделей нет")
	}
	return licensings, nil
}

func (r *Repository) FilterLicensingsByCommission(maxCommission float64) ([]ds.Licensing, error) {
	var licensings []ds.Licensing
	err := r.db.Where("status = ? AND commission_per_unit <= ?", "published", maxCommission).
		Order("id ASC").Find(&licensings).Error
	if err != nil {
		return nil, err
	}
	return licensings, nil
}

func (r *Repository) GetLicensingByID(id int) (*ds.Licensing, error) {
	var licensing ds.Licensing
	err := r.db.Where("id = ? AND status != ?", id, "deleted").First(&licensing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &licensing, nil
}

func (r *Repository) GetNextPublishedLicensing(afterID int) (*ds.Licensing, error) {
	var licensing ds.Licensing
	err := r.db.Where("id > ? AND status = ?", afterID, "published").
		Order("id ASC").First(&licensing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &licensing, nil
}

func (r *Repository) GetDraft() (*ds.Licensing, error) {
	var licensing ds.Licensing
	err := r.db.Where("creator_id = ? AND status = ?", 1, "draft").First(&licensing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &licensing, nil
}

func (r *Repository) CreateDraft(title, description, imageURL, videoURL string, commission float64, minForCalc int) (*ds.Licensing, error) {
	licensing := &ds.Licensing{
		Title:             title,
		Description:       description,
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

func (r *Repository) PublishLicensing(id uint) error {
	return r.db.Model(&ds.Licensing{}).
		Where("id = ? AND status = ?", id, "draft").
		Updates(map[string]interface{}{
			"status":       "published",
			"published_at": time.Now(),
		}).Error
}

func (r *Repository) DeleteLicensing(id uint) error {
	query := "UPDATE licensings SET status = 'deleted' WHERE id = $1"
	return r.db.Exec(query, id).Error
}

func (r *Repository) GetLikesCount(licensingID uint) int64 {
	var count int64
	r.db.Model(&ds.Like{}).Where("licensing_id = ?", licensingID).Count(&count)
	return count
}
