package company

import (
	"fmt"
	_ "main-server/docs"
	middlewareConstant "main-server/pkg/constant/middleware"
	roleConstant "main-server/pkg/constant/role"
	"main-server/pkg/constant/route"
	service "main-server/pkg/service"

	"github.com/gin-gonic/gin"
	_ "github.com/swaggo/files"
	_ "github.com/swaggo/gin-swagger"
)

type CompanyHandler struct {
	rootHandler *gin.Engine
	services    *service.Service
}

func NewCompanyHandler(root *gin.Engine, services *service.Service) *CompanyHandler {
	return &CompanyHandler{
		rootHandler: root,
		services:    services,
	}
}

/* Инициализация маршрутов для авторизации пользователя */
func (h *CompanyHandler) InitRoutes(
	middleware *map[string]func(c *gin.Context),
	hasRole func(role string) func(c *gin.Context),
	hasRoles func(exp string, roles ...string) func(c *gin.Context),
) {
	// URL: /company
	company := h.rootHandler.Group(
		route.COMPANY_MAIN_ROUTE,
		(*middleware)[middlewareConstant.MN_UI],
		hasRoles("OR", roleConstant.ROLE_ADMIN, roleConstant.ROLE_BUILDER_ADMIN),
	)
	{
		// URL: /project
		project := company.Group(route.PROJECT_MAIN_ROUTE)
		{
			// URL: /company/project/create
			project.POST(route.CREATE_ROUTE, hasRole(roleConstant.ROLE_BUILDER_ADMIN), h.createProject)

			// URL: /company/project/update
			project.POST(
				route.UPDATE_ROUTE,
				hasRoles(
					"OR",
					roleConstant.ROLE_BUILDER_MANAGER,
					roleConstant.ROLE_BUILDER_ADMIN,
				),
				h.projectUpdate,
			)

			// URL: /company/project/update/image
			project.POST(
				fmt.Sprintf("%s/%s", route.UPDATE_ROUTE, route.RESOURCE_IMAGE_ROUTE),
				hasRoles(
					"OR",
					roleConstant.ROLE_BUILDER_MANAGER,
					roleConstant.ROLE_BUILDER_ADMIN,
				),
				h.projectUpdateImage,
			)

			// URL: /company/project/get
			project.POST(route.GET_ROUTE, h.getProject)

			// URL: /company/project/get/all
			project.POST(route.GET_ALL_ROUTE, h.getProjects)
		}

		// URL: /manager
		manager := company.Group(route.MANAGER_MAIN_ROUTE)
		{
			// URL: /company/manager/get/all
			manager.POST(route.GET_ALL_ROUTE, hasRole(roleConstant.ROLE_BUILDER_ADMIN), h.getManagers)

			// URL: /company/manager/get
			manager.POST(route.GET_ROUTE, h.companyGetManager)
		}

		// URL: /company/update/image
		company.POST(fmt.Sprintf("%s/%s", route.UPDATE_ROUTE, route.RESOURCE_IMAGE_ROUTE),
			hasRoles(
				"OR",
				roleConstant.ROLE_BUILDER_ADMIN,
				roleConstant.ROLE_ADMIN,
				roleConstant.ROLE_MANAGER,
				roleConstant.ROLE_SUPER_ADMIN,
			),
			h.companyUpdateImage,
		)

		// URL: /company/update
		company.POST(route.UPDATE_ROUTE,
			hasRoles(
				"OR",
				roleConstant.ROLE_BUILDER_ADMIN,
				roleConstant.ROLE_ADMIN,
				roleConstant.ROLE_MANAGER,
				roleConstant.ROLE_SUPER_ADMIN,
			),
			h.companyUpdate,
		)
	}
}
