package ds

import "time"

type Licensing struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	Title             string     `gorm:"type:varchar(100);not null" json:"title"`
	Description       string     `gorm:"type:varchar(255)" json:"description"`
	CommissionPerUnit float64    `gorm:"not null" json:"commission_per_unit"`
	MinForCalc        int        `gorm:"not null" json:"min_for_calc"`
	ImageURL          string     `gorm:"type:varchar(255)" json:"image_url"`
	VideoURL          string     `gorm:"type:varchar(255)" json:"video_url"`
	Status            string     `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
	CreatorID         uint       `gorm:"not null" json:"creator_id"`
	CreatedAt         time.Time  `gorm:"not null" json:"created_at"`
	PublishedAt       *time.Time `json:"published_at"`

	Creator User   `gorm:"foreignKey:CreatorID"`
	Likes   []Like `gorm:"foreignKey:LicensingID"`
}
