package route

const (
	AUTH_MAIN_ROUTE = "/auth"
)

const (
	// Локальная авторизация пользователя
	SIGN_IN              = "/sign-in"
	SIGN_UP              = "/sign-up"
	SIGN_UP_UPLOAD_IMAGE = "/sign-up/upload/image"

	// Авторизация через внешний сервис VK
	SIGN_IN_VK          = "/sign-in/vk"
	SIGN_IN_VK_CALLBACK = "/sign-in/vk/callback"

	// Авторизация через внешний сервис Google
	SIGN_IN_OAUTH2 = "/sign-in/oauth2"

	// Сброс пароля
	RECOVERY_PASSWORD = "/recovery/password"
	RESET_PASSWORD    = "/reset/password"

	// Дополнительные маршруты
	REFRESH       = "/refresh"
	LOGOUT        = "/logout"
	ACTIVATE_LINK = "/activate/:link"
)
