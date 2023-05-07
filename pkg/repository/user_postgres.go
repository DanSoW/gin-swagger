package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	middlewareConstant "main-server/pkg/constant/middleware"
	tableConstant "main-server/pkg/constant/table"
	rbacModel "main-server/pkg/model/rbac"
	resourceModel "main-server/pkg/model/resource"
	userModel "main-server/pkg/model/user"
	"strconv"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type UserPostgres struct {
	db       *sqlx.DB
	enforcer *casbin.Enforcer
	domain   *DomainPostgres
	role     *RolePostgres
}

/*
* Функция создания экземпляра сервиса
 */
func NewUserPostgres(
	db *sqlx.DB, enforcer *casbin.Enforcer,
	domain *DomainPostgres, role *RolePostgres,
) *UserPostgres {
	return &UserPostgres{
		db:       db,
		enforcer: enforcer,
		domain:   domain,
		role:     role,
	}
}

func (r *UserPostgres) Get(column string, value interface{}, check bool) (*userModel.UserModel, error) {
	var users []userModel.UserModel
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1", tableConstant.U_USERS, column)

	var err error

	switch value.(type) {
	case int:
		err = r.db.Select(&users, query, value.(int))
		break
	case string:
		err = r.db.Select(&users, query, value.(string))
		break
	}

	if len(users) <= 0 {
		if check {
			return nil, errors.New(fmt.Sprintf("Ошибка: пользователя по запросу %s:%s не найдено!", column, value))
		}

		return nil, nil
	}

	return &users[len(users)-1], err
}

func (r *UserPostgres) GetProfile(c *gin.Context) (userModel.UserProfileModel, error) {
	usersId, _ := c.Get(middlewareConstant.USER_CTX)

	var profile userModel.UserProfileModel
	var email userModel.UserEmailModel

	query := fmt.Sprintf("SELECT data FROM %s tl WHERE tl.users_id = $1 LIMIT 1",
		tableConstant.U_USERS_DATA,
	)

	err := r.db.Get(&profile, query, usersId)
	if err != nil {
		return userModel.UserProfileModel{}, err
	}

	query = fmt.Sprintf("SELECT email FROM %s tl WHERE tl.id = $1 LIMIT 1", tableConstant.U_USERS)

	err = r.db.Get(&email, query, usersId)
	if err != nil {
		return userModel.UserProfileModel{}, err
	}

	return userModel.UserProfileModel{
		Email: email.Email,
		Data:  profile.Data,
	}, nil
}

func (r *UserPostgres) UpdateProfile(c *gin.Context, data userModel.UserProfileUpdateDataModel) (userModel.UserDataDbModel, error) {
	usersId, _ := c.Get(middlewareConstant.USER_CTX)

	userJsonb, err := json.Marshal(data)
	if err != nil {
		return userModel.UserDataDbModel{}, err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return userModel.UserDataDbModel{}, err
	}

	query := fmt.Sprintf("UPDATE %s tl SET data=$1 WHERE tl.users_id = $2", tableConstant.U_USERS_DATA)

	// Update data about user profile
	_, err = tx.Exec(query, userJsonb, usersId)
	if err != nil {
		tx.Rollback()
		return userModel.UserDataDbModel{}, err
	}

	query = fmt.Sprintf("SELECT data FROM %s tl WHERE users_id=$1 LIMIT 1", tableConstant.U_USERS_DATA)
	var userData []userModel.UserDataModel

	err = r.db.Select(&userData, query, usersId)

	if err != nil {
		tx.Rollback()
		return userModel.UserDataDbModel{}, err
	}

	if len(userData) <= 0 {
		tx.Rollback()
		return userModel.UserDataDbModel{}, errors.New("Данных у пользователя нет")
	}

	var dataFromJson userModel.UserDataDbModel
	err = json.Unmarshal([]byte(userData[0].Data), &dataFromJson)

	if err != nil {
		tx.Rollback()
		return userModel.UserDataDbModel{}, err
	}

	// Change password
	if data.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*data.Password), viper.GetInt("crypt.cost"))
		if err != nil {
			tx.Rollback()
			return userModel.UserDataDbModel{}, err
		}

		query := fmt.Sprintf("UPDATE %s SET password=$1 WHERE id=$2", tableConstant.U_USERS)
		_, err = r.db.Exec(query, string(hashedPassword), usersId)

		if err != nil {
			tx.Rollback()
			return userModel.UserDataDbModel{}, err
		}
	}

	err = tx.Commit()

	if err != nil {
		tx.Rollback()
		return userModel.UserDataDbModel{}, err
	}

	return dataFromJson, nil
}

/* Обновление изображения профиля пользователя */
func (r *UserPostgres) UpdateProfileImage(userIdentity *userModel.UserIdentityModel, resource *resourceModel.ImageModel) (*resourceModel.ImageModel, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	// Запрос на обновление информации о изображении пользователя
	query := fmt.Sprintf(
		`UPDATE %s u SET data = jsonb_set(data, '{avatar}', to_jsonb($1::text), true) WHERE u.id = $2`,
		tableConstant.U_USERS_DATA,
	)

	// Выполнение запроса на обновление
	if _, err = r.db.Exec(query, resource.Filepath, userIdentity.UserId); err != nil {
		tx.Rollback()
		return nil, err
	}

	// Фиксация изменений в транзакции
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return nil, err
	}

	return resource, nil
}

/* Проверка доступа пользователя */
func (r *UserPostgres) AccessCheck(userId, domainId int, value rbacModel.RoleValueModel) (bool, error) {
	role, err := r.role.Get("value", value.Value, true)
	if err != nil {
		return false, err
	}

	result, err := r.enforcer.HasRoleForUser(strconv.Itoa(userId), strconv.Itoa(role.Id), strconv.Itoa(domainId))
	if err != nil {
		return false, err
	}

	return result, nil
}

/* Метод для получения информации о всех ролях пользователя (функциональные модули пользователя) */
func (r *UserPostgres) GetAllRoles(user userModel.UserIdentityModel) (*userModel.UserRoleModel, error) {
	// Определение всех групп, в которых участвует пользователь
	roles, err := r.enforcer.GetRolesForUser(strconv.Itoa(user.UserId), strconv.Itoa(user.DomainId))
	var userRole userModel.UserRoleModel
	userRole.Domain = user.DomainUuid

	query := fmt.Sprintf(`
		SELECT value FROM %s ru WHERE ru.id = $1 LIMIT 1
	`, tableConstant.AC_ROLES)

	// Определяем данные для пользователя
	for _, item := range roles {
		var value userModel.RoleContextModel
		value.Context = nil

		subItems := strings.Split(item, rbacModel.Separator)
		if len(subItems) > 1 {
			value.Context = &subItems[len(subItems)-1]
		}

		var nameRole []string
		if err = r.db.Select(&nameRole, query, subItems[0]); err != nil {
			return nil, err
		}
		if len(nameRole) == 0 {
			return nil, errors.New("Ошибка: определённой роли не присутствует в базе данных")
		}

		value.Name = nameRole[len(nameRole)-1]
		userRole.Roles = append(userRole.Roles, value)
	}

	return &userRole, nil
}
