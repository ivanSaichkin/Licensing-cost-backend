package handler

import (
	"net/http"
	"strconv"
	"time"

	"licensing-cost/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repo *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repo: r}
}

func (h *Handler) GridLicenses(c *gin.Context) {
	priceParam := c.Query("max_price")
	var licenses []repository.License
	var err error

	if priceParam == "" {
		licenses, err = h.Repo.GetAllPublishedLicenses()
	} else {
		maxPrice, convErr := strconv.ParseFloat(priceParam, 64)
		if convErr != nil {
			logrus.Error("Неверный параметр max_price:", convErr)
			licenses, err = h.Repo.GetAllPublishedLicenses()
		} else {
			licenses, err = h.Repo.FilterLicensesByPrice(maxPrice)
		}
	}
	if err != nil {
		logrus.Error(err)
		licenses = []repository.License{}
	}

	var leftCol, rightCol []repository.License
	for i, l := range licenses {
		if i%2 == 0 {
			leftCol = append(leftCol, l)
		} else {
			rightCol = append(rightCol, l)
		}
	}

	c.HTML(http.StatusOK, "licensesGrid.html", gin.H{
		"time":      time.Now().Format("15:04:05"),
		"leftCol":   leftCol,
		"rightCol":  rightCol,
		"max_price": priceParam,
	})
}

func (h *Handler) FeedLicense(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("Неверный ID:", err)
		c.String(http.StatusBadRequest, "Неверный ID")
		return
	}

	if c.Query("next") == "true" {
		next, err := h.Repo.GetNextPublishedLicense(id)
		if err == nil {
			c.Redirect(http.StatusFound, "/feed/"+strconv.Itoa(next.ID))
			return
		}
	}

	license, err := h.Repo.GetLicenseByID(id)
	if err != nil {
		logrus.Error(err)
		c.String(http.StatusNotFound, "Лицензия не найдена")
		return
	}

	full := c.Query("full") == "true"
	shortDesc := license.Description
	if len(shortDesc) > 100 {
		shortDesc = shortDesc[:100] + "..."
	}

	c.HTML(http.StatusOK, "licensesFeed.html", gin.H{
		"license":   license,
		"likeCount": len(license.Likes),
		"shortDesc": shortDesc,
		"fullDesc":  license.Description,
		"showFull":  full,
	})
}

func (h *Handler) AddLicense(c *gin.Context) {
	draft, err := h.Repo.GetDraft()
	if err != nil {
		logrus.Error(err)
		c.String(http.StatusNotFound, "Черновик не найден")
		return
	}
	c.HTML(http.StatusOK, "addLicense.html", gin.H{
		"draft": draft,
	})
}
