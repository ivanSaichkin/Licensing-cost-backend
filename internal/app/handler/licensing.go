package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"licensing-cost/internal/app/ds"
)

const defaultImage = "/static/img/default_image.png"
const defaultVideo = "/static/img/default_video.mp4"

// GET /licensings
func (h *Handler) LicensingsGrid(ctx *gin.Context) {
	commissionParam := ctx.Query("max_commission")
	var licensings []ds.LicensingModel
	var err error

	if commissionParam == "" {
		licensings, err = h.Repository.GetAllPublishedLicensings()
	} else {
		maxComm, convErr := strconv.ParseFloat(commissionParam, 64)
		if convErr != nil {
			licensings, err = h.Repository.GetAllPublishedLicensings()
		} else {
			licensings, err = h.Repository.FilterLicensingsByCommission(maxComm)
		}
	}
	if err != nil {
		logrus.Error(err)
		licensings = []ds.LicensingModel{}
	}

	// Подсчёт лайков для каждой модели
	type Card struct {
		ds.LicensingModel
		LikesCount int64
	}
	var leftCol, rightCol []Card
	for i, l := range licensings {
		c := Card{LicensingModel: l, LikesCount: h.Repository.GetLikesCount(l.ID)}
		if i%2 == 0 {
			leftCol = append(leftCol, c)
		} else {
			rightCol = append(rightCol, c)
		}
	}

	ctx.HTML(http.StatusOK, "licensingsGrid.html", gin.H{
		"time":           time.Now().Format("15:04:05"),
		"leftCol":        leftCol,
		"rightCol":       rightCol,
		"max_commission": commissionParam,
	})
}

// GET /licensing/:id
func (h *Handler) LicensingFeed(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// ?next=true
	if ctx.Query("next") == "true" {
		next, err := h.Repository.GetNextPublishedLicensing(id)
		if err == nil && next != nil {
			ctx.Redirect(http.StatusFound, "/licensing/"+strconv.Itoa(int(next.ID)))
			return
		}
	}

	licensing, err := h.Repository.GetLicensingByID(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if licensing == nil {
		ctx.String(http.StatusNotFound, "Модель не найдена")
		return
	}

	// Если URL пустые — подставляем дефолтные
	imageURL := licensing.ImageURL
	if imageURL == "" {
		imageURL = defaultImage
	}
	videoURL := licensing.VideoURL
	if videoURL == "" {
		videoURL = defaultVideo
	}

	full := ctx.Query("full") == "true"
	shortDesc := licensing.Description
	if len(shortDesc) > 80 {
		shortDesc = shortDesc[:80] + "..."
	}

	ctx.HTML(http.StatusOK, "licensingsFeed.html", gin.H{
		"licensing": *licensing,
		"imageURL":  imageURL,
		"videoURL":  videoURL,
		"likeCount": h.Repository.GetLikesCount(licensing.ID),
		"shortDesc": shortDesc,
		"fullDesc":  licensing.Description,
		"showFull":  full,
	})
}

// GET /licensings/add
func (h *Handler) AddLicensing(ctx *gin.Context) {
	draft, err := h.Repository.GetDraft()
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Если черновика нет — отдаём пустую модель
	if draft == nil {
		draft = &ds.LicensingModel{}
	}

	ctx.HTML(http.StatusOK, "addLicensing.html", gin.H{
		"draft": *draft,
	})
}

// POST /licensings — создание через ORM
func (h *Handler) CreateLicensing(ctx *gin.Context) {
	title := ctx.PostForm("title")
	description := ctx.PostForm("description")
	licenseType := ctx.PostForm("license_type")
	imageURL := ctx.PostForm("image_url")
	videoURL := ctx.PostForm("video_url")
	commission, _ := strconv.ParseFloat(ctx.PostForm("commission"), 64)
	minForCalc, _ := strconv.Atoi(ctx.PostForm("min_for_calc"))

	licensing, err := h.Repository.CreateDraft(
		title, description, licenseType, imageURL, videoURL, commission, minForCalc,
	)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/licensings/add?id="+strconv.Itoa(int(licensing.ID)))
}

// POST /licensings/publish — публикация через ORM
func (h *Handler) PublishLicensing(ctx *gin.Context) {
	idStr := ctx.PostForm("licensing_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.PublishLicensing(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/licensings")
}

// POST /licensings/delete — удаление через SQL UPDATE
func (h *Handler) DeleteLicensing(ctx *gin.Context) {
	idStr := ctx.PostForm("licensing_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.DeleteLicensing(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/licensings")
}
