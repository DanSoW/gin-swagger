package rbac

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const Separator = ";"

/* Group Policy Subject Model - Модель для работы с объектом запроса роли */
type GPSubjectModel struct {
	RoleId     int    `json:"role_id"`     // Идентификатор роли
	ObjectUuid string `json:"object_uuid"` // Идентификатор объекта, в контексте которой действует роль
}

/* Конвертация модели GPSubjectModel в строку*/
func (gpsm *GPSubjectModel) ToString() string {
	return fmt.Sprintf("%d%s%s", gpsm.RoleId, Separator, gpsm.ObjectUuid)
}

/* Генерация новой модели */
func NewGPSubjectModel(model string) (*GPSubjectModel, error) {
	arr := strings.Split(model, Separator)

	if len(arr) != 2 {
		return nil, errors.New("Ошибка: несоответствие строки модели")
	}
	if len(arr[1]) < 36 {
		return nil, errors.New("Ошибка: идентификатор объекта должен быть длиной 36 символов и формата UUID")
	}

	id, err := strconv.Atoi(arr[0])
	if err != nil {
		return nil, errors.New("Ошибка: id роли не может быть строкой")
	}

	return &GPSubjectModel{
		RoleId:     id,
		ObjectUuid: arr[1],
	}, nil
}
