package service

import (
	rbacModel "main-server/pkg/model/rbac"
	repository "main-server/pkg/repository"
)

/* Структура сервиса ролей */
type RoleService struct {
	repo repository.Role
}

/* Функция для создания нового сервиса ролей */
func NewRoleService(repo repository.Role) *RoleService {
	return &RoleService{
		repo: repo,
	}
}

/* Метод для получения роли */
func (s *RoleService) Get(column string, value interface{}, check bool) (*rbacModel.RoleModel, error) {
	return s.repo.Get(column, value, check)
}

/* Проверка существования у пользователя конкретной роли */
func (s *RoleService) HasRole(usersId, domainsId int, roleValue string) (bool, error) {
	return s.repo.HasRole(usersId, domainsId, roleValue)
}

/* Проверка существования у пользователя конкретной роли в рамках определённого субъекта */
func (s *RoleService) HasRoleWithSubject(userId, domainId int, roleValue, subjectId string) (bool, error) {
	return s.repo.HasRoleWithSubject(userId, domainId, roleValue, subjectId)
}
