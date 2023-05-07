package service

import (
	rbacModel "main-server/pkg/model/rbac"
	repository "main-server/pkg/repository"
)

type DomainService struct {
	repo repository.Domain
}

func NewDomainService(repo repository.Domain) *DomainService {
	return &DomainService{
		repo: repo,
	}
}

func (s *DomainService) Get(column string, value interface{}, check bool) (*rbacModel.DomainModel, error) {
	return s.repo.Get(column, value, check)
}
