package repository

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	config "main-server/config"
	authConstants "main-server/pkg/constant/auth"
	middlewareConstants "main-server/pkg/constant/middleware"
	tableConstants "main-server/pkg/constant/table"
	"main-server/pkg/model/email"
	rbacModel "main-server/pkg/model/rbac"
	"main-server/pkg/model/user"
	userModel "main-server/pkg/model/user"
	authService "main-server/pkg/service/auth"
	smtpService "main-server/pkg/service/smtp"

	roleConstant "main-server/pkg/constant/role"

	"github.com/casbin/casbin/v2"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

type AuthPostgres struct {
	db           *sqlx.DB
	enforcer     *casbin.Enforcer
	userPostgres UserPostgres
}

/* Функция создания нового экземлпяра структуры AuthPostgres */
func NewAuthPostgres(db *sqlx.DB, enforcer *casbin.Enforcer, userPostgres UserPostgres) *AuthPostgres {
	return &AuthPostgres{
		db:           db,
		enforcer:     enforcer,
		userPostgres: userPostgres,
	}
}

/* Метод регистрации нового пользователя в системе */
func (r *AuthPostgres) CreateUser(user userModel.UserSignUpModel) (userModel.UserAuthDataModel, error) {
	check := CheckRowExists(r.db, tableConstants.U_USERS, "email", user.Email)
	if check {
		return userModel.UserAuthDataModel{}, errors.New("Пользователь с данным email-адресом уже существует!")
	}

	// Начало транзакции
	tx, err := r.db.Begin()
	if err != nil {
		return userModel.UserAuthDataModel{}, err
	}

	// Хэширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), viper.GetInt("crypt.cost"))
	if err != nil {
		return userModel.UserAuthDataModel{}, err
	}

	user.Password = string(hashedPassword)

	var id int
	var userUuid string

	// Запрос на добавление нового пользователя в систему
	query := fmt.Sprintf("INSERT INTO %s (email, password, uuid) values ($1, $2, $3) RETURNING id, uuid", tableConstants.U_USERS)

	// Генерация UUID
	u1 := uuid.NewV4()

	row := tx.QueryRow(query, user.Email, user.Password, u1)
	if err := row.Scan(&id, &userUuid); err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, errors.New("Пользователь с данными регистрационными данными уже существует!")
	}

	// Запрос на добавление пользовательских данных
	query = fmt.Sprintf(
		`INSERT INTO %s (data, created_at, updated_at, users_id) 
		values ($1, $2, $3, $4)`,
		tableConstants.U_USERS_DATA)

	currentDate := time.Now()
	_, err = tx.Exec(query, user.Data, currentDate, currentDate, id)

	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	query = fmt.Sprintf("SELECT * FROM %s WHERE value = $1 LIMIT 1", tableConstants.AC_DOMAINS)
	var domain rbacModel.DomainModel
	err = r.db.Get(&domain, query, viper.GetString("domain"))
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, errors.New("Домена не существует!")
	}

	query = fmt.Sprintf("SELECT * FROM %s WHERE value = $1 AND domains_id = $2 LIMIT 1", tableConstants.AC_ROLES)
	var role rbacModel.RoleModel
	err = r.db.Get(&role, query, roleConstant.ROLE_CLIENT, domain.Id)
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, errors.New("Роли пользователя не существует!")
	}

	// Добавление роли пользователю (по-умолчанию данная роль - USER)
	r.enforcer.AddRoleForUserInDomain(strconv.Itoa(id), strconv.Itoa(role.Id), strconv.Itoa(domain.Id))

	// Установка типа авторизации для пользователя (в данном случае - локальная авторизация, не через внешний сервис)
	var authTypes userModel.AuthTypeModel
	query = fmt.Sprintf("SELECT * FROM %s WHERE value=$1 LIMIT 1", tableConstants.U_AUTH_TYPES)
	err = r.db.Get(&authTypes, query, authConstants.AUTH_TYPE_LOCAL)
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, errors.New(err.Error())
	}

	query = fmt.Sprintf("INSERT INTO %s (users_id, auth_types_id) values ($1, $2)", tableConstants.U_USERS_AUTH_TYPES)
	_, err = tx.Exec(query, id, authTypes.Id)
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	// Генерация пары токенов (токен доступа и токен обновления)
	accessToken, err := GenerateToken(userUuid, authTypes.Uuid, nil, authConstants.TOKEN_TLL_ACCESS, viper.GetString("token.signing_key_access"))
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	refreshToken, err := GenerateToken(userUuid, authTypes.Uuid, nil, authConstants.TOKEN_TLL_REFRESH, viper.GetString("token.signing_key_refresh"))
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	// Добавления токенов пользователю
	query = fmt.Sprintf("INSERT INTO %s (users_id, access_token, refresh_token) values ($1, $2, $3)", tableConstants.U_TOKENS)
	_, err = tx.Exec(query, id, accessToken, refreshToken)
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	// Добавление пользователю ссылки для активации аккаунта
	u2 := uuid.NewV4()
	query = fmt.Sprintf("INSERT INTO %s (users_id, is_activated, activation_link) values ($1, $2, $3)", tableConstants.U_ACTIVATIONS)
	_, err = tx.Exec(query, id, false, u2)
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	// Отправка сообщения пользователю
	err = smtpService.SendMessage(user.Email, smtpService.BuildMessage(email.Mail{
		Sender:  viper.GetString("smtp.email"),
		To:      []string{user.Email},
		Subject: "Подтверждение аккаунта \"Rental housing\"",
		Body: fmt.Sprintf(`<html>
		<head>
			<meta charset="utf-8" />
			<title></title>
		</head>
		<style>
			body {background-color: #FEFEF9;}
			h2   {color: #181511;}
			button {
				color: rgb(0, 0, 0);
				outline: none;
				border: none;
				border-radius: 30px;
				background-color: #B19472;
				padding: 8px 16px;
				margin-top: 16px;
				cursor: pointer;
			}
		</style>
		<body>
			<h2>Подтверждение E-mail</h2>
			<br><text>Вы получили это письмо, так как Ваш почтовый адрес был указан в приложении "Rental housing".</text> 
			</br><text>Чтобы подтвердить Вашу почту перейдите по ссылке: </text></br>
			<a href="%s">
			<button>Подтвердить E-mail</button>
			</a>
			<br><br><br>
			<text>Если Вы не проходили процедуру регистрации в приложении "Rental housing", то не отвечайте на данное сообщение.</text>
		</body>
	</html>`, viper.GetString("api_url")+"/auth/activate/"+u2.String()),
	}))

	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	return userModel.UserAuthDataModel{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (r *AuthPostgres) UploadProfileImage(c *gin.Context, filepath string) (bool, error) {
	usersId, _ := c.Get(middlewareConstants.USER_CTX)

	query := fmt.Sprintf("SELECT data FROM %s tl WHERE users_id=$1 LIMIT 1", tableConstants.U_USERS_DATA)
	var userData []userModel.UserDataModel

	if err := r.db.Select(&userData, query, usersId); err != nil {
		return false, err
	}

	if len(userData) <= 0 {
		return false, errors.New("Данных у пользователя нет")
	}

	var dataFromJson userModel.UserDataDbModel
	if err := json.Unmarshal([]byte(userData[0].Data), &dataFromJson); err != nil {
		return false, err
	}
	dataFromJson.Avatar = filepath

	userJsonb, err := json.Marshal(dataFromJson)
	if err != nil {
		return false, err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return false, err
	}

	query = fmt.Sprintf("UPDATE %s tl SET data=$1 WHERE tl.users_id = $2", tableConstants.U_USERS_DATA)

	// Обновление данных о пользователе
	if _, err = tx.Exec(query, userJsonb, usersId); err != nil {
		tx.Rollback()
		return false, err
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return false, err
	}

	return true, nil
}

/* Авторизация пользователя */
func (r *AuthPostgres) LoginUser(user userModel.UserSignInModel) (userModel.UserAuthDataModel, error) {
	var findUser userModel.UserModel
	query := fmt.Sprintf("SELECT * FROM %s tl WHERE tl.email = $1 LIMIT 1", tableConstants.U_USERS)
	if err := r.db.Get(&findUser, query, user.Email); err != nil {
		return userModel.UserAuthDataModel{}, errors.New("Пользователя с данным почтовым адресом не существует!")
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(findUser.Password), []byte(user.Password)); err != nil {
		return userModel.UserAuthDataModel{}, errors.New("Не правильный пароль! Повторите попытку")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return userModel.UserAuthDataModel{}, err
	}

	query = fmt.Sprintf("DELETE FROM %s tl WHERE tl.users_id = $1", tableConstants.U_TOKENS)
	if _, err := r.db.Exec(query, findUser.Id); err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	query = fmt.Sprintf("SELECT * FROM %s WHERE value = $1 LIMIT 1", tableConstants.AC_DOMAINS)
	var domain rbacModel.DomainModel
	if err = r.db.Get(&domain, query, viper.GetString("domain")); err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, errors.New("Домена не существует!")
	}

	query = fmt.Sprintf("SELECT * FROM %s WHERE value = $1 AND domains_id = $2 LIMIT 1", tableConstants.AC_ROLES)
	var role rbacModel.RoleModel
	if err = r.db.Get(&role, query, roleConstant.ROLE_CLIENT, domain.Id); err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, errors.New("Роли пользователя для данного домена не существует!")
	}

	if _, err = r.enforcer.GetRolesForUser(strconv.Itoa(findUser.Id), strconv.Itoa(domain.Id)); err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	// Проверка имеет ли пользователь конкретную роль в рамках текущего домена
	flag, err := r.enforcer.HasRoleForUser(strconv.Itoa(findUser.Id), strconv.Itoa(role.Id), strconv.Itoa(domain.Id))
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	if !flag {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, errors.New("Данный пользователь не имеет доступа к данному домену!")
	}

	// Получение типа аутентификации (в данном случае - LOCAL)
	var authTypes userModel.AuthTypeModel
	query = fmt.Sprintf("SELECT * FROM %s WHERE value=$1 LIMIT 1", tableConstants.U_AUTH_TYPES)
	err = r.db.Get(&authTypes, query, authConstants.AUTH_TYPE_LOCAL)
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, errors.New(err.Error())
	}

	// Генерация токенов доступа и обновления
	accessToken, err := GenerateToken(findUser.Uuid, authTypes.Uuid, nil, authConstants.TOKEN_TLL_ACCESS, viper.GetString("token.signing_key_access"))
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}
	refreshToken, err := GenerateToken(findUser.Uuid, authTypes.Uuid, nil, authConstants.TOKEN_TLL_REFRESH, viper.GetString("token.signing_key_refresh"))
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	// Установка токенов пользователю
	query = fmt.Sprintf("INSERT INTO %s (users_id, access_token, refresh_token) values ($1, $2, $3)", tableConstants.U_TOKENS)
	_, err = tx.Exec(query, findUser.Id, accessToken, refreshToken)
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	return userModel.UserAuthDataModel{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

/* Авторизация пользователя через OAuth2 */
func (r *AuthPostgres) CreateUserOAuth2(user user.UserRegisterOAuth2Model, token *oauth2.Token) (userModel.UserAuthDataModel, error) {
	check := CheckRowExists(r.db, tableConstants.U_USERS, "email", user.Email)

	if check {
		return userModel.UserAuthDataModel{}, errors.New("Пользователь с данным email-адресом уже существует!")
	}

	// Начало транзакции
	tx, err := r.db.Begin()
	if err != nil {
		return userModel.UserAuthDataModel{}, err
	}

	// Хэширование пароля
	// user.Password = generatePasswordHash(user.Password)
	/*hashedPassword, err := bcrypt.GenerateFromPassword([]byte(token.AccessToken), viper.GetInt("crypt.cost"))
	if err != nil {
		return userModel.UserAuthDataModel{}, err
	}*/

	var id int
	var userUuid string

	// Запрос на добавление нового пользователя в систему
	query := fmt.Sprintf("INSERT INTO %s (email, password, uuid) values ($1, $2, $3) RETURNING id, uuid", tableConstants.U_USERS)

	// Генерация UUID
	u1 := uuid.NewV4()

	row := tx.QueryRow(query, user.Email, token.AccessToken, u1)
	if err := row.Scan(&id, &userUuid); err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, errors.New("Пользователь с данными регистрационными данными уже существует!")
	}

	// Запрос на добавление пользовательских данных
	query = fmt.Sprintf(
		`INSERT INTO %s (data, date_registration, users_id) 
		values ($1, $2, $3)`,
		tableConstants.U_USERS_DATA)

	userJsonb, err := json.Marshal(userModel.UserDataDbModel{
		Name:     user.Name,
		Surname:  user.FamilyName,
		Nickname: user.GivenName,
	})

	_, err = tx.Exec(query, userJsonb, time.Now(), id)

	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	query = fmt.Sprintf("SELECT * FROM %s WHERE value = $1 LIMIT 1", tableConstants.AC_DOMAINS)
	var domain rbacModel.DomainModel
	err = r.db.Get(&domain, query, viper.GetString("domain"))
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, errors.New("Домена не существует!")
	}

	query = fmt.Sprintf("SELECT * FROM %s WHERE value = $1 AND domains_id = $2 LIMIT 1", tableConstants.AC_ROLES)
	var role rbacModel.RoleModel
	err = r.db.Get(&role, query, roleConstant.ROLE_CLIENT, domain.Id)
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, errors.New("Роли пользователя не существует!")
	}

	/* Added default user roles */
	r.enforcer.AddRoleForUserInDomain(strconv.Itoa(id), strconv.Itoa(role.Id), strconv.Itoa(domain.Id))

	// Установка типа аутентификации пользователя (в данном случае - GOOGLE)
	var authTypes userModel.AuthTypeModel
	query = fmt.Sprintf("SELECT * FROM %s WHERE value=$1 LIMIT 1", tableConstants.U_AUTH_TYPES)
	err = r.db.Get(&authTypes, query, authConstants.AUTH_TYPE_GOOGLE)
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, errors.New(err.Error())
	}

	query = fmt.Sprintf("INSERT INTO %s (users_id, auth_types_id) values ($1, $2)", tableConstants.U_USERS_AUTH_TYPES)
	_, err = tx.Exec(query, id, authTypes.Id)
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	// Генерация токенов доступа
	accessToken, err := GenerateToken(userUuid, authTypes.Uuid, &token.AccessToken, authConstants.TOKEN_TLL_ACCESS, viper.GetString("token.signing_key_access"))
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	// Генерация токена обновления
	refreshToken, err := GenerateToken(userUuid, authTypes.Uuid, &token.RefreshToken, authConstants.TOKEN_TLL_REFRESH, viper.GetString("token.signing_key_access"))
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	// Установка токенов пользователю
	query = fmt.Sprintf("INSERT INTO %s (users_id, access_token, refresh_token) values ($1, $2, $3)", tableConstants.U_TOKENS)
	_, err = tx.Exec(query, id, accessToken, refreshToken)
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	// Добавление ссылки на активацию аккаунта
	// Генерация UUID
	u2 := uuid.NewV4()
	query = fmt.Sprintf("INSERT INTO %s (users_id, is_activated, activation_link) values ($1, $2, $3)", tableConstants.U_ACTIVATIONS)
	_, err = tx.Exec(query, id, true, u2)
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	return userModel.UserAuthDataModel{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, tx.Commit()
}

/*
* Функция авторизации пользователя через Google OAuth2
 */
func (r *AuthPostgres) LoginUserOAuth2(code string) (userModel.UserAuthDataModel, error) {
	token, err := config.AppOAuth2Config.GoogleLogin.Exchange(oauth2.NoContext, code)

	if err != nil {
		return userModel.UserAuthDataModel{}, err
	}

	is_verify, err := authService.VerifyAccessToken(token.AccessToken)

	if err != nil {
		return userModel.UserAuthDataModel{}, err
	}

	if !is_verify {
		return userModel.UserAuthDataModel{}, errors.New("Данный токен не принадлежит данному пользователю!")
	}

	var findUser userModel.UserModel
	var userData userModel.UserRegisterOAuth2Model

	userData, err = authService.GetUserInfo(token)

	if err != nil {
		return userModel.UserAuthDataModel{}, err
	}

	query := fmt.Sprintf("SELECT * FROM %s tl WHERE tl.email = $1 LIMIT 1", tableConstants.U_USERS)
	if err := r.db.Get(&findUser, query, userData.Email); err != nil {
		// Если пользователя не существует - создаём его
		return r.CreateUserOAuth2(userData, token)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return userModel.UserAuthDataModel{}, err
	}

	/*hashedPassword, err := bcrypt.GenerateFromPassword([]byte(token.AccessToken), viper.GetInt("crypt.cost"))
	if err != nil {
		return userModel.UserAuthDataModel{}, err
	}*/

	// Запрос на обновление пароля в базе данных для пользователя
	query = fmt.Sprintf("UPDATE %s SET password=$1 WHERE email=$2", tableConstants.U_USERS)

	if _, err := r.db.Exec(query, token.AccessToken, userData.Email); err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	query = fmt.Sprintf("DELETE FROM %s tl WHERE tl.users_id = $1", tableConstants.U_TOKENS)
	if _, err := r.db.Exec(query, findUser.Id); err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	// Получение типа аутентификации (в данном случае - GOOGLE)
	var authTypes userModel.AuthTypeModel
	query = fmt.Sprintf("SELECT * FROM %s WHERE value=$1 LIMIT 1", tableConstants.U_AUTH_TYPES)
	err = r.db.Get(&authTypes, query, authConstants.AUTH_TYPE_GOOGLE)
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, errors.New(err.Error())
	}

	// Генерация токена доступа
	accessToken, err := GenerateToken(findUser.Uuid, authTypes.Uuid, &token.AccessToken, authConstants.TOKEN_TLL_ACCESS, viper.GetString("token.signing_key_access"))
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	// Генерация токена обновления
	refreshToken, err := GenerateToken(findUser.Uuid, authTypes.Uuid, &token.RefreshToken, authConstants.TOKEN_TLL_REFRESH, viper.GetString("token.signing_key_access"))
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	// Установка токенов пользователю
	query = fmt.Sprintf("INSERT INTO %s (users_id, access_token, refresh_token) values ($1, $2, $3)", tableConstants.U_TOKENS)
	_, err = tx.Exec(query, findUser.Id, accessToken, refreshToken)
	if err != nil {
		tx.Rollback()
		return userModel.UserAuthDataModel{}, err
	}

	return userModel.UserAuthDataModel{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, tx.Commit()
}

/**
 * Функция для обновления токена доступа по токену обновления
 * @param {userModel.TokenLogoutDataModel} data - Подробная информация об авторизационной информации пользователя
 * @param {string} rToken - Токен обновления (refresh token)
 * @param {token userModel.TokenOutputParse} token - Данные, полученные после дешифровки токена доступа вне зависимости от его валидности
 * @returns {userModel.UserAuthDataModel, error} Пара токенов (access и refresh) или ошибка
 */
func (r *AuthPostgres) Refresh(data userModel.TokenLogoutDataModel, rToken string, token userModel.TokenOutputParse) (userModel.UserAuthDataModel, error) {
	user, err := r.userPostgres.Get("id", token.UsersId, true)
	if err != nil {
		return userModel.UserAuthDataModel{}, err
	}

	var findToken userModel.TokenModel
	query := fmt.Sprintf("SELECT * FROM %s tl WHERE tl.refresh_token = $1 AND tl.users_id = $2 LIMIT 1", tableConstants.U_TOKENS)

	if err := r.db.Get(&findToken, query, rToken, user.Id); err != nil {
		return userModel.UserAuthDataModel{}, errors.New("Пользователя с данным токеном обновления не существует!")
	}

	isValid := ValidToken(rToken, viper.GetString("token.signing_key_refresh"))

	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	var refreshToken string

	if !isValid {
		switch token.AuthType.Value {
		case "LOCAL":
			refreshToken, err = GenerateToken(user.Uuid, token.AuthType.Uuid, nil, authConstants.TOKEN_TLL_REFRESH, viper.GetString("token.signing_key_refresh"))
			break

		case "GOOGLE":
			// Если токен от Google OAuth2 не валиден, то нужно чтобы пользователь перезашёл в приложение заново
			//google_oauth2.RevokeToken(*token.TokenApi)
			//r.Logout(data)
			refreshToken, err = GenerateToken(user.Uuid, token.AuthType.Uuid, token.TokenApi, authConstants.TOKEN_TLL_REFRESH, viper.GetString("token.signing_key_refresh"))
			break
		}

		if err != nil {
			return userModel.UserAuthDataModel{}, err
		}

		setValues = append(setValues, fmt.Sprintf("refresh_token=$%d", argId))
		args = append(args, refreshToken)
		argId++
	} else {
		refreshToken = rToken
	}

	var accessToken string

	switch token.AuthType.Value {
	case authConstants.AUTH_TYPE_LOCAL:
		accessToken, err = GenerateToken(user.Uuid, token.AuthType.Uuid, nil, authConstants.TOKEN_TLL_ACCESS, viper.GetString("token.signing_key_access"))
		break

	case authConstants.AUTH_TYPE_GOOGLE:
		tokenData, err := authService.RefreshAccessToken(oauth2.NoContext, *token.TokenApi)
		accessToken, err = GenerateToken(user.Uuid, token.AuthType.Uuid, &tokenData.AccessToken, authConstants.TOKEN_TLL_ACCESS, viper.GetString("token.signing_key_access"))

		if err != nil {
			return userModel.UserAuthDataModel{}, err
		}
	}

	if err != nil {
		return userModel.UserAuthDataModel{}, err
	}

	setValues = append(setValues, fmt.Sprintf("access_token=$%d", argId))
	args = append(args, accessToken)
	argId++

	setQuery := strings.Join(setValues, ", ")

	query = fmt.Sprintf("UPDATE %s tl SET %s WHERE tl.users_id = $%d",
		tableConstants.U_TOKENS, setQuery, argId)
	args = append(args, user.Id)

	// Обновление данных о токене пользователя
	_, err = r.db.Exec(query, args...)
	if err != nil {
		return userModel.UserAuthDataModel{}, err
	}

	// Возвращение авторизационных данных пользователя
	return userModel.UserAuthDataModel{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

/*
*	Функция подтверждения аккаунта
 */
func (r *AuthPostgres) Activate(link string) (bool, error) {
	var findActivate userModel.UserActivateModel
	query := fmt.Sprintf("SELECT activation_link, is_activated FROM %s WHERE activation_link = $1", tableConstants.U_ACTIVATIONS)

	if err := r.db.Get(&findActivate, query, link); err != nil {
		return false, errors.New(err.Error())
	}

	if findActivate.IsActivated {
		return true, nil
	}

	query = fmt.Sprintf("UPDATE %s SET is_activated=%s WHERE activation_link = $1", tableConstants.U_ACTIVATIONS, "true")

	_, err := r.db.Exec(query, link)

	if err != nil {
		return false, err
	}

	return true, nil
}

/*
* Функция разлогирования пользователя
 */
func (r *AuthPostgres) Logout(data userModel.TokenLogoutDataModel) (bool, error) {
	// Выход из аккаунта зависит от метода аутентификации (предварительная проверка обязательна)
	switch data.AuthTypeValue {
	case authConstants.AUTH_TYPE_GOOGLE:
		authService.RevokeToken(*data.TokenApi)
		break
	}

	query := fmt.Sprintf("DELETE FROM %s tl WHERE tl.access_token=$1 AND tl.refresh_token=$2 RETURNING id", tableConstants.U_TOKENS)
	row := r.db.QueryRow(query, data.AccessToken, data.RefreshToken)

	var id int
	if err := row.Scan(&id); err != nil {
		return false, err
	}

	return true, nil
}

type AuthTypeValue struct {
	Value string `json:"value" db:"value"`
}

/*
* Функция обработки запроса на восстановление пароля
 */
func (r *AuthPostgres) RecoveryPassword(userEmail string) (bool, error) {
	// Check exists user in system
	user, err := r.GetUser("email", userEmail)
	if err != nil {
		return false, errors.New("Пользователя с данным email-адресом не существует!")
	}

	query := fmt.Sprintf(`SELECT value FROM %s tl
	INNER JOIN %s td on td.auth_types_id = tl.id WHERE td.users_id=$1 LIMIT 1`,
		tableConstants.U_AUTH_TYPES, tableConstants.U_USERS_AUTH_TYPES)

	var authType AuthTypeValue
	err = r.db.Get(&authType, query, user.Id)
	if err != nil {
		return false, err
	}

	if authType.Value != authConstants.AUTH_TYPE_LOCAL {
		return false, errors.New(`
		Восстановление пароля для данного пользователя не поддерживается, так как 
		пользователь авторизовался через сторонний сервис (Google, VK).
		Пожалуйста, воспользуйтесь сторонним сервисом для авторизации пользователя
		`)
	}

	// Delete other reset tokens for current user
	tx, err := r.db.Begin()
	if err != nil {
		return false, err
	}

	query = fmt.Sprintf("DELETE FROM %s tl WHERE users_id=$1", tableConstants.U_RESET_TOKENS)

	_, err = tx.Exec(query, user.Id)
	if err != nil {
		tx.Rollback()
		return false, err
	}

	// Generate new reset token with email and uuid current user
	token, err := GenerateResetToken(user.Uuid, user.Email, authConstants.TOKEN_TLL_RESET, viper.GetString("token.signing_key_reset"))

	if err != nil {
		tx.Rollback()
		return false, err
	}

	query = fmt.Sprintf("INSERT INTO %s (users_id, token) values ($1, $2)", tableConstants.U_RESET_TOKENS)
	_, err = tx.Exec(query, user.Id, token)
	if err != nil {
		tx.Rollback()
		return false, err
	}

	err = smtpService.SendMessage(user.Email, smtpService.BuildMessage(email.Mail{
		Sender:  viper.GetString("smtp.email"),
		To:      []string{user.Email},
		Subject: "Восстановление пароля \"МИСУ Мирный\"",
		Body: fmt.Sprintf(`<html>
		<head>
			<meta charset="utf-8" />
			<title></title>
		</head>
		<style>
			body {background-color: #FEFEF9;}
			h2   {color: #181511;}
			button {
				color: rgb(0, 0, 0);
				outline: none;
				border: none;
				border-radius: 30px;
				background-color: #B19472;
				padding: 8px 16px;
				margin-top: 16px;
				cursor: pointer;
			}
		</style>
		<body>
			<h2>Восстановление пароля по Email-адресу</h2>
			<br><text>Вы получили это письмо, так как Ваш почтовый адрес был указан в приложении "МИСУ Мирный".</text> 
			</br><text>Чтобы восстановить пароль перейдите по указанной ссылке: </text></br>
			<a href="%s">
			<button>Восстановить пароль</button>
			</a>
			<br><br><br>
			<text>Если Вы не проходили процедуру восстановления пароля в приложении "МИСУ Мирный", то не отвечайте на данное сообщение.</text>
		</body>
	</html>`, viper.GetString("crm_url")+"/auth/reset/password/"+token),
	}))

	err = tx.Commit()

	if err != nil {
		tx.Rollback()
		return false, err
	}

	return true, nil
}

/* Reset user password */
func (r *AuthPostgres) ResetPassword(data userModel.ResetPasswordModel, token userModel.ResetTokenOutputParse) (bool, error) {
	// Checking whether the token belongs to the current user
	resetToken, err := r.GetResetToken("token", data.Token)
	if err != nil {
		return false, err
	}

	if resetToken.UsersId != token.UsersId {
		return false, errors.New("Данный токен сброса пароля не принадлежит данному пользователю")
	}

	// Password reset procedure
	tx, err := r.db.Begin()
	if err != nil {
		return false, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password), viper.GetInt("crypt.cost"))
	if err != nil {
		return false, err
	}

	query := fmt.Sprintf("UPDATE %s SET password=$1 WHERE id=$2", tableConstants.U_USERS)
	_, err = tx.Exec(query, string(hashedPassword), token.UsersId)
	if err != nil {
		tx.Rollback()
		return false, err
	}

	// Delete all reset tokens for current users
	query = fmt.Sprintf("DELETE FROM %s tl WHERE users_id=$1", tableConstants.U_RESET_TOKENS)

	_, err = tx.Exec(query, token.UsersId)
	if err != nil {
		tx.Rollback()
		return false, err
	}

	err = tx.Commit()

	if err != nil {
		tx.Rollback()
		return false, err
	}

	return true, nil
}

/*
* User data acquisition function
 */
func (r *AuthPostgres) GetUser(column, value string) (userModel.UserModel, error) {
	var user userModel.UserModel
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1", tableConstants.U_USERS, column)

	err := r.db.Get(&user, query, value)

	return user, err
}

/*
* Function for getting role data
 */
func (r *AuthPostgres) GetRole(column, value string) (rbacModel.RoleModel, error) {
	var user rbacModel.RoleModel
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1", tableConstants.AC_ROLES, column)

	err := r.db.Get(&user, query, value)

	return user, err
}

/*
* User reset tokens
 */
func (r *AuthPostgres) GetResetToken(column, value string) (userModel.ResetTokenModel, error) {
	var token userModel.ResetTokenModel
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1", tableConstants.U_RESET_TOKENS, column)

	err := r.db.Get(&token, query, value)

	return token, err
}

/*
* Password hashing function
 */
func generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(viper.GetString("crypt.salt"))))
}

/* Working with user authentication tokens */
/* Token Body Structure */
type tokenClaims struct {
	jwt.StandardClaims
	UsersId     string  `json:"users_id"`      // ID пользователя
	AuthTypesId string  `json:"auth_types_id"` // Тип аутентификации пользователя
	TokenApi    *string `json:"token_api"`     // Внешний токен доступа
}

/*
* Token generation function
 */
func GenerateToken(uuid, authTypesUuid string, tokenApi *string, tokenTTL time.Duration, signingKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		uuid,
		authTypesUuid,
		tokenApi,
	})

	return token.SignedString([]byte(signingKey))
}

/*
* Token validity verification function
 */
func ValidToken(pToken, signingKey string) bool {
	_, err := jwt.ParseWithClaims(pToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(signingKey), nil
	})

	if err != nil {
		return false
	}

	return true
}

/* Working with Password Reset Tokens */
/* Token Body Structure */
type tokenResetClaims struct {
	jwt.StandardClaims
	UsersId string `json:"users_id"` // ID пользователя
	Email   string `json:"email"`    // Email пользователя
}

/*
* Reset token generation function
 */
func GenerateResetToken(uuid, email string, tokenTTL time.Duration, signingKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenResetClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		uuid,
		email,
	})

	return token.SignedString([]byte(signingKey))
}

/*
* Reset token validity verification function
 */
func ValidResetToken(pToken, signingKey string) bool {
	_, err := jwt.ParseWithClaims(pToken, &tokenResetClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(signingKey), nil
	})

	if err != nil {
		return false
	}

	return true
}
