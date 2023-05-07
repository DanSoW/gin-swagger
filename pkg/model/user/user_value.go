package user

const (
	AUTHORIZATION_HEADER = "Authorization"
	USER_CTX             = "users_id"
	ROLES_CTX            = "roles_id"
	AUTH_TYPE_VALUE_CTX  = "auth_type_value"
	ACCESS_TOKEN_CTX     = "access_token"
	TOKEN_API_CTX        = "token_api"
)

/* User Identification Data model */
type UserValueModel struct {
	UsersId int
	RolesId int
}
