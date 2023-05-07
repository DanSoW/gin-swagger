package service

import (
	emailModel "main-server/pkg/model/email"
	rbacModel "main-server/pkg/model/rbac"
	resourceModel "main-server/pkg/model/resource"
	userModel "main-server/pkg/model/user"
	repository "main-server/pkg/repository"

	"github.com/gin-gonic/gin"
)

type Authorization interface {
	CreateUser(user userModel.UserSignUpModel) (userModel.UserAuthDataModel, error)
	UploadProfileImage(c *gin.Context, filepath string) (bool, error)
	LoginUser(user userModel.UserSignInModel) (userModel.UserAuthDataModel, error)
	LoginUserOAuth2(code string) (userModel.UserAuthDataModel, error)
	Refresh(data userModel.TokenLogoutDataModel, refreshToken string) (userModel.UserAuthDataModel, error)
	Logout(tokens userModel.TokenLogoutDataModel) (bool, error)
	Activate(link string) (bool, error)
	RecoveryPassword(email string) (bool, error)
	ResetPassword(data userModel.ResetPasswordModel) (bool, error)
}

type Token interface {
	ParseToken(token, signingKey string) (userModel.TokenOutputParse, error)
	ParseTokenWithoutValid(token, signingKey string) (userModel.TokenOutputParse, error)
	ParseResetToken(pToken, signingKey string) (userModel.ResetTokenOutputParse, error)
}

type AuthType interface {
	Get(column string, value interface{}, check bool) (*userModel.AuthTypeModel, error)
}

type User interface {
	GetProfile(c *gin.Context) (userModel.UserProfileModel, error)
	UpdateProfile(c *gin.Context, data userModel.UserProfileUpdateDataModel) (userModel.UserDataDbModel, error)
	UpdateProfileImage(userIdentity *userModel.UserIdentityModel, resource *resourceModel.ImageModel) (*resourceModel.ImageModel, error)
	AccessCheck(userId, domainId int, value rbacModel.RoleValueModel) (bool, error)
	GetAllRoles(user userModel.UserIdentityModel) (*userModel.UserRoleModel, error)
}

type Domain interface {

	// CRUD
	Get(column string, value interface{}, check bool) (*rbacModel.DomainModel, error)
}

type Role interface {
	HasRole(usersId, domainsId int, roleValue string) (bool, error)
	HasRoleWithSubject(userId, domainId int, roleValue, subjectId string) (bool, error)

	// CRUD
	Get(column string, value interface{}, check bool) (*rbacModel.RoleModel, error)
}

type ServiceMain interface {
	SendEmail(user *userModel.UserIdentityModel, body *emailModel.MessageInputModel) (bool, error)
}

type Service struct {
	Authorization
	Token
	User
	Domain
	Role
	ServiceMain
}

func NewService(repos *repository.Repository) *Service {
	tokenService := NewTokenService(repos.Role, repos.User, repos.AuthType)

	return &Service{
		Token:         tokenService,
		Authorization: NewAuthService(repos.Authorization, *tokenService),
		User:          NewUserService(repos.User),
		Domain:        NewDomainService(repos.Domain),
		Role:          NewRoleService(repos.Role),
		ServiceMain:   NewServiceMainService(repos.ServiceMain),
	}
}
