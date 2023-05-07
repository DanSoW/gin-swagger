package auth

import "time"

const (
	SALT              = "a;ugb*(AW^GFA&WTVFawtfva79wf6g7a6f2r8tc127tVIYTAWCFA&(T"
	SIGNING_KEY       = "AOgnaiouGHA()wH8WFG8uga8eya7G9g9UBA@e@h(rh@u(!"
	TOKEN_TLL_ACCESS  = 1 * time.Hour
	TOKEN_TLL_REFRESH = 12 * time.Hour
	TOKEN_TLL_RESET   = 5 * time.Minute

	AUTH_TYPE_LOCAL  = "local"
	AUTH_TYPE_GOOGLE = "google"
)
