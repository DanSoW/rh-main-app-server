package handler

import (
	companyModel "main-server/pkg/model/company"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary GetManagers
// @Tags company
// @Description Получение среза из общего числа менеджеров компании
// @ID company-manager-get-all
// @Accept  json
// @Produce  json
// @Param input body companyModel.ManagerCountModel true "credentials"
// @Success 200 {object} companyModel.ManagerAnyCountModel "data"
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /company/manager/get/all [post]
func (h *Handler) getManagers(c *gin.Context) {
	var input companyModel.ManagerCountModel

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	userId, domainId, err := getContextUserInfo(c)
	if err != nil {
		newErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}

	data, err := h.services.Company.GetManagers(userId, domainId, input)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, data)
}
