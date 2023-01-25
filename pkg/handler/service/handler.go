package service

import (
	_ "main-server/docs"

	"main-server/pkg/constant/route"
	service "main-server/pkg/service"

	middlewareConstant "main-server/pkg/constant/middleware"

	"github.com/gin-gonic/gin"
	_ "github.com/swaggo/files"
	_ "github.com/swaggo/gin-swagger"
)

type ServiceHandler struct {
	rootHandler *gin.Engine
	services    *service.Service
}

func NewServiceHandler(root *gin.Engine, services *service.Service) *ServiceHandler {
	return &ServiceHandler{
		rootHandler: root,
		services:    services,
	}
}

/* Инициализация маршрутов для сервисов */
func (h *ServiceHandler) InitRoutes(
	middleware *map[string]func(c *gin.Context),
) {
	// URL: /service
	service := h.rootHandler.Group(route.SERVICE, (*middleware)[middlewareConstant.MN_UI])
	{
		// URL: /main
		main := service.Group(route.SERVICE_MAIN)
		{
			// URL: /verify
			main.POST(route.SERVICE_VERIFY, h.serviceMainVerify)

			// URL: /mail/send
			main.POST(route.SERVICE_EMAIL_SEND, h.serviceMainEmailSend)
		}
	}
}
