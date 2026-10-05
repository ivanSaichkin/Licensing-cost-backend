package singleton

// currentUserID — заглушка создателя (до ЛР4)
const currentUserID uint = 1

func CurrentUserID() uint {
	return currentUserID
}
