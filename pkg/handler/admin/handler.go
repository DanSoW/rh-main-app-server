package admin

import (
	_ "main-server/docs"

	middlewareConstant "main-server/pkg/constant/middleware"
	"main-server/pkg/constant/route"
	service "main-server/pkg/service"

	"github.com/gin-gonic/gin"
	_ "github.com/swaggo/files"
	_ "github.com/swaggo/gin-swagger"
)

type AdminHandler struct {
	rootHandler *gin.Engine
	services    *service.Service
}

func NewAdminHandler(root *gin.Engine, services *service.Service) *AdminHandler {
	return &AdminHandler{
		rootHandler: root,
		services:    services,
	}
}

/* Инициализация маршрутов для авторизации пользователя */
func (h *AdminHandler) InitRoutes(
	middleware *map[string]func(c *gin.Context),
	hasRole func(role string) func(c *gin.Context),
) {
	// route.ADMIN_MAIN_ROUTE, (*middleware)[middlewareConstant.MN_UI], hasRole(roleConstant.ROLE_ADMIN)
	// URL: /admin
	admin := h.rootHandler.Group(route.ADMIN_MAIN_ROUTE, (*middleware)[middlewareConstant.MN_UI])
	{
		// URL: /admin/user
		user := admin.Group(route.ADMIN_USER)
		{
			// URL: /admin/user/get/all
			user.POST(route.GET_ALL_ROUTE, h.getAllUsers)
		}

		// URL: /admin/company
		company := admin.Group(route.ADMIN_COMPANY)
		{
			// URL: /admin/company/create
			company.POST(route.CREATE_ROUTE, h.createCompany)
		}

		// URL: /admin/system
		system := admin.Group(route.SYSTEM)
		{
			// URL: /admin/system/user/add/access
			system.POST(route.USER_MAIN_ROUTE+route.ADD_ROUTE+route.ACCESS, h.systemUserAddAccess)
		}
	}
}
