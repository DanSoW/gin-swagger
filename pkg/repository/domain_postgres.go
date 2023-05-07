package repository

import (
	"errors"
	"fmt"
	tableConstants "main-server/pkg/constant/table"
	rbacModel "main-server/pkg/model/rbac"

	"github.com/jmoiron/sqlx"
)

type DomainPostgres struct {
	db *sqlx.DB
}

/*
* Функция создания экземпляра сервиса
 */
func NewDomainPostgres(db *sqlx.DB) *DomainPostgres {
	return &DomainPostgres{db: db}
}

/* Получение информации о домене */
func (r *DomainPostgres) Get(column string, value interface{}, check bool) (*rbacModel.DomainModel, error) {
	var domains []rbacModel.DomainModel
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1", tableConstants.AC_DOMAINS, column)

	var err error

	switch value.(type) {
	case int:
		err = r.db.Select(&domains, query, value.(int))
		break
	case string:
		err = r.db.Select(&domains, query, value.(string))
		break
	}

	if len(domains) <= 0 {
		if check {
			return nil, errors.New(fmt.Sprintf("Ошибка: домена по запросу %s:%s не найдено!", column, value))
		}

		return nil, nil
	}

	return &domains[len(domains)-1], err
}
