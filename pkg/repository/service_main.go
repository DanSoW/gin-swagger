package repository

import (
	"main-server/pkg/model/email"
	emailModel "main-server/pkg/model/email"
	userModel "main-server/pkg/model/user"
	smtpService "main-server/pkg/service/smtp"

	"github.com/casbin/casbin/v2"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

/* Структура описывающая пакет текущего репозитория */
type ServiceMainRepository struct {
	db           *sqlx.DB
	enforcer     *casbin.Enforcer
	userPostgres *UserPostgres
}

/* Создание нового экземпляра репозитория */
func NewServiceMainRepository(db *sqlx.DB, enforcer *casbin.Enforcer, userPostgres *UserPostgres) *ServiceMainRepository {
	return &ServiceMainRepository{
		db:           db,
		enforcer:     enforcer,
		userPostgres: userPostgres,
	}
}

/* Отправка сообщения пользователю */
func (r *ServiceMainRepository) SendEmail(user *userModel.UserIdentityModel, body *emailModel.MessageInputModel) (bool, error) {
	var emailReceivers []string

	// Конвертация UUID пользователей в Email-адреса
	for _, item := range body.UuidReceivers {
		userItem, err := r.userPostgres.Get("uuid", item, true)
		if err != nil {
			return false, err
		}

		emailReceivers = append(emailReceivers, userItem.Email)
	}

	// Отправка сообщения
	err := smtpService.SendMessageToLot(
		emailReceivers,
		smtpService.BuildMessage(email.Mail{
			Sender:  viper.GetString("smtp.email"),
			To:      emailReceivers,
			Subject: body.Subject,
			Body:    body.Message,
		}),
	)

	if err != nil {
		return false, err
	}

	return true, nil
}
