package service

import (
	rbacModel "main-server/pkg/model/rbac"
	resourceModel "main-server/pkg/model/resource"
	userModel "main-server/pkg/model/user"
	repository "main-server/pkg/repository"

	"github.com/gin-gonic/gin"
)

/* Структура текущего файла */
type UserService struct {
	repo repository.User
}

/* Метод создания экземпляра структуры UserService */
func NewUserService(repo repository.User) *UserService {
	return &UserService{
		repo: repo,
	}
}

/* Получение информации о профиле пользователя */
func (s *UserService) GetProfile(c *gin.Context) (userModel.UserProfileModel, error) {
	return s.repo.GetProfile(c)
}

/* Обновление текстовых данных профиля пользователя*/
func (s *UserService) UpdateProfile(c *gin.Context, data userModel.UserProfileUpdateDataModel) (userModel.UserDataDbModel, error) {
	return s.repo.UpdateProfile(c, data)
}

/* Обновление изображения пользователя */
func (s *UserService) UpdateProfileImage(userIdentity *userModel.UserIdentityModel, resource *resourceModel.ImageModel) (*resourceModel.ImageModel, error) {
	return s.repo.UpdateProfileImage(userIdentity, resource)
}

/* Проверка доступа пользователя */
func (s *UserService) AccessCheck(userId, domainId int, value rbacModel.RoleValueModel) (bool, error) {
	return s.repo.AccessCheck(userId, domainId, value)
}

/* Получение всех ролей пользователя */
func (s *UserService) GetAllRoles(user userModel.UserIdentityModel) (*userModel.UserRoleModel, error) {
	return s.repo.GetAllRoles(user)
}
