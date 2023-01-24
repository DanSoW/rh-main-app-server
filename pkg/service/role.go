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
func (s *RoleService) GetRole(column, value interface{}) (rbacModel.RoleModel, error) {
	return s.repo.GetRole(column, value)
}

/* Проверка существования у пользователя конкретной роли */
func (s *RoleService) HasRole(usersId, domainsId int, roleValue string) (bool, error) {
	return s.repo.HasRole(usersId, domainsId, roleValue)
}
