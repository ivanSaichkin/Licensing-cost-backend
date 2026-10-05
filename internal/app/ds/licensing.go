package ds

import "time"

type Licensing struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	Title             string     `gorm:"type:varchar(100);not null" json:"title"`
	Description       string     `gorm:"type:varchar(255)" json:"description"`
	CommissionPerUnit float64    `gorm:"not null" json:"commission_per_unit"`
	MinForCalc        int        `gorm:"not null" json:"min_for_calc"`
	ImageURL          string     `gorm:"type:varchar(255);not null;default:''" json:"image_url"`
	VideoURL          string     `gorm:"type:varchar(255);not null;default:''" json:"video_url"`
	Status            string     `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
	CreatorID         uint       `gorm:"not null" json:"creator_id"`
	CreatedAt         time.Time  `gorm:"not null" json:"created_at"`
	PublishedAt       *time.Time `json:"published_at"`

	Creator User   `gorm:"foreignKey:CreatorID" json:"-"`
	Likes   []Like `gorm:"foreignKey:LicensingID" json:"-"`
}

// --- Сериализаторы ---

// LicensingListSerializer — для GET /api/licensings
type LicensingListSerializer struct {
	Licensing
	IsCreator  int   `json:"is_creator"` // 0/1
	LikesCount int64 `json:"likes_count"`
}

func NewLicensingListSerializer(l Licensing, likesCount int64, currentUserID uint) LicensingListSerializer {
	isCreator := 0
	if l.CreatorID == currentUserID {
		isCreator = 1
	}
	return LicensingListSerializer{
		Licensing:  l,
		IsCreator:  isCreator,
		LikesCount: likesCount,
	}
}

// LicensingFeedSerializer — для GET /api/licensing/feed
type LicensingFeedSerializer struct {
	Licensing
	IsLiked    int   `json:"is_liked"` // 0/1
	LikesCount int64 `json:"likes_count"`
}

func NewLicensingFeedSerializer(l Licensing, likesCount int64, isLiked bool) LicensingFeedSerializer {
	liked := 0
	if isLiked {
		liked = 1
	}
	return LicensingFeedSerializer{
		Licensing:  l,
		IsLiked:    liked,
		LikesCount: likesCount,
	}
}

// LicensingDraftSerializer — для GET /api/licensings/draft
type LicensingDraftSerializer struct {
	Licensing
}

func NewLicensingDraftSerializer(l Licensing) LicensingDraftSerializer {
	return LicensingDraftSerializer{Licensing: l}
}

// LicensingLikeRequest — тело POST /api/licensings/:id/like
type LicensingLikeRequest struct {
	Like *int `json:"like" binding:"required,oneof=0 1"`
}
