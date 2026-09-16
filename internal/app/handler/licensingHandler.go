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

// GET /licensings — плитка с фильтром по комиссии (слайдер)
func (h *Handler) LicensingsGrid(c *gin.Context) {
	commissionParam := c.Query("max_commission")
	var licensings []repository.LicensingModel
	var err error

	if commissionParam == "" {
		licensings, err = h.Repo.GetAllPublishedLicensings()
	} else {
		maxComm, convErr := strconv.ParseFloat(commissionParam, 64)
		if convErr != nil {
			logrus.Error("Неверный параметр max_commission:", convErr)
			licensings, err = h.Repo.GetAllPublishedLicensings()
		} else {
			licensings, err = h.Repo.FilterLicensingsByCommission(maxComm)
		}
	}
	if err != nil {
		logrus.Error(err)
		licensings = []repository.LicensingModel{}
	}

	var leftCol, rightCol []repository.LicensingModel
	for i, l := range licensings {
		if i%2 == 0 {
			leftCol = append(leftCol, l)
		} else {
			rightCol = append(rightCol, l)
		}
	}

	c.HTML(http.StatusOK, "licensingsGrid.html", gin.H{
		"time":           time.Now().Format("15:04:05"),
		"leftCol":        leftCol,
		"rightCol":       rightCol,
		"max_commission": commissionParam,
	})
}

// GET /licensing/:id — лента
func (h *Handler) LicensingFeed(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("Неверный ID:", err)
		c.String(http.StatusBadRequest, "Неверный ID")
		return
	}

	if c.Query("next") == "true" {
		next, err := h.Repo.GetNextPublishedLicensing(id)
		if err == nil {
			c.Redirect(http.StatusFound, "/licensing/"+strconv.Itoa(next.ID))
			return
		}
	}

	licensing, err := h.Repo.GetLicensingByID(id)
	if err != nil {
		logrus.Error(err)
		c.String(http.StatusNotFound, "Модель не найдена")
		return
	}

	full := c.Query("full") == "true"
	// Короткое описание — примерно 2 строки (≈80 символов)
	shortDesc := licensing.Description
	if len(shortDesc) > 80 {
		shortDesc = shortDesc[:80] + "..."
	}

	c.HTML(http.StatusOK, "licensingsFeed.html", gin.H{
		"licensing": licensing,
		"likeCount": len(licensing.Likes),
		"shortDesc": shortDesc,
		"fullDesc":  licensing.Description,
		"showFull":  full,
	})
}

// GET /licensings/add — страница добавления
func (h *Handler) AddLicensing(c *gin.Context) {
	draft, err := h.Repo.GetDraft()
	if err != nil {
		logrus.Error(err)
		c.String(http.StatusNotFound, "Черновик не найден")
		return
	}
	c.HTML(http.StatusOK, "addLicensing.html", gin.H{
		"draft": draft,
	})
}
