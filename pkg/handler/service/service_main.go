package service

import (
	utilContext "main-server/pkg/handler/util"
	emailModel "main-server/pkg/model/email"
	httpModel "main-server/pkg/model/http"
	serviceModel "main-server/pkg/model/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Проверка токена доступа
// @Tags API для внешних сервисов
// @Description Проверка токена доступа
// @ID service-external-verify
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Токен доступа для текущего пользователя" example(Bearer access_token)
// @Success 200 {object} serviceModel.TokenVerifyModel "data"
// @Failure 400,404 {object} httpModel.ResponseMessage
// @Failure 500 {object} httpModel.ResponseMessage
// @Failure default {object} httpModel.ResponseMessage
// @Router /service/external/verify [post]
func (h *ServiceHandler) serviceExternalVerify(c *gin.Context) {
	userIdentity, err := utilContext.GetContextUserInfo(c)
	if err != nil {
		utilContext.NewErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}

	c.JSON(http.StatusOK, serviceModel.TokenVerifyModel{
		Uuid: userIdentity.UserUuid,
	})
}

// @Summary Отправка сообщения пользователю
// @Tags API для внешних сервисов
// @Description Проверка токена доступа
// @ID service-main-email-send
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Токен доступа для текущего пользователя" example(Bearer access_token)
// @Param input body emailModel.MessageInputModel true "Информация для отправки сообщения"
// @Success 200 {object} httpModel.ResponseValue "data"
// @Failure 400,404 {object} httpModel.ResponseMessage
// @Failure 500 {object} httpModel.ResponseMessage
// @Failure default {object} httpModel.ResponseMessage
// @Router /service/main/email/send [post]
func (h *ServiceHandler) serviceMainEmailSend(c *gin.Context) {
	var input emailModel.MessageInputModel

	if err := c.BindJSON(&input); err != nil {
		utilContext.NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	userIdentity, err := utilContext.GetContextUserInfo(c)
	if err != nil {
		utilContext.NewErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}

	data, err := h.services.SendEmail(userIdentity, &input)

	if err != nil {
		utilContext.NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, httpModel.ResponseValue{
		Value: data,
	})
}
