package config

import (
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"
)

type VkAuthConfig struct {
	VkAuth oauth2.Config
}

var AppVKAuthConfig VkAuthConfig

func InitVKAuthConfig() {
	AppOAuth2Config.GoogleLogin = oauth2.Config{
		ClientID:     viper.GetString("vk_oauth2.client_id"),
		ClientSecret: viper.GetString("vk_oauth2.client_secret"),
		Endpoint:     vk.Endpoint,
		RedirectURL:  "http://localhost:3000",
		Scopes:       []string{"account"},
	}
}
