package config

import (
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleLoginConfig struct {
	GoogleLogin oauth2.Config
}

var AppOAuth2Config GoogleLoginConfig

func InitOAuth2Config() {
	AppOAuth2Config.GoogleLogin = oauth2.Config{
		ClientID:     viper.GetString("oauth2.client_id"),
		ClientSecret: viper.GetString("oauth2.client_secret"),
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:3000",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	}
}
