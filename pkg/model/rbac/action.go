package rbac

type ActionsModel struct {
	Id          int     `json:"id" db:"id"`
	Value       string  `json:"value" db:"value"`
	Description string  `json:"description" db:"description"`
	UsersId     *string `json:"users_id" db:"users_id"`
	DomainsId   *int    `json:"domains_id" db:"domains_id"`
}
