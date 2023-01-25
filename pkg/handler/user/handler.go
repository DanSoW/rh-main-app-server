package user

import (
	_ "main-server/docs"

	middlewareConstant "main-server/pkg/constant/middleware"
	"main-server/pkg/constant/route"
	service "main-server/pkg/service"

	"github.com/gin-gonic/gin"
	_ "github.com/swaggo/files"
	_ "github.com/swaggo/gin-swagger"
)

type UserHandler struct {
	rootHandler *gin.Engine
	services    *service.Service
}

func NewUserHandler(root *gin.Engine, services *service.Service) *UserHandler {
	return &UserHandler{
		rootHandler: root,
		services:    services,
	}
}

/* Инициализация маршрутов для сервиса пользователей */
func (h *UserHandler) InitRoutes(
	middleware *map[string]func(c *gin.Context),
) {
	// URL: /user
	user := h.rootHandler.Group(route.USER_MAIN_ROUTE, (*middleware)[middlewareConstant.MN_UI])
	{
		// URL: /user/access/check
		user.POST(route.USER_CHECK_ACCESS_ROUTE, h.accessCheck)

		// URL: /user/role/get/all
		user.POST(route.USER_ROLES+"/"+route.GET_ALL_ROUTE, h.getUserRoles)

		// URL: /user/profile
		profile := user.Group(route.USER_PROFILE_ROUTE)
		{
			// URL: /user/profile/get
			profile.POST(route.GET_ROUTE, h.getProfile)

			// URL: /user/profile/update
			profile.POST(route.UPDATE_ROUTE, h.updateProfile)
		}

		// URL: /user/company
		company := user.Group(route.COMPANY_MAIN_ROUTE)
		{
			// URL: /user/company/get
			company.POST(route.GET_ROUTE, h.getUserCompany)
		}
	}
}
