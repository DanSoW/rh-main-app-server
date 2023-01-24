package util

import (
	"errors"
	middlewareConstants "main-server/pkg/constant/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

/* Получение пользовательского идентификатора */
func GetUserId(c *gin.Context) (int, error) {
	id, ok := c.Get(middlewareConstants.USER_CTX)
	if !ok {
		return 0, errors.New("Пользователя не найдено")
	}

	idInt, ok := id.(int)
	if !ok {
		return 0, errors.New("Идентификатор пользователя недопустимого типа")
	}

	return idInt, nil
}

/* Функция получения пользовательских данных из контекста */
func GetContextUserInfo(c *gin.Context) (int, string, int, error) {
	usersId, exist := c.Get(middlewareConstants.USER_CTX)

	if !exist {
		return -1, "", -1, errors.New("Нет доступа!")
	}

	usersUuid, exist := c.Get(middlewareConstants.USER_UUID_CTX)

	if !exist {
		return -1, "", -1, errors.New("Нет доступа!")
	}

	domainsId, exist := c.Get(middlewareConstants.DOMAINS_ID)

	if !exist {
		return -1, "", -1, errors.New("Нет доступа!")
	}

	return usersId.(int), usersUuid.(string), domainsId.(int), nil
}

/* Структура сообщения об ошибке */
type ResponseMessage struct {
	Message string `json:"message" binding:"required"`
}

/* Генерация ответа об ошибке */
func NewErrorResponse(c *gin.Context, statusCode int, message string) {
	logrus.Error(message)
	c.AbortWithStatusJSON(statusCode, ResponseMessage{Message: message})
}
