package handler

import (
	middlewareConstant "main-server/pkg/constant/middleware"
	authHandler "main-server/pkg/handler/auth"
	serviceHandler "main-server/pkg/handler/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	_ "main-server/docs"

	service "main-server/pkg/service"

	_ "github.com/swaggo/files"
	swaggerFiles "github.com/swaggo/files"
	_ "github.com/swaggo/gin-swagger"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

/* Инициализация маршрутов */
func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	// Установка максимального размера тела Multipart
	router.MaxMultipartMemory = 50 << 20 // 50 MiB

	// Установка статической директории
	router.Static("/public", "./public")

	// Установка глобального каталога для хранения HTML-страниц
	router.LoadHTMLGlob("pkg/template/*")

	// Установка CORS-политик
	router.Use(cors.New(cors.Config{
		//AllowAllOrigins: true,
		AllowOrigins:     []string{viper.GetString("client_url")},
		AllowMethods:     []string{"POST", "GET"},
		AllowHeaders:     []string{"Origin", "Content-type", "Authorization"},
		AllowCredentials: true,
	}))

	// URL: /swagger/index.html
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Инициализация списка обработчиков в цепочке middleware
	middleware := make(map[string]func(c *gin.Context))
	middleware[middlewareConstant.MN_UI] = h.userIdentity
	middleware[middlewareConstant.MN_UI_LOGOUT] = h.userIdentityLogout

	// Инициализация маршрутов для сервиса service
	service := serviceHandler.NewServiceHandler(router, h.services)
	service.InitRoutes(&middleware)

	// Инициализация маршрутов для сервиса auth
	auth := authHandler.NewAuthHandler(router, h.services)
	auth.InitRoutes(&middleware)

	return router
}
