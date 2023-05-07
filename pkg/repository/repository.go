package repository

import (
	emailModel "main-server/pkg/model/email"
	rbacModel "main-server/pkg/model/rbac"
	resourceModel "main-server/pkg/model/resource"
	userModel "main-server/pkg/model/user"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
)

type Authorization interface {
	CreateUser(user userModel.UserSignUpModel) (userModel.UserAuthDataModel, error)
	UploadProfileImage(c *gin.Context, filepath string) (bool, error)
	LoginUser(user userModel.UserSignInModel) (userModel.UserAuthDataModel, error)
	LoginUserOAuth2(code string) (userModel.UserAuthDataModel, error)
	CreateUserOAuth2(user userModel.UserRegisterOAuth2Model, token *oauth2.Token) (userModel.UserAuthDataModel, error)
	Refresh(data userModel.TokenLogoutDataModel, refreshToken string, token userModel.TokenOutputParse) (userModel.UserAuthDataModel, error)
	Logout(tokens userModel.TokenLogoutDataModel) (bool, error)
	Activate(link string) (bool, error)
	GetUser(column, value string) (userModel.UserModel, error)
	GetRole(column, value string) (rbacModel.RoleModel, error)
	RecoveryPassword(email string) (bool, error)
	ResetPassword(data userModel.ResetPasswordModel, token userModel.ResetTokenOutputParse) (bool, error)
}

type Role interface {
	HasRole(usersId, domainsId int, roleValue string) (bool, error)
	HasRoleWithSubject(userId, domainId int, roleValue, subjectId string) (bool, error)
	Get(column string, value interface{}, check bool) (*rbacModel.RoleModel, error)
}

type Domain interface {
	Get(column string, value interface{}, check bool) (*rbacModel.DomainModel, error)
}

type User interface {
	GetProfile(c *gin.Context) (userModel.UserProfileModel, error)
	UpdateProfile(c *gin.Context, data userModel.UserProfileUpdateDataModel) (userModel.UserDataDbModel, error)
	UpdateProfileImage(userIdentity *userModel.UserIdentityModel, resource *resourceModel.ImageModel) (*resourceModel.ImageModel, error)
	AccessCheck(userId, domainId int, value rbacModel.RoleValueModel) (bool, error)
	GetAllRoles(user userModel.UserIdentityModel) (*userModel.UserRoleModel, error)
	Get(column string, value interface{}, check bool) (*userModel.UserModel, error)
}

type AuthType interface {
	Get(column string, value interface{}, check bool) (*userModel.AuthTypeModel, error)
}

type ServiceMain interface {
	SendEmail(*userModel.UserIdentityModel, *emailModel.MessageInputModel) (bool, error)
}

type Repository struct {
	Authorization
	Role
	Domain
	User
	AuthType
	ServiceMain
}

/* Создание нового экземпляра глобального репозитория */
func NewRepository(db *sqlx.DB, enforcer *casbin.Enforcer) *Repository {

	role := NewRolePostgres(db, enforcer)
	domain := NewDomainPostgres(db)
	user := NewUserPostgres(db, enforcer, domain, role)
	serviceMain := NewServiceMainRepository(db, enforcer, user)

	return &Repository{
		Authorization: NewAuthPostgres(db, enforcer, *user),
		Role:          role,
		Domain:        domain,
		User:          user,
		AuthType:      NewAuthTypePostgres(db),
		ServiceMain:   serviceMain,
	}
}
