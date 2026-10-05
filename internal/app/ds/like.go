package ds

type Like struct {
	ID          uint `gorm:"primaryKey" json:"id"`
	UserID      uint `gorm:"not null;uniqueIndex:idx_user_licensing" json:"user_id"`
	LicensingID uint `gorm:"not null;uniqueIndex:idx_user_licensing" json:"licensing_id"`

	User      User      `gorm:"foreignKey:UserID" json:"-"`
	Licensing Licensing `gorm:"foreignKey:LicensingID" json:"-"`
}
