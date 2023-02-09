package service

import (
	rbacModel "main-server/pkg/model/rbac"
	repository "main-server/pkg/repository"
)

type ObjectService struct {
	repo repository.Object
}

func NewObjectService(repo repository.Object) *ObjectService {
	return &ObjectService{
		repo: repo,
	}
}

func (s *ObjectService) Get(column string, value interface{}, check bool) (*rbacModel.ObjectDbModel, error) {
	return s.repo.Get(column, value, check)
}

func (s *ObjectService) AddResource(resource *rbacModel.ResourceModel) (*rbacModel.ObjectDbModel, error) {
	return s.repo.AddResource(resource)
}
