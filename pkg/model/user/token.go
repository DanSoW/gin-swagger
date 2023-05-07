package user

type TokenModel struct {
	Id           int    `json:"id" db:"id"`
	UsersId      int    `json:"users_id" db:"users_id"`
	AccessToken  string `json:"access_token" db:"access_token"`
	RefreshToken string `json:"refresh_token" db:"refresh_token"`
}

type TokenDataModel struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenLogoutDataModel struct {
	AccessToken   string  `json:"access_token"`
	TokenApi      *string `json:"token_api"`
	RefreshToken  string  `json:"refresh_token"`
	AuthTypeValue string  `json:"auth_type_value"`
}

type TokenOutputParse struct {
	UsersId   int           `json:"users_id"`
	UsersUuid string        `json:"uuid"`
	AuthType  AuthTypeModel `json:"auth_types"`
	TokenApi  *string       `json:"token_api"`
}

type TokenOutputParseUU struct {
	UsersId  string        `json:"users_id"`
	AuthType AuthTypeModel `json:"auth_types"`
	TokenApi *string       `json:"token_api"`
}

type TokenOutputParseString struct {
	UsersId string `json:"users_id"`
}

type TokenRefreshModel struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenAccessModel struct {
	AccessToken string `json:"access_token" binding:"required"`
}

/* Reset token */
type ResetTokenOutputParse struct {
	UsersId int    `json:"users_id"`
	Email   string `json:"email"`
}

type ResetTokenModel struct {
	Id      int    `json:"id" db:"id"`
	UsersId int    `json:"users_id" db:"users_id"`
	Token   string `json:"token" db:"token"`
}
