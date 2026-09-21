package ds

type Like struct {
	ID          uint `gorm:"primaryKey"`
	UserID      uint `gorm:"not null;uniqueIndex:idx_user_licensing"`
	LicensingID uint `gorm:"not null;uniqueIndex:idx_user_licensing"`

	User      User           `gorm:"foreignKey:UserID"`
	Licensing LicensingModel `gorm:"foreignKey:LicensingID"`
}
