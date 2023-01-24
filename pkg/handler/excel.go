package handler

import (
	utilContext "main-server/pkg/handler/util"
	excelModel "main-server/pkg/model/excel"
	infoModel "main-server/pkg/module/excel_analysis/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Получение основной информации из Excel-таблицы
// @Tags Работа с удалёнными таблицами
// @Description Проверка токена доступа
// @ID excel-analysis
// @Accept  json
// @Produce  json
// @Param input body excelModel.DocumentIdModel true "Идентификатор удалённой таблицы"
// @Success 200 {object} infoModel.HeaderInfoModel "data"
// @Failure 400,404 {object} ResponseMessage
// @Failure 500 {object} ResponseMessage
// @Failure default {object} ResponseMessage
// @Router /excel/analysis [post]
func (h *Handler) excelAnalysis(c *gin.Context) {
	var input excelModel.DocumentIdModel

	if err := c.BindJSON(&input); err != nil {
		utilContext.NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	data, err := h.services.GetHeaderInfoDocument(input)

	if err != nil {
		utilContext.NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, &infoModel.HeaderInfoModel{
		Title:              data.Title,
		AddressItem:        data.AddressItem,
		TimeDelivery:       data.TimeDelivery,
		PaymentVariant:     data.PaymentVariant,
		PropertyItem:       data.PropertyItem,
		CommunicateVariant: data.CommunicateVariant,
	})
}
