package repository

import (
	"fmt"
	"sort"
	"time"
)

const minioBaseURL = "http://localhost:9000/licensing-images/"

type License struct {
	ID           int
	Title        string
	Description  string
	LicenseType  string
	PricePerUnit float64
	MinQuantity  int
	ImageURL     string
	VideoURL     string
	Likes        []int
	Status       string
	CreatedAt    string
}

type Repository struct {
	licenses []License
}

func NewRepository() (*Repository, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	licenses := []License{
		{
			ID:           1,
			Title:        "Microsoft 365 Business",
			Description:  "Комплексное решение для малого и среднего бизнеса: Office, Teams, 1 ТБ облачного хранилища.",
			LicenseType:  "per_user",
			PricePerUnit: 1200.00,
			MinQuantity:  1,
			ImageURL:     minioBaseURL + "m365_business.jpg",
			VideoURL:     minioBaseURL + "m365_business.MOV",
			Likes:        []int{5, 10, 18, 25, 33, 42, 50, 60, 70, 80, 90, 100},
			Status:       "published",
			CreatedAt:    now,
		},
		{
			ID:           3,
			Title:        "Azure DevOps Pro",
			Description:  "Предоставляет полноценный конвейер CI/CD, неограниченные репозитории Git и доски управления проектами для гибких команд.",
			LicenseType:  "per_user",
			PricePerUnit: 2800.00,
			MinQuantity:  5,
			ImageURL:     minioBaseURL + "devops_pro.jpg",
			VideoURL:     minioBaseURL + "devops_pro.MP4",
			Likes:        []int{2, 8, 14, 21, 29, 36, 44, 52, 61, 70, 79, 88, 97},
			Status:       "published",
			CreatedAt:    now,
		},
		{
			ID:           4,
			Title:        "Power BI Premium",
			Description:  "Платформа бизнес-аналитики с расширенными возможностями визуализации, отчетами и AI-инсайтами.",
			LicenseType:  "per_core",
			PricePerUnit: 3500.00,
			MinQuantity:  2,
			ImageURL:     minioBaseURL + "powerbi_premium.jpg",
			VideoURL:     minioBaseURL + "powerbi_premium.mp4",
			Likes:        []int{1, 4, 9, 16, 23, 30, 38, 47, 55, 63, 72, 81, 90, 99, 108, 117},
			Status:       "published",
			CreatedAt:    now,
		},
		{
			ID:           5,
			Title:        "Dynamics 365 Sales",
			Description:  "CRM-система для управления взаимоотношениями с клиентами, автоматизации продаж и анализа данных.",
			LicenseType:  "per_user",
			PricePerUnit: 5900.00,
			MinQuantity:  1,
			ImageURL:     minioBaseURL + "dynamics_sales.jpg",
			VideoURL:     minioBaseURL + "dynamics_sales.mp4",
			Likes:        []int{6, 12, 19, 27, 34, 42, 50, 58, 66, 74, 82, 90, 98, 106, 114, 122, 130},
			Status:       "published",
			CreatedAt:    now,
		},
		{
			ID:           6,
			Title:        "Copilot Pro (черновик)",
			Description:  "Интеграция ИИ в офисные приложения, приоритетный доступ к новым функциям.",
			LicenseType:  "per_user",
			PricePerUnit: 300.00,
			MinQuantity:  1,
			ImageURL:     minioBaseURL + "copilot_pro.jpg",
			VideoURL:     minioBaseURL + "copilot_pro.mp4",
			Likes:        []int{4, 9},
			Status:       "draft",
			CreatedAt:    now,
		},
		{
			ID:           7,
			Title:        "Legacy Business (удалён)",
			Description:  "Устаревший план, заменён на Premium",
			LicenseType:  "per_user",
			PricePerUnit: 800.00,
			MinQuantity:  1,
			ImageURL:     minioBaseURL + "legacy.jpg",
			VideoURL:     minioBaseURL + "legacy.mp4",
			Likes:        []int{11, 22},
			Status:       "deleted",
			CreatedAt:    now,
		},
	}
	return &Repository{licenses: licenses}, nil
}

func (r *Repository) GetAllPublishedLicenses() ([]License, error) {
	var res []License
	for _, l := range r.licenses {
		if l.Status == "published" {
			res = append(res, l)
		}
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("опубликованных лицензий нет")
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})
	return res, nil
}

func (r *Repository) FilterLicensesByPrice(maxPrice float64) ([]License, error) {
	pub, err := r.GetAllPublishedLicenses()
	if err != nil {
		return nil, err
	}
	var res []License
	for _, l := range pub {
		if l.PricePerUnit <= maxPrice {
			res = append(res, l)
		}
	}
	return res, nil
}

func (r *Repository) GetLicenseByID(id int) (License, error) {
	for _, l := range r.licenses {
		if l.ID == id && l.Status != "deleted" {
			return l, nil
		}
	}
	return License{}, fmt.Errorf("лицензия с ID %d не найдена", id)
}

func (r *Repository) GetNextPublishedLicense(afterID int) (License, error) {
	pub, err := r.GetAllPublishedLicenses()
	if err != nil {
		return License{}, err
	}
	for i, l := range pub {
		if l.ID == afterID && i+1 < len(pub) {
			return pub[i+1], nil
		}
	}
	return License{}, fmt.Errorf("следующей лицензии нет")
}

func (r *Repository) GetDraft() (License, error) {
	for _, l := range r.licenses {
		if l.Status == "draft" {
			return l, nil
		}
	}
	return License{}, fmt.Errorf("черновик не найден")
}
