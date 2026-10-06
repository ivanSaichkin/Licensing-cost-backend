package handler

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"licensing-cost/internal/app/ds"
	"licensing-cost/internal/app/repository"
)

var licensingImageExtensions = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

var licensingVideoExtensions = map[string]string{
	"video/webm": ".webm",
	"video/mp4":  ".mp4",
}

// GET /api/licensings?max_commission=
func (h *Handler) GetLicensingsAPI(ctx *gin.Context) {
	var licensings []ds.Licensing
	var err error

	maxCommission, _ := strconv.ParseFloat(ctx.Query("max_commission"), 64)
	licensings, err = h.Repository.GetPublishedLicensings(maxCommission)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ids := make([]uint, 0, len(licensings))
	for _, l := range licensings {
		ids = append(ids, l.ID)
	}
	likesCounts, err := h.Repository.GetLicensingsLikesCounts(ids)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	serializers := make([]ds.LicensingListSerializer, 0, len(licensings))
	for _, l := range licensings {
		serializers = append(
			serializers,
			ds.NewLicensingListSerializer(l, likesCounts[l.ID], CurrentUserID()),
		)
	}

	ctx.JSON(http.StatusOK, serializers)
}

// GET /api/licensing/feed
func (h *Handler) GetLicensingFeedAPI(ctx *gin.Context) {
	var licensing ds.Licensing
	var err error

	idStr := ctx.Query("id")
	nextStr := ctx.Query("next")

	switch {
	case idStr == "":
		// первая
		licensing, err = h.Repository.GetFirstPublishedLicensing()
	case nextStr == "true":
		// следующая
		id, convErr := strconv.Atoi(idStr)
		if convErr != nil {
			h.errorHandler(ctx, http.StatusBadRequest, convErr)
			return
		}
		licensing, err = h.Repository.GetNextPublishedLicensing(id)
	default:
		// конкретная по id
		id, convErr := strconv.Atoi(idStr)
		if convErr != nil {
			h.errorHandler(ctx, http.StatusBadRequest, convErr)
			return
		}
		licensing, err = h.Repository.GetPublishedLicensingByID(id)
	}

	if err != nil {
		h.repositoryErrorHandler(ctx, err)
		return
	}

	h.respondLicensing(ctx, http.StatusOK, licensing)
}

// GET /api/licensings/draft
func (h *Handler) GetLicensingDraftAPI(ctx *gin.Context) {
	draft, err := h.Repository.GetLicensingDraft(CurrentUserID())
	if err != nil {
		h.repositoryErrorHandler(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, ds.NewLicensingDraftSerializer(draft))
}

// POST /api/licensings
func (h *Handler) CreateLicensingAPI(ctx *gin.Context) {
	if err := ctx.Request.ParseMultipartForm(50 << 20); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	_, err := h.Repository.GetLicensingDraft(CurrentUserID())
	if err == nil {
		h.errorHandler(ctx, http.StatusConflict, errors.New("черновик уже существует"))
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	title := ctx.Request.FormValue("title")
	if title == "" {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("не указано название"))
		return
	}

	imageHeader, err := ctx.FormFile("image")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	videoHeader, err := ctx.FormFile("video")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	imageContentType, imageExtension, err := detectLicensingMediaType(imageHeader, licensingImageExtensions)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	videoContentType, videoExtension, err := detectLicensingMediaType(videoHeader, licensingVideoExtensions)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	commission, _ := strconv.ParseFloat(ctx.Request.FormValue("commission_per_unit"), 64)
	minForCalc, _ := strconv.Atoi(ctx.Request.FormValue("min_for_calc"))

	licensing, err := h.Repository.CreateLicensingDraft(
		CurrentUserID(),
		title,
		ctx.Request.FormValue("description"),
		commission,
		minForCalc,
	)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	err = h.Repository.AddLicensingMedia(
		&licensing,
		repository.LicensingMediaFile{
			Header:      imageHeader,
			ContentType: imageContentType,
			Filename:    fmt.Sprintf("licensing-%d-image%s", licensing.ID, imageExtension),
		},
		repository.LicensingMediaFile{
			Header:      videoHeader,
			ContentType: videoContentType,
			Filename:    fmt.Sprintf("licensing-%d-video%s", licensing.ID, videoExtension),
		},
	)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.respondLicensing(ctx, http.StatusCreated, licensing)
}

// PUT /api/licensings/:id/publish
func (h *Handler) PublishLicensingAPI(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.PublishLicensing(id, CurrentUserID()); err != nil {
		h.repositoryErrorHandler(ctx, err)
		return
	}

	licensing, err := h.Repository.GetPublishedLicensingByID(id)
	if err != nil {
		h.repositoryErrorHandler(ctx, err)
		return
	}

	h.respondLicensing(ctx, http.StatusOK, licensing)
}

// DELETE /api/licensings/:id
func (h *Handler) DeleteLicensingAPI(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeleteLicensing(id, CurrentUserID()); err != nil {
		h.repositoryErrorHandler(ctx, err)
		return
	}
	ctx.Status(http.StatusOK)
}

// POST /api/licensings/:id/like
func (h *Handler) LikeLicensingAPI(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var request ds.LicensingLikeRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	licensing, err := h.Repository.GetPublishedLicensingByID(id)
	if err != nil {
		h.repositoryErrorHandler(ctx, err)
		return
	}

	if *request.Like == 1 {
		err = h.Repository.LikeLicensing(CurrentUserID(), licensing.ID)
	} else {
		err = h.Repository.UnlikeLicensing(CurrentUserID(), licensing.ID)
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.respondLicensing(ctx, http.StatusOK, licensing)
}

// respondLicensing — сериализация + likes
func (h *Handler) respondLicensing(ctx *gin.Context, statusCode int, l ds.Licensing) {
	likesCount, err := h.Repository.GetLicensingLikesCount(l.ID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	isLiked, err := h.Repository.IsLicensingLiked(CurrentUserID(), l.ID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(statusCode, ds.NewLicensingFeedSerializer(l, likesCount, isLiked))
}

func (h *Handler) repositoryErrorHandler(ctx *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}
	h.errorHandler(ctx, http.StatusInternalServerError, err)
}

func detectLicensingMediaType(
	header *multipart.FileHeader,
	extensions map[string]string,
) (string, string, error) {
	file, err := header.Open()
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return "", "", err
	}

	contentType := http.DetectContentType(buffer[:n])
	extension, ok := extensions[contentType]
	if !ok {
		return "", "", fmt.Errorf("недопустимый тип файла: %s", contentType)
	}
	return contentType, extension, nil
}
