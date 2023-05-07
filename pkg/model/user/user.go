package user

type UserIdentityModel struct {
	UserId     int
	UserUuid   string
	DomainId   int
	DomainUuid string
}

/* A model for user uuid */
type UserUuidModel struct {
	Uuid string `json:"uuid" binding:"required" db:"uuid"`
}

/* A model for working with an instance of user data from the users table */
type UserModel struct {
	Id       int    `json:"id" db:"id"`
	Uuid     string `json:"uuid" binding:"required" db:"uuid"`
	Email    string `json:"email" binding:"required" db:"email"`
	Password string `json:"password" binding:"required" db:"password"`
}

/* Model for registration via Google OAuth 2 */
type UserRegisterOAuth2Model struct {
	Email      string `json:"email" binding:"required"`
	FamilyName string `json:"family_name" binding:"required"`
	GivenName  string `json:"given_name" binding:"required"`
	Name       string `json:"name" binding:"required"`
}

/* A model for working with data during user authorization (JSON parsing, etc.) */
type UserSignInModel struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

/* A model for working with data during user authorization via Google OAuth 2 */
type UserLoginOAuth2Model struct {
	Code string `json:"code" binding:"required"`
}

/* A model representing user authorization data */
type UserAuthDataModel struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

/* A model representing the user's activation data */
type UserActivateModel struct {
	ActivationLink string `json:"activation_link" db:"activation_link"`
	IsActivated    bool   `json:"is_activated" db:"is_activated"`
}

/* A model for representing authorization types */
type AuthTypeModel struct {
	Id    int    `json:"id" db:"id"`
	Uuid  string `json:"uuid" db:"uuid"`
	Value string `json:"value" db:"value"`
}

/* A model for linking users with specific types of authorizations */
type UserAuthTypeModel struct {
	Id          int `json:"id" db:"id"`
	UsersId     int `json:"users_id" db:"users_id"`
	AuthTypesId int `json:"auth_types_id" db:"auth_types_id"`
}

/* A model for email address every users */
type UserEmailModel struct {
	Email string `json:"email" db:"email"`
}

/* Структура для ролей пользователя*/
type UserRoleModel struct {
	Roles  []RoleContextModel `json:"roles" binding:"required"`
	Domain string             `json:"domain" binding:"required"`
}

/* Структура для конкретной роли пользователя */
type RoleContextModel struct {
	Name    string  `json:"name" binding:"required"` // Название роли
	Context *string `json:"context"`                 // Контекст, в котором действует роль (nil значит в рамках всей системы)
}
