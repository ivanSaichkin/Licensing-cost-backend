package ds

import "time"

type LicensingModel struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	Title             string     `gorm:"type:varchar(100);not null" json:"title"`
	Description       string     `gorm:"type:varchar(255)" json:"description"`
	LicenseType       string     `gorm:"type:varchar(30);not null" json:"license_type"`
	CommissionPerUnit float64    `gorm:"not null" json:"commission_per_unit"`
	MinForCalc        int        `gorm:"not null" json:"min_for_calc"`
	ImageURL          string     `gorm:"type:varchar(255)" json:"image_url"`
	VideoURL          string     `gorm:"type:varchar(255)" json:"video_url"`
	Status            string     `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
	CreatorID         uint       `gorm:"not null" json:"creator_id"`
	CreatedAt         time.Time  `gorm:"not null" json:"created_at"`
	PublishedAt       *time.Time `json:"published_at"`
	IsDeleted         bool       `gorm:"type:boolean;not null;default:false" json:"is_deleted"`

	// связи
	Creator User   `gorm:"foreignKey:CreatorID"`
	Likes   []Like `gorm:"foreignKey:LicensingID"`
}
