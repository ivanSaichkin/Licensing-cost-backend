package repository

import (
	"fmt"
	"sort"
	"time"
)

const minioBaseURL = "http://localhost:9000/licensing-images/"

// LicensingModel — модель лицензирования (услуга)
type LicensingModel struct {
	ID                int
	Title             string // название модели, например "Per User"
	Description       string
	LicenseType       string  // тип: per_user, per_core, subscription, enterprise, concurrent
	CommissionPerUnit float64 // комиссия за 1 единицу
	MinForCalc        int     // минимальное число для расчёта
	ImageURL          string
	VideoURL          string
	Likes             []int
	Status            string
	CreatedAt         string
}

type Repository struct {
	licensings []LicensingModel
}

func NewRepository() (*Repository, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	licensings := []LicensingModel{
		{
			ID:                1,
			Title:             "Per User",
			Description:       "Модель лицензирования на одного пользователя. Комиссия начисляется за каждое активное рабочее место.",
			LicenseType:       "per_user",
			CommissionPerUnit: 120.00,
			MinForCalc:        1,
			ImageURL:          minioBaseURL + "per_user.jpg",
			VideoURL:          minioBaseURL + "per_user.mp4",
			Likes:             []int{5, 10, 18, 25, 33, 42, 50, 60, 70, 80, 90, 100},
			Status:            "published",
			CreatedAt:         now,
		},
		{
			ID:                2,
			Title:             "Per Core",
			Description:       "Модель лицензирования по ядрам процессора. Комиссия рассчитывается на каждое физическое ядро сервера.",
			LicenseType:       "per_core",
			CommissionPerUnit: 280.00,
			MinForCalc:        4,
			ImageURL:          minioBaseURL + "per_core.jpg",
			VideoURL:          minioBaseURL + "per_core.mp4",
			Likes:             []int{2, 8, 14, 21, 29, 36, 44, 52, 61, 70, 79, 88, 97},
			Status:            "published",
			CreatedAt:         now,
		},
		{
			ID:                3,
			Title:             "Subscription",
			Description:       "Подписочная модель. Ежемесячная комиссия за доступ к продукту без права бессрочного использования.",
			LicenseType:       "subscription",
			CommissionPerUnit: 350.00,
			MinForCalc:        1,
			ImageURL:          minioBaseURL + "subscription.jpg",
			VideoURL:          minioBaseURL + "subscription.mp4",
			Likes:             []int{1, 4, 9, 16, 23, 30, 38, 47, 55, 63, 72, 81, 90, 99, 108, 117},
			Status:            "published",
			CreatedAt:         now,
		},
		{
			ID:                4,
			Title:             "Enterprise Agreement",
			Description:       "Корпоративное соглашение с фиксированной комиссией за весь пакет. Минимальный порог входа — 50 единиц.",
			LicenseType:       "enterprise",
			CommissionPerUnit: 590.00,
			MinForCalc:        50,
			ImageURL:          minioBaseURL + "enterprise.jpg",
			VideoURL:          minioBaseURL + "enterprise.mp4",
			Likes:             []int{6, 12, 19, 27, 34, 42, 50, 58, 66, 74, 82, 90, 98, 106, 114, 122, 130},
			Status:            "published",
			CreatedAt:         now,
		},
		{
			ID:                5,
			Title:             "Concurrent",
			Description:       "Плавающая лицензия. Комиссия начисляется за одновременные подключения, а не за всех пользователей.",
			LicenseType:       "concurrent",
			CommissionPerUnit: 220.00,
			MinForCalc:        5,
			ImageURL:          minioBaseURL + "concurrent.jpg",
			VideoURL:          minioBaseURL + "concurrent.mp4",
			Likes:             []int{3, 7, 11, 15, 22, 28, 35, 43, 51, 59, 67, 75},
			Status:            "published",
			CreatedAt:         now,
		},
		{
			ID:                6,
			Title:             "Trial (черновик)",
			Description:       "Пробная модель. Нулевая комиссия на период тестирования.",
			LicenseType:       "trial",
			CommissionPerUnit: 0.00,
			MinForCalc:        1,
			ImageURL:          minioBaseURL + "trial.jpg",
			VideoURL:          minioBaseURL + "trial.mp4",
			Likes:             []int{4, 9},
			Status:            "draft",
			CreatedAt:         now,
		},
		{
			ID:                7,
			Title:             "Legacy (удалён)",
			Description:       "Устаревшая модель лицензирования",
			LicenseType:       "legacy",
			CommissionPerUnit: 80.00,
			MinForCalc:        1,
			ImageURL:          minioBaseURL + "legacy.jpg",
			VideoURL:          minioBaseURL + "legacy.mp4",
			Likes:             []int{11, 22},
			Status:            "deleted",
			CreatedAt:         now,
		},
	}
	return &Repository{licensings: licensings}, nil
}

func (r *Repository) GetAllPublishedLicensings() ([]LicensingModel, error) {
	var res []LicensingModel
	for _, l := range r.licensings {
		if l.Status == "published" {
			res = append(res, l)
		}
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("опубликованных моделей нет")
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})
	return res, nil
}

func (r *Repository) FilterLicensingsByCommission(maxCommission float64) ([]LicensingModel, error) {
	pub, err := r.GetAllPublishedLicensings()
	if err != nil {
		return nil, err
	}
	var res []LicensingModel
	for _, l := range pub {
		if l.CommissionPerUnit <= maxCommission {
			res = append(res, l)
		}
	}
	return res, nil
}

func (r *Repository) GetLicensingByID(id int) (LicensingModel, error) {
	for _, l := range r.licensings {
		if l.ID == id && l.Status != "deleted" {
			return l, nil
		}
	}
	return LicensingModel{}, fmt.Errorf("модель с ID %d не найдена", id)
}

func (r *Repository) GetNextPublishedLicensing(afterID int) (LicensingModel, error) {
	pub, err := r.GetAllPublishedLicensings()
	if err != nil {
		return LicensingModel{}, err
	}
	for i, l := range pub {
		if l.ID == afterID && i+1 < len(pub) {
			return pub[i+1], nil
		}
	}
	return LicensingModel{}, fmt.Errorf("следующей модели нет")
}

func (r *Repository) GetDraft() (LicensingModel, error) {
	for _, l := range r.licensings {
		if l.Status == "draft" {
			return l, nil
		}
	}
	return LicensingModel{}, fmt.Errorf("черновик не найден")
}
