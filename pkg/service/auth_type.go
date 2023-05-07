package service

import (
	userModel "main-server/pkg/model/user"
	repository "main-server/pkg/repository"
)

type AuthTypeService struct {
	authType repository.AuthType
}

func NewAuthTypeService(role repository.AuthType) *AuthTypeService {
	return &AuthTypeService{authType: role}
}

func (s *AuthTypeService) Get(column string, value interface{}, check bool) (*userModel.AuthTypeModel, error) {
	return s.authType.Get(column, value, check)
}
