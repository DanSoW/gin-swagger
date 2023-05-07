package service

import (
	"errors"
	userModel "main-server/pkg/model/user"
	repository "main-server/pkg/repository"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

/* Structure for current repository */
type AuthService struct {
	repo         repository.Authorization
	tokenService TokenService
}

/* Function for create a new repository */
func NewAuthService(repo repository.Authorization, tokenService TokenService) *AuthService {
	return &AuthService{
		repo:         repo,
		tokenService: tokenService,
	}
}

/* Create user */
func (s *AuthService) CreateUser(user userModel.UserSignUpModel) (userModel.UserAuthDataModel, error) {
	return s.repo.CreateUser(user)
}

/* Upload profile image */
func (s *AuthService) UploadProfileImage(c *gin.Context, filepath string) (bool, error) {
	return s.repo.UploadProfileImage(c, filepath)
}

/* Login user */
func (s *AuthService) LoginUser(user userModel.UserSignInModel) (userModel.UserAuthDataModel, error) {
	return s.repo.LoginUser(user)
}

/* Login user with Google OAuth2 */
func (s *AuthService) LoginUserOAuth2(code string) (userModel.UserAuthDataModel, error) {
	return s.repo.LoginUserOAuth2(code)
}

/**
 * Функция для обновления токена доступа по токену обновления
 * @param {userModel.TokenLogoutDataModel} data - Подробная информация об авторизационной информации пользователя
 * @param {string} refreshToken - Токен обновления
 * @returns {userModel.UserAuthDataModel, error} Пара токенов (access и refresh) или ошибка
 */
func (s *AuthService) Refresh(data userModel.TokenLogoutDataModel, refreshToken string) (userModel.UserAuthDataModel, error) {
	token, err := s.tokenService.ParseTokenWithoutValid(refreshToken, viper.GetString("token.signing_key_refresh"))

	if err != nil {
		return userModel.UserAuthDataModel{}, err
	}

	return s.repo.Refresh(data, refreshToken, token)
}

/* Logout user */
func (s *AuthService) Logout(tokens userModel.TokenLogoutDataModel) (bool, error) {
	return s.repo.Logout(tokens)
}

/* Activation account of user */
func (s *AuthService) Activate(link string) (bool, error) {
	return s.repo.Activate(link)
}

/* Recover password */
func (s *AuthService) RecoveryPassword(email string) (bool, error) {
	return s.repo.RecoveryPassword(email)
}

/* Reset password */
func (s *AuthService) ResetPassword(data userModel.ResetPasswordModel) (bool, error) {
	token, err := s.tokenService.ParseResetToken(data.Token, viper.GetString("token.signing_key_reset"))

	if err != nil {
		return false, errors.New("Некорректный токен сброса пароля")
	}

	return s.repo.ResetPassword(data, token)
}
