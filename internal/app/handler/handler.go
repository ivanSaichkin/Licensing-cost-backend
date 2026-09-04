package handler

import (
	"licensing-cost/internal/app/repository"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Repo *repository.Repository
}

func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{Repo: repo}
}

// FeedHandler — страница ленты (отображает один тариф)
// Параметры: id (опционально), next=true (следующий)
// Если id не указан – показывает первый опубликованный
func (h *Handler) FeedHandler(c *gin.Context) {
	idStr := c.Query("id")
	next := c.Query("next") == "true"

	var tariff repository.Tariff
	var err error

	if idStr != "" {
		id, errConv := strconv.Atoi(idStr)
		if errConv != nil {
			c.String(http.StatusBadRequest, "Неверный ID")
			return
		}
		if next {
			tariff, err = h.Repo.GetNext(id)
		} else {
			tariff, err = h.Repo.GetByID(id)
		}
	} else {
		// Без id – первый опубликованный
		published := h.Repo.GetPublished()
		if len(published) > 0 {
			tariff = published[0]
		} else {
			c.String(http.StatusNotFound, "Нет опубликованных тарифов")
			return
		}
	}

	if err != nil {
		log.Println("Ошибка получения тарифа:", err)
		c.String(http.StatusNotFound, "Тариф не найден")
		return
	}

	c.HTML(http.StatusOK, "feed.html", gin.H{
		"tariff": tariff,
		"time":   time.Now().Format("15:04:05"),
	})
}

// AddHandler — страница добавления (показывает черновик)
func (h *Handler) AddHandler(c *gin.Context) {
	draft, err := h.Repo.GetDraft()
	if err != nil {
		// Если черновика нет, передаём пустой тариф (показываем пустую форму)
		draft = repository.Tariff{}
	}
	c.HTML(http.StatusOK, "add.html", gin.H{
		"draft": draft,
	})
}

// GridHandler — страница плитки (список всех опубликованных тарифов с фильтром по цене)
// Параметр: min_price (необязательный)
func (h *Handler) GridHandler(c *gin.Context) {
	minPriceStr := c.Query("min_price")
	var tariffs []repository.Tariff

	if minPriceStr != "" {
		minPrice, err := strconv.Atoi(minPriceStr)
		if err != nil {
			// Если не число – игнорируем фильтр
			tariffs = h.Repo.GetPublished()
		} else {
			tariffs = h.Repo.FilterByPrice(minPrice)
		}
	} else {
		tariffs = h.Repo.GetPublished()
	}

	c.HTML(http.StatusOK, "grid.html", gin.H{
		"tariffs":  tariffs,
		"minPrice": minPriceStr,
		"time":     time.Now().Format("15:04:05"),
	})
}
