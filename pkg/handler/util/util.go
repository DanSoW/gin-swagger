package util

import (
	"errors"
	middlewareConstants "main-server/pkg/constant/middleware"
	userModel "main-server/pkg/model/user"

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
func GetContextUserInfo(c *gin.Context) (*userModel.UserIdentityModel, error) {
	values := map[string]interface{}{}

	for _, item := range []string{
		middlewareConstants.USER_CTX,
		middlewareConstants.USER_UUID_CTX,
		middlewareConstants.DOMAINS_ID,
		middlewareConstants.DOMAINS_UUID,
	} {
		value, exists := c.Get(item)
		if !exists {
			return nil, errors.New("Нет доступа!")
		}

		values[item] = value
	}

	return &userModel.UserIdentityModel{
		UserId:     values[middlewareConstants.USER_CTX].(int),
		UserUuid:   values[middlewareConstants.USER_UUID_CTX].(string),
		DomainId:   values[middlewareConstants.DOMAINS_ID].(int),
		DomainUuid: values[middlewareConstants.DOMAINS_UUID].(string),
	}, nil
}

/* Структура сообщения об ошибке */
type ResponseMessage struct {
	Message string `json:"message" binding:"required"`
}

/* Генерация сообщения об ошибке */
func NewErrorResponse(c *gin.Context, statusCode int, message string) {
	// Локальное логирование ошибок (в файл)
	logrus.Error(message)

	// Завершение HTTP-запроса с ошибкой
	c.AbortWithStatusJSON(statusCode, ResponseMessage{Message: message})
}
