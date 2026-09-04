package repository

import (
	"fmt"
)

// Tariff — услуга (тариф облачного провайдера)
type Tariff struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Price       int      `json:"price"`     // цена в рублях
	VCPU        int      `json:"vcpu"`      // количество vCPU
	RAM         int      `json:"ram"`       // объём RAM в ГБ
	Disk        int      `json:"disk"`      // объём диска в ГБ
	ImageURL    string   `json:"image_url"` // имя файла изображения (будет подставлен URL MinIO)
	VideoURL    string   `json:"video_url"` // имя файла видео
	Likes       []string `json:"likes"`     // список ID пользователей, поставивших лайк
	Status      string   `json:"status"`    // "draft", "published", "deleted"
}

// Repository хранит коллекцию тарифов (имитация БД)
type Repository struct {
	tariffs []Tariff
}

// NewRepository создаёт репозиторий с начальными данными
func NewRepository() *Repository {
	// Базовый URL для MinIO (убедитесь, что MinIO запущен и бакет создан)
	// В реальном приложении лучше вынести в конфиг, но по заданию конфига нет.
	baseURL := "http://localhost:9000/licensing/"

	tariffs := []Tariff{
		{
			ID:          1,
			Title:       "Basic S",
			Description: "Для небольших проектов и тестирования",
			Price:       3200,
			VCPU:        1,
			RAM:         1,
			Disk:        30,
			ImageURL:    baseURL + "basic_s.jpg",
			VideoURL:    baseURL + "basic_s.mp4",
			Likes:       []string{"user1"},
			Status:      "published",
		},
		{
			ID:          2,
			Title:       "Standard M",
			Description: "Для веб-приложений средней нагрузки",
			Price:       6400,
			VCPU:        2,
			RAM:         4,
			Disk:        30,
			ImageURL:    baseURL + "standard_m.jpg",
			VideoURL:    baseURL + "standard_m.mp4",
			Likes:       []string{"user1", "user2", "user3"},
			Status:      "published",
		},
		{
			ID:          3,
			Title:       "Pro L",
			Description: "Для высоконагруженных систем и баз данных",
			Price:       12400,
			VCPU:        4,
			RAM:         8,
			Disk:        50,
			ImageURL:    baseURL + "pro_l.jpg",
			VideoURL:    baseURL + "pro_l.mp4",
			Likes:       []string{"user2", "user4"},
			Status:      "published",
		},
		{
			ID:          4,
			Title:       "Enterprise XL",
			Description: "Для корпоративных решений и кластеров",
			Price:       24800,
			VCPU:        8,
			RAM:         16,
			Disk:        100,
			ImageURL:    baseURL + "enterprise_xl.jpg",
			VideoURL:    baseURL + "enterprise_xl.mp4",
			Likes:       []string{"user1", "user5", "user6", "user7"},
			Status:      "published",
		},
		{
			ID:          5,
			Title:       "Черновик",
			Description: "Новый тариф в разработке",
			Price:       500,
			VCPU:        1,
			RAM:         2,
			Disk:        20,
			ImageURL:    baseURL + "draft.jpg",
			VideoURL:    baseURL + "draft.mp4",
			Likes:       []string{},
			Status:      "draft",
		},
		{
			ID:          6,
			Title:       "Удалённый",
			Description: "Больше не доступен",
			Price:       0,
			VCPU:        0,
			RAM:         0,
			Disk:        0,
			ImageURL:    "",
			VideoURL:    "",
			Likes:       []string{},
			Status:      "deleted",
		},
	}

	return &Repository{tariffs: tariffs}
}

// GetPublished возвращает все опубликованные тарифы (статус "published")
func (r *Repository) GetPublished() []Tariff {
	var result []Tariff
	for _, t := range r.tariffs {
		if t.Status == "published" {
			result = append(result, t)
		}
	}
	return result
}

// GetDraft возвращает тариф в статусе "draft" (предполагается, что он один)
func (r *Repository) GetDraft() (Tariff, error) {
	for _, t := range r.tariffs {
		if t.Status == "draft" {
			return t, nil
		}
	}
	return Tariff{}, fmt.Errorf("черновик не найден")
}

// GetByID возвращает тариф по ID (независимо от статуса, но для ленты нужен опубликованный)
func (r *Repository) GetByID(id int) (Tariff, error) {
	for _, t := range r.tariffs {
		if t.ID == id {
			return t, nil
		}
	}
	return Tariff{}, fmt.Errorf("тариф с ID %d не найден", id)
}

// GetNext возвращает следующий опубликованный тариф после указанного ID (зацикливание)
func (r *Repository) GetNext(id int) (Tariff, error) {
	published := r.GetPublished()
	for i, t := range published {
		if t.ID == id {
			if i+1 < len(published) {
				return published[i+1], nil
			}
			// Если последний – возвращаем первый
			return published[0], nil
		}
	}
	return Tariff{}, fmt.Errorf("тариф с ID %d не найден среди опубликованных", id)
}

// FilterByPrice возвращает опубликованные тарифы с ценой >= minPrice
func (r *Repository) FilterByPrice(minPrice int) []Tariff {
	var result []Tariff
	for _, t := range r.GetPublished() {
		if t.Price >= minPrice {
			result = append(result, t)
		}
	}
	return result
}
