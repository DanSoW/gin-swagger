package user

/* Model for data profile */
type UserProfileModel struct {
	Email string `json:"email" binding:"required"`
	Data  string `json:"data" binding:"required" db:"data"`
}

type UserDataModel struct {
	Data string `json:"data" binding:"required" db:"data"`
}

/* Model for request profile user */
type UserProfileDataModel struct {
	Email      string `json:"email"`
	Name       string `json:"name" binding:"required"`
	Surname    string `json:"surname" binding:"required"`
	Nickname   string `json:"nickname" binding:"required"`
	Patronymic string `json:"patronymic"`
	Position   string `json:"position"`
	Avatar     string `json:"avatar"`
}

/* Model for request update profile user */
type UserProfileUpdateDataModel struct {
	Name       string  `json:"name" binding:"required"`
	Surname    string  `json:"surname" binding:"required"`
	Nickname   string  `json:"nickname" binding:"required"`
	Patronymic string  `json:"patronymic"`
	Position   string  `json:"position"`
	Password   *string `json:"password"`
}
