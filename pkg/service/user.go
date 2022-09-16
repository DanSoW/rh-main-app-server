package service

import (
	userModel "main-server/pkg/model/user"
	repository "main-server/pkg/repository"

	"github.com/gin-gonic/gin"
)

/* Structure for this service */
type UserService struct {
	repo repository.User
}

/* Function for create new service */
func NewUserService(repo repository.User) *UserService {
	return &UserService{
		repo: repo,
	}
}

/* ********** */
/* Methods for profile */

/* Get information about profile user */
func (s *UserService) GetProfile(c *gin.Context) (userModel.UserProfileModel, error) {
	return s.repo.GetProfile(c)
}

func (s *UserService) UpdateProfile(c *gin.Context, data userModel.UserProfileUpdateDataModel) (userModel.UserJSONBModel, error) {
	return s.repo.UpdateProfile(c, data)
}

/* ********** */
