package user

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

/* Основная модель пользователя для регистрации */
type UserSignUpModel struct {
	Id       int             `json:"id" db:"id"`
	Email    string          `json:"email" binding:"required"`
	Password string          `json:"password" binding:"required"`
	Data     UserDataDbModel `json:"data" binding:"required"`
}

/* Модель пользователя для данных */
type UserDataDbModel struct {
	Name       string `json:"name" binding:"required"`
	Surname    string `json:"surname" binding:"required"`
	Nickname   string `json:"nickname" binding:"required"`
	Patronymic string `json:"patronymic"`
	Avatar     string `json:"avatar"`
}

/* Переопределение метода для получения структуры из JSON-строки */
func (pdm *UserDataDbModel) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		json.Unmarshal(v, &pdm)
		return nil
	case string:
		json.Unmarshal([]byte(v), &pdm)
		return nil
	default:
		return errors.New(fmt.Sprintf("Неподдерживаемый тип: %T", v))
	}
}

/* Переопределение метода для получения JSON-строки из структуры */
func (pdm UserDataDbModel) Value() (driver.Value, error) {
	return json.Marshal(&pdm)
}
