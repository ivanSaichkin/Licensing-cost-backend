package handler

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

// GetLicensings godoc
// @Summary      Список моделей
// @Tags         licensings
// @Produce      json
// @Param        max_commission  query   number  false  "Макс. комиссия"
// @Success      200  {array}   ds.LicensingListSerializer
// @Router       /licensings [get]
func (h *Handler) GetLicensings(ctx *gin.Context) {
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

	userID := CurrentUserID(ctx)
	serializers := make([]ds.LicensingListSerializer, 0, len(licensings))
	for _, l := range licensings {
		serializers = append(serializers,
			ds.NewLicensingListSerializer(l, likesCounts[l.ID], userID),
		)
	}

	ctx.JSON(http.StatusOK, serializers)
}

// GetLicensingReel godoc
// @Summary      Лента
// @Tags         licensings
// @Produce      json
// @Param        id    query     int   false  "ID"
// @Param        next  query     bool  false  "Следующая"
// @Success      200   {object}  ds.LicensingFeedSerializer
// @Failure      404   {string}  string  "Not found"
// @Router       /licensing/feed [get]
func (h *Handler) GetLicensingReel(ctx *gin.Context) {
	var licensing ds.Licensing
	var err error

	idStr := ctx.Query("id")
	switch {
	case idStr == "":
		licensing, err = h.Repository.GetFirstPublishedLicensing()
	case ctx.Query("next") == "true":
		id, convErr := strconv.Atoi(idStr)
		if convErr != nil {
			h.errorHandler(ctx, http.StatusBadRequest, convErr)
			return
		}
		licensing, err = h.Repository.GetNextPublishedLicensing(id)
	default:
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

// GetLicensingDraft godoc
// @Summary      Черновик
// @Tags         licensings
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  ds.LicensingDraftSerializer
// @Failure      403  {string}  string  "Forbidden"
// @Failure      404  {string}  string  "Not found"
// @Router       /licensings/draft [get]
func (h *Handler) GetLicensingDraft(ctx *gin.Context) {
	draft, err := h.Repository.GetLicensingDraft(CurrentUserID(ctx))
	if err != nil {
		h.repositoryErrorHandler(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, ds.NewLicensingDraftSerializer(draft))
}

// CreateLicensing godoc
// @Summary      Создать черновик
// @Tags         licensings
// @Accept       mpfd
// @Produce      json
// @Security     Bearer
// @Param        title        formData  string  true   "Название"
// @Param        description  formData  string  false  "Описание"
// @Param        image        formData  file    true   "Изображение"
// @Param        video        formData  file    true   "Видео"
// @Success      201  {object}  ds.LicensingFeedSerializer
// @Failure      409  {string}  string  "Conflict"
// @Router       /licensings [post]
func (h *Handler) CreateLicensing(ctx *gin.Context) {
	if err := ctx.Request.ParseMultipartForm(50 << 20); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	userID := CurrentUserID(ctx)
	logrus.Infof("CreateLicensing: userID=%d", userID)
	_, err := h.Repository.GetLicensingDraft(userID)
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

	licensing, err := h.Repository.CreateLicensingDraft(
		userID, title, ctx.Request.FormValue("description"),
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

// PublishLicensing godoc
// @Summary      Публикация
// @Tags         licensings
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id       path      int                         true  "ID"
// @Param        request  body      ds.PublishLicensingRequest  true  "Данные"
// @Success      200  {object}  ds.LicensingFeedSerializer
// @Failure      404  {string}  string  "Not found"
// @Router       /licensings/{id}/publish [put]
func (h *Handler) PublishLicensing(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var request ds.PublishLicensingRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.PublishLicensing(id, CurrentUserID(ctx), request); err != nil {
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

// DeleteLicensing godoc
// @Summary      Удаление (soft)
// @Tags         licensings
// @Produce      json
// @Security     Bearer
// @Param        id  path  int  true  "ID"
// @Success      200  {string}  string  "OK"
// @Failure      404  {string}  string  "Not found"
// @Router       /licensings/{id} [delete]
func (h *Handler) DeleteLicensing(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeleteLicensing(id, CurrentUserID(ctx)); err != nil {
		h.repositoryErrorHandler(ctx, err)
		return
	}
	ctx.Status(http.StatusOK)
}

// LikeLicensing godoc
// @Summary      Лайк
// @Tags         licensings
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id       path      int                     true  "ID"
// @Param        request  body      ds.LicensingLikeRequest true  "0 или 1"
// @Success      200  {object}  ds.LicensingFeedSerializer
// @Failure      404  {string}  string  "Not found"
// @Router       /licensings/{id}/like [post]
func (h *Handler) LikeLicensing(ctx *gin.Context) {
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

	userID := CurrentUserID(ctx)
	if *request.Like == 1 {
		err = h.Repository.LikeLicensing(userID, licensing.ID)
	} else {
		err = h.Repository.UnlikeLicensing(userID, licensing.ID)
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.respondLicensing(ctx, http.StatusOK, licensing)
}

func (h *Handler) respondLicensing(ctx *gin.Context, statusCode int, l ds.Licensing) {
	likesCount, err := h.Repository.GetLicensingLikesCount(l.ID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	userID := CurrentUserID(ctx)
	isLiked, err := h.Repository.IsLicensingLiked(userID, l.ID)
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
