package auth

import (
	"context"
	"encoding/json"
	route "main-server/pkg/constant/route"
	userModel "main-server/pkg/model/user"
	"net/http"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

type VerifyEmailModel struct {
	VerifyEmail bool `json:"verified_email" binding:"required"`
}

func RefreshAccessToken(c context.Context, refreshToken string) (userModel.TokenDataModel, error) {
	url := route.OAUTH2_REFRESH_TOKEN_ROUTE + viper.GetString("oauth2.client_id")
	url = url + "&client_secret=" + viper.GetString("oauth2.client_secret")
	url = url + "&refresh_token=" + refreshToken + "&grant_type=refresh_token"

	response, err := http.Post(url, "application/x-www-form-urlencoded", nil)

	var j userModel.TokenDataModel

	err = json.NewDecoder(response.Body).Decode(&j)

	if err != nil {
		return userModel.TokenDataModel{}, err
	}

	return j, nil
}

func GetInfoToken(accessToken string) (interface{}, error) {
	response, err := http.Get(route.OAUTH2_TOKEN_INFO_ROUTE + accessToken)

	var j interface{}

	err = json.NewDecoder(response.Body).Decode(&j)

	if err != nil {
		return false, err
	}

	return j, nil
}

func VerifyAccessToken(accessToken string) (bool, error) {
	response, err := http.Get(route.OAUTH2_TOKEN_INFO_ROUTE + accessToken)

	var j VerifyEmailModel

	err = json.NewDecoder(response.Body).Decode(&j)

	if err != nil {
		return false, err
	}

	return j.VerifyEmail, nil
}

func GetUserInfo(token *oauth2.Token) (userModel.UserRegisterOAuth2Model, error) {
	response, err := http.Get(route.OAUTH2_USER_INFO_ROUTE + token.AccessToken)

	var data userModel.UserRegisterOAuth2Model

	err = json.NewDecoder(response.Body).Decode(&data)

	if err != nil {
		return userModel.UserRegisterOAuth2Model{}, err
	}

	return data, nil
}

func RevokeToken(accessToken string) (bool, error) {
	response, err := http.Post(
		route.OAUTH2_REVOKE_TOKEN_ROUTE+accessToken,
		"Content-type:application/x-www-form-urlencoded",
		nil)

	if err != nil {
		return false, err
	}

	return (response.StatusCode == 200), nil
}
