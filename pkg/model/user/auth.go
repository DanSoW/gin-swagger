package user

type GoogleOAuth2Code struct {
	Code string `json:"code" binding:"required"`
}

type ResetPasswordModel struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required"`
}
