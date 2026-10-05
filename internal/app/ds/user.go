package ds

type User struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Login       string `gorm:"type:varchar(50);unique;not null" json:"login"`
	Password    string `gorm:"type:varchar(100);not null" json:"-"`
	IsModerator bool   `gorm:"type:boolean;default:false" json:"is_moderator"`
}

// RegisterRequest — тело POST /api/users/register
type RegisterRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginRequest — тело POST /api/users/login (заглушка)
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
