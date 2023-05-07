package middleware

const (
	AUTHORIZATION_HEADER = "Authorization"
	USER_CTX             = "users_id"
	USER_UUID_CTX        = "users_uuid"
	AUTH_TYPE_VALUE_CTX  = "auth_type_value"
	ACCESS_TOKEN_CTX     = "access_token"
	TOKEN_API_CTX        = "token_api"
	DOMAINS_ID           = "domains_id"
	DOMAINS_UUID         = "domains_uuid"

	MN_UI                                            = "ui"
	MN_UI_LOGOUT                                     = "ui_logout"
	MN_UI_HAS_ROLE                                   = "ui_has_role"
	MN_UI_HAS_ROLES_ADMIN_OR_BUILDER_ADMIN           = "ui_has_roles_admin_or_builder_admin"
	MN_UI_HAS_ROLES_BUILDER_MANAGER_OR_BUILDER_ADMIN = "ui_has_roles_builder_manager_or_builder_admin"
	MN_UI_HAS_ROLE_ADMIN                             = "ui_has_role_admin"
	MN_UI_HAS_ROLE_MANAGER                           = "ui_has_role_manager"
	MN_UI_HAS_ROLE_BUILDER_MANAGER                   = "ui_has_role_builder_manager"
	MN_UI_HAS_ROLE_BUILDER_ADMIN                     = "ui_has_role_builder_admin"
)
