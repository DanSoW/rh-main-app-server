package handler

import (
	serviceModel "main-server/pkg/model/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Проверка токена доступа
// @Tags Общие API для сервисов
// @Description Проверка токена доступа
// @ID service-token-verify
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Токен доступа для текущего пользователя" example(Bearer access_token)
// @Success 200 {object} serviceModel.TokenVerifyModel "data"
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /service/main/verify [post]
func (h *Handler) serviceMainVerify(c *gin.Context) {
	_, userUuid, _, err := getContextUserInfo(c)
	if err != nil {
		newErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}

	c.JSON(http.StatusOK, serviceModel.TokenVerifyModel{
		Uuid: userUuid,
	})
}
