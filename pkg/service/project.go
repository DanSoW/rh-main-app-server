package service

import (
	projectModel "main-server/pkg/model/project"
	repository "main-server/pkg/repository"
)

/* Structure for this service */
type ProjectService struct {
	repo repository.Project
}

/* Function for create new struct of ProjectService */
func NewProjectService(repo repository.Project) *ProjectService {
	return &ProjectService{
		repo: repo,
	}
}

/* Method for create new project */
func (s *ProjectService) CreateProject(userId, domainId int, data projectModel.ProjectModel) (projectModel.ProjectModel, error) {
	return s.repo.CreateProject(userId, domainId, data)
}

/* Method for add logo to project */
func (s *ProjectService) AddLogoProject(userId, domainId int, data projectModel.ProjectLogoModel) (projectModel.ProjectLogoModel, error) {
	return s.repo.AddLogoProject(userId, domainId, data)
}
