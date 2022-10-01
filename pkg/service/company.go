package service

import (
	companyModel "main-server/pkg/model/company"
	repository "main-server/pkg/repository"
)

/* Structure for this service */
type CompanyService struct {
	repo repository.Company
}

/* Function for create new struct of CompanyService */
func NewCompanyService(repo repository.Company) *CompanyService {
	return &CompanyService{
		repo: repo,
	}
}

/* Method for create new project */
func (s *CompanyService) GetManagers(userId, domainId int, data companyModel.ManagerCountModel) (companyModel.ManagerAnyCountModel, error) {
	return s.repo.GetManagers(userId, domainId, data)
}
