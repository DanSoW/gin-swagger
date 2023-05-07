package rbac

type TypesObjectsModel struct {
	Id          int     `json:"id" db:"id"`
	Value       string  `json:"value" db:"value"`
	Description string  `json:"description" db:"description"`
	TableName   string  `json:"table_name" db:"table_name"`
	UsersId     *string `json:"users_id" db:"users_id"`
}
