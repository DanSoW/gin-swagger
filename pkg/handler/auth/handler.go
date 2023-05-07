package auth

import (
	_ "main-server/docs"

	middlewareConstant "main-server/pkg/constant/middleware"
	"main-server/pkg/constant/route"
	service "main-server/pkg/service"

	"github.com/gin-gonic/gin"
	_ "github.com/swaggo/files"
	_ "github.com/swaggo/gin-swagger"
)

type AuthHandler struct {
	rootHandler *gin.Engine
	services    *service.Service
}

func NewAuthHandler(root *gin.Engine, services *service.Service) *AuthHandler {
	return &AuthHandler{
		rootHandler: root,
		services:    services,
	}
}

/* Инициализация маршрутов для авторизации пользователя */
func (h *AuthHandler) InitRoutes(
	middleware *map[string]func(c *gin.Context),
) {
	// URL: /auth
	auth := h.rootHandler.Group(route.AUTH_MAIN_ROUTE)
	{
		// URL: /auth/sign-up
		auth.POST(route.SIGN_UP, h.signUp)

		// URL: /auth/sign-in
		auth.POST(route.SIGN_IN, h.signIn)

		// URL: /auth/sign-in/oauth2
		auth.POST(route.SIGN_IN_OAUTH2, h.signInOAuth2)

		// URL: /auth/activate/:link
		auth.GET(route.ACTIVATE_LINK, h.activate)

		// URL: /auth/refresh
		auth.POST(route.REFRESH, (*middleware)[middlewareConstant.MN_UI_LOGOUT], h.refresh)

		// URL: /auth/logout
		auth.POST(route.LOGOUT, (*middleware)[middlewareConstant.MN_UI_LOGOUT], h.logout)

		// URL: /auth/sign-up/upload/image
		auth.POST(route.SIGN_UP_UPLOAD_IMAGE, (*middleware)[middlewareConstant.MN_UI], h.uploadProfileImage)

		// URL: /auth/recovery/password
		auth.POST(route.RECOVERY_PASSWORD, h.recoveryPassword)

		// URL: /auth/reset/password
		auth.POST(route.RESET_PASSWORD, h.resetPassword)
	}
}
