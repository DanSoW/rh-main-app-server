package service

import (
	utilContext "main-server/pkg/handler/util"
	emailModel "main-server/pkg/model/email"
	httpModel "main-server/pkg/model/http"
	serviceModel "main-server/pkg/model/service"
	"main-server/pkg/model/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Проверка токена доступа
// @Tags Общие API для сервисов
// @Description Проверка токена доступа
// @ID service-main-token-verify
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Токен доступа для текущего пользователя" example(Bearer access_token)
// @Success 200 {object} serviceModel.TokenVerifyModel "data"
// @Failure 400,404 {object} ResponseMessage
// @Failure 500 {object} ResponseMessage
// @Failure default {object} ResponseMessage
// @Router /service/main/verify [post]
func (h *ServiceHandler) serviceMainVerify(c *gin.Context) {
	_, userUuid, _, err := utilContext.GetContextUserInfo(c)
	if err != nil {
		utilContext.NewErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}

	c.JSON(http.StatusOK, serviceModel.TokenVerifyModel{
		Uuid: userUuid,
	})
}

// @Summary Отправка сообщения пользователю
// @Tags Общие API для сервисов
// @Description Проверка токена доступа
// @ID service-main-email-send
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Токен доступа для текущего пользователя" example(Bearer access_token)
// @Param input body emailModel.MessageInputModel true "Информация для отправки сообщения"
// @Success 200 {object} httpModel.ResponseValue "data"
// @Failure 400,404 {object} ResponseMessage
// @Failure 500 {object} ResponseMessage
// @Failure default {object} ResponseMessage
// @Router /service/main/email/send [post]
func (h *ServiceHandler) serviceMainEmailSend(c *gin.Context) {
	var input emailModel.MessageInputModel

	if err := c.BindJSON(&input); err != nil {
		utilContext.NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	userId, _, domainId, err := utilContext.GetContextUserInfo(c)

	data, err := h.services.SendEmail(&user.UserIdentityModel{UserId: userId, DomainId: domainId}, &input)

	if err != nil {
		utilContext.NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, httpModel.ResponseValue{
		Value: data,
	})
}
