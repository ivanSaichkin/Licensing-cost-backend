package repository

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"

	"licensing-cost/internal/app/ds"
)

// --- GET список опубликованных с фильтром ---

func (r *Repository) GetPublishedLicensings(maxCommission float64) ([]ds.Licensing, error) {
	var licensings []ds.Licensing
	q := r.db.Where("status = ?", "published")
	if maxCommission > 0 {
		q = q.Where("commission_per_unit <= ?", maxCommission)
	}
	err := q.Order("id ASC").Find(&licensings).Error
	if err != nil {
		return nil, err
	}
	return licensings, nil
}

// --- GET лента (первая published) ---

func (r *Repository) GetFirstPublishedLicensing() (ds.Licensing, error) {
	var l ds.Licensing
	err := r.db.Where("status = ?", "published").
		Order("id ASC").First(&l).Error
	if err != nil {
		return ds.Licensing{}, err
	}
	return l, nil
}

func (r *Repository) GetPublishedLicensingByID(id int) (ds.Licensing, error) {
	var l ds.Licensing
	err := r.db.Where("id = ? AND status = ?", id, "published").First(&l).Error
	if err != nil {
		return ds.Licensing{}, err
	}
	return l, nil
}

func (r *Repository) GetNextPublishedLicensing(afterID int) (ds.Licensing, error) {
	var l ds.Licensing
	err := r.db.Where("id > ? AND status = ?", afterID, "published").
		Order("id ASC").First(&l).Error
	if err != nil {
		return ds.Licensing{}, err
	}
	return l, nil
}

// --- GET черновик ---

func (r *Repository) GetLicensingDraft(creatorID uint) (ds.Licensing, error) {
	var l ds.Licensing
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, "draft").
		First(&l).Error
	if err != nil {
		return ds.Licensing{}, err
	}
	return l, nil
}

// --- POST создание с файлами ---

type LicensingMediaFile struct {
	Header      *multipart.FileHeader
	ContentType string
	Filename    string
}

func (r *Repository) CreateLicensingDraft(
	creatorID uint,
	title, description string,
	commission float64,
	minForCalc int,
) (ds.Licensing, error) {
	l := ds.Licensing{
		Title:             title,
		Description:       description,
		CommissionPerUnit: commission,
		MinForCalc:        minForCalc,
		Status:            "draft",
		CreatorID:         creatorID,
		CreatedAt:         time.Now(),
	}
	if err := r.db.Create(&l).Error; err != nil {
		return ds.Licensing{}, err
	}
	return l, nil
}

func (r *Repository) AddLicensingMedia(l *ds.Licensing, image, video LicensingMediaFile) error {
	if image.Filename != "" {
		if err := r.uploadLicensingMedia(image); err != nil {
			return err
		}
		l.ImageURL = r.buildLicensingMediaURL(image.Filename)
	}
	if video.Filename != "" {
		if err := r.uploadLicensingMedia(video); err != nil {
			return err
		}
		l.VideoURL = r.buildLicensingMediaURL(video.Filename)
	}
	return r.db.Model(l).Updates(map[string]interface{}{
		"image_url": l.ImageURL,
		"video_url": l.VideoURL,
	}).Error
}

func (r *Repository) uploadLicensingMedia(media LicensingMediaFile) error {
	file, err := media.Header.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = r.minio.PutObject(
		context.Background(),
		r.minioBucketName,
		media.Filename,
		file,
		media.Header.Size,
		minio.PutObjectOptions{ContentType: media.ContentType},
	)
	return err
}

func (r *Repository) buildLicensingMediaURL(filename string) string {
	return fmt.Sprintf("http://localhost:9000/%s/%s", r.minioBucketName, filename)
}

// --- PUT публикация ---

func (r *Repository) PublishLicensing(id int, creatorID uint) error {
	result := r.db.Model(&ds.Licensing{}).
		Where("id = ? AND creator_id = ? AND status = ?", id, creatorID, "draft").
		Updates(map[string]interface{}{
			"status":       "published",
			"published_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// --- DELETE soft delete ---

func (r *Repository) DeleteLicensing(id int, creatorID uint) error {
	result := r.db.Exec(
		"UPDATE licensings SET status = 'deleted' WHERE id = $1 AND creator_id = $2 AND status != 'deleted'",
		id, creatorID,
	)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// --- Лайки ---

func (r *Repository) LikeLicensing(userID, licensingID uint) error {
	like := ds.Like{UserID: userID, LicensingID: licensingID}
	return r.db.Create(&like).Error
}

func (r *Repository) UnlikeLicensing(userID, licensingID uint) error {
	return r.db.Where("user_id = ? AND licensing_id = ?", userID, licensingID).
		Delete(&ds.Like{}).Error
}

func (r *Repository) IsLicensingLiked(userID, licensingID uint) (bool, error) {
	var count int64
	err := r.db.Model(&ds.Like{}).
		Where("user_id = ? AND licensing_id = ?", userID, licensingID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) GetLicensingLikesCount(id uint) (int64, error) {
	var count int64
	err := r.db.Model(&ds.Like{}).Where("licensing_id = ?", id).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) GetLicensingsLikesCounts(ids []uint) (map[uint]int64, error) {
	type row struct {
		LicensingID uint
		Count       int64
	}
	var rows []row
	if len(ids) == 0 {
		return map[uint]int64{}, nil
	}
	err := r.db.Model(&ds.Like{}).
		Select("licensing_id, count(*) as count").
		Where("licensing_id IN ?", ids).
		Group("licensing_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]int64, len(rows))
	for _, r := range rows {
		result[r.LicensingID] = r.Count
	}
	return result, nil
}

// --- MinIO ---

func (r *Repository) UploadFile(objectName string, reader interface {
	Read([]byte) (int, error)
}, size int64, contentType string) error {
	return nil
}

func (r *Repository) GetMinioBucketName() string {
	return r.minioBucketName
}
