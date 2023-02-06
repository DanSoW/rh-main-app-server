package service

import (
	"main-server/pkg/model/admin"
	adminModel "main-server/pkg/model/admin"
	userModel "main-server/pkg/model/user"
	repository "main-server/pkg/repository"

	"github.com/gin-gonic/gin"
)

/* Структура сервиса */
type AdminService struct {
	repo repository.Admin
}

/* Создание нового экземпляра структуры */
func NewAdminService(repo repository.Admin) *AdminService {
	return &AdminService{
		repo: repo,
	}
}

/* Получение списка всех пользователей */
func (s *AdminService) GetAllUsers(c *gin.Context) (admin.UsersResponseModel, error) {
	return s.repo.GetAllUsers(c)
}

/* Создание новой компании */
func (s *AdminService) CreateCompany(c *gin.Context, data adminModel.CompanyModel) (adminModel.CompanyModel, error) {
	return s.repo.CreateCompany(c, data)
}

/* Создание нового менеджера системы */
func (s *AdminService) SystemAddManager(user *userModel.UserIdentityModel, data adminModel.SystemPermissionModel) (bool, error) {
	return s.repo.SystemAddManager(user, data)
}
