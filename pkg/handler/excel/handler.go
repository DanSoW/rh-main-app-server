package excel

import (
	_ "main-server/docs"

	"main-server/pkg/constant/route"
	service "main-server/pkg/service"

	"github.com/gin-gonic/gin"
	_ "github.com/swaggo/files"
	_ "github.com/swaggo/gin-swagger"
)

type ExcelHandler struct {
	rootHandler *gin.Engine
	services    *service.Service
}

func NewExcelHandler(root *gin.Engine, services *service.Service) *ExcelHandler {
	return &ExcelHandler{
		rootHandler: root,
		services:    services,
	}
}

/* Инициализация маршрутов для авторизации пользователя */
func (h *ExcelHandler) InitRoutes(
	middleware *map[string]func(c *gin.Context),
) {
	// URL: /excel
	excel := h.rootHandler.Group(route.EXCEL_MAIN)
	{
		// URL: /excel/analysis
		excel.POST(route.EXCEL_ANALYSIS, h.excelAnalysis)
	}
}
