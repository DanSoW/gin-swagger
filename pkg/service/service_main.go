package service

import (
	emailModel "main-server/pkg/model/email"
	userModel "main-server/pkg/model/user"
	repository "main-server/pkg/repository"
)

type ServiceMainService struct {
	repo repository.ServiceMain
}

func NewServiceMainService(repo repository.ServiceMain) *ServiceMainService {
	return &ServiceMainService{
		repo: repo,
	}
}

func (s *ServiceMainService) SendEmail(user *userModel.UserIdentityModel, body *emailModel.MessageInputModel) (bool, error) {
	return s.repo.SendEmail(user, body)
}
