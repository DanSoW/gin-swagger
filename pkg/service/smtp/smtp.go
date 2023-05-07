package smtp

import (
	"fmt"
	"main-server/pkg/model/email"
	"net/smtp"
	"os"

	"github.com/spf13/viper"
)

func BuildMessage(mail email.Mail) string {
	msg := ""
	msg += fmt.Sprintf("from: %s\r\n", mail.Sender)

	if len(mail.To) > 0 {
		msg += fmt.Sprintf("to: %s\r\n", mail.To[0])
	}

	msg += fmt.Sprintf("Subject: %s\r\n", mail.Subject)
	msg += fmt.Sprintf("MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n")
	msg += fmt.Sprintf("\r\n%s\r\n", mail.Body)

	return msg
}

func SendMessage(to, message string) error {
	auth := smtp.PlainAuth("", viper.GetString("smtp.email"), os.Getenv("SMTP_PASSWORD"), viper.GetString("smtp.host"))

	err := smtp.SendMail(viper.GetString("smtp.host")+":"+viper.GetString("smtp.port"), auth,
		viper.GetString("smtp.email"), []string{to}, []byte(message))

	return err
}

func SendMessageToLot(to []string, message string) error {
	auth := smtp.PlainAuth("", viper.GetString("smtp.email"), os.Getenv("SMTP_PASSWORD"), viper.GetString("smtp.host"))

	err := smtp.SendMail(viper.GetString("smtp.host")+":"+viper.GetString("smtp.port"), auth,
		viper.GetString("smtp.email"), to, []byte(message))

	return err
}
