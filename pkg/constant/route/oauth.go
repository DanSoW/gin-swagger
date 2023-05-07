package route

const (
	// Token info
	OAUTH2_TOKEN_INFO_ROUTE = "https://www.googleapis.com/oauth2/v1/tokeninfo?access_token="

	// User info
	OAUTH2_USER_INFO_ROUTE = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

	// Revoke token
	OAUTH2_REVOKE_TOKEN_ROUTE = "https://oauth2.googleapis.com/revoke?token="

	// Refresh token
	OAUTH2_REFRESH_TOKEN_ROUTE = "https://oauth2.googleapis.com/token?client_id="
)
