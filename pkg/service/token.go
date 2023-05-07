package service

import (
	"errors"
	userModel "main-server/pkg/model/user"
	repository "main-server/pkg/repository"

	"github.com/dgrijalva/jwt-go"
)

/* Структура TokenService */
type TokenService struct {
	role     repository.Role
	user     repository.User
	authType repository.AuthType
}

/* Функция создания нового сервиса TokenService */
func NewTokenService(role repository.Role,
	user repository.User,
	authType repository.AuthType,
) *TokenService {
	return &TokenService{
		role:     role,
		user:     user,
		authType: authType,
	}
}

/* Структура полезных данных JWT-токена */
type tokenClaims struct {
	jwt.StandardClaims
	UsersId     string  `json:"users_id"`      // ID for user
	AuthTypesId string  `json:"auth_types_id"` // Type auth for user
	TokenApi    *string `json:"token_api"`     // External token access
}

/* Парсинг токена с предварительной валидацией */
func (s *TokenService) ParseToken(pToken, signingKey string) (userModel.TokenOutputParse, error) {
	token, err := jwt.ParseWithClaims(pToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(signingKey), nil
	})
	if err != nil {
		return userModel.TokenOutputParse{}, err
	}

	if !token.Valid {
		return userModel.TokenOutputParse{}, errors.New("Ошибка: некорректный токен")
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return userModel.TokenOutputParse{}, errors.New("Ошибка: некорректный токен")
	}

	user, err := s.user.Get("uuid", claims.UsersId, true)
	if err != nil {
		return userModel.TokenOutputParse{}, err
	}

	authType, err := s.authType.Get("uuid", claims.AuthTypesId, true)
	if err != nil {
		return userModel.TokenOutputParse{}, err
	}

	return userModel.TokenOutputParse{
		UsersId:   user.Id,
		UsersUuid: claims.UsersId,
		AuthType:  *authType,
		TokenApi:  claims.TokenApi,
	}, nil
}

/* Parse token without validate check */
func (s *TokenService) ParseTokenWithoutValid(pToken, signingKey string) (userModel.TokenOutputParse, error) {
	token, err := jwt.ParseWithClaims(pToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(signingKey), nil
	})

	// Получение данных из токена (с преобразованием к указателю на tokenClaims)
	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return userModel.TokenOutputParse{}, errors.New("token claims are not of type")
	}

	user, err := s.user.Get("uuid", claims.UsersId, true)
	if err != nil {
		return userModel.TokenOutputParse{}, err
	}

	authType, err := s.authType.Get("uuid", claims.AuthTypesId, true)
	if err != nil {
		return userModel.TokenOutputParse{}, err
	}

	return userModel.TokenOutputParse{
		UsersId:  user.Id,
		AuthType: *authType,
		TokenApi: claims.TokenApi,
	}, nil
}

/* Структура тела токена для смены пароля пользователя (частный случай) */
type tokenResetClaims struct {
	jwt.StandardClaims
	UsersId string `json:"users_id"` // ID пользователя
	Email   string `json:"email"`    // Email пользователя
}

/* Parse reset token with validate check */
func (s *TokenService) ParseResetToken(pToken, signingKey string) (userModel.ResetTokenOutputParse, error) {
	token, err := jwt.ParseWithClaims(pToken, &tokenResetClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(signingKey), nil
	})
	if err != nil {
		return userModel.ResetTokenOutputParse{}, err
	}
	if !token.Valid {
		return userModel.ResetTokenOutputParse{}, errors.New("Ошибка: некорректный токен")
	}

	// Получение данных из токена (с преобразованием к указателю на tokenClaims)
	claims, ok := token.Claims.(*tokenResetClaims)
	if !ok {
		return userModel.ResetTokenOutputParse{}, errors.New("Ошибка: некорректный токен")
	}

	_, err = s.user.Get("email", claims.Email, true)
	if err != nil {
		return userModel.ResetTokenOutputParse{}, err
	}

	user, err := s.user.Get("uuid", claims.UsersId, true)
	if err != nil {
		return userModel.ResetTokenOutputParse{}, err
	}

	return userModel.ResetTokenOutputParse{
		UsersId: user.Id,
		Email:   claims.Email,
	}, nil
}
