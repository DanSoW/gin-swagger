package repository

import (
	"errors"
	"fmt"
	tableConstant "main-server/pkg/constant/table"
	userModel "main-server/pkg/model/user"

	"github.com/jmoiron/sqlx"
)

type AuthTypePostgres struct {
	db *sqlx.DB
}

/*
* Функция создания экземпляра сервиса
 */
func NewAuthTypePostgres(db *sqlx.DB) *AuthTypePostgres {
	return &AuthTypePostgres{db: db}
}

/*
* Функция получения данных о роли
 */
func (r *AuthTypePostgres) Get(column string, value interface{}, check bool) (*userModel.AuthTypeModel, error) {
	var authTypes []userModel.AuthTypeModel
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1", tableConstant.U_AUTH_TYPES, column)

	var err error

	switch value.(type) {
	case int:
		err = r.db.Select(&authTypes, query, value.(int))
		break
	case string:
		err = r.db.Select(&authTypes, query, value.(string))
		break
	}

	if len(authTypes) <= 0 {
		if check {
			return nil, errors.New(fmt.Sprintf("Ошибка: типа авторизации по запросу %s:%s не найдено!", column, value))
		}

		return nil, nil
	}

	return &authTypes[len(authTypes)-1], err
}
