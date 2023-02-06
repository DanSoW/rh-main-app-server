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

func (s *ObjectService) GetObject(column, value interface{}) (*rbacModel.ObjectDbModel, error) {
	return s.repo.GetObject(column, value)
}

func (s *ObjectService) AddResource(resource *rbacModel.ResourceModel) (*rbacModel.ObjectDbModel, error) {
	return s.repo.AddResource(resource)
}
