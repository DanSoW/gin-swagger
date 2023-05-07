package rbac

type ObjectDbModel struct {
	Id             int     `json:"id" binding:"required" db:"id"`
	Value          string  `json:"value" binding:"required" db:"value"`
	Description    *string `json:"description" binding:"required" db:"description"`
	ParentId       *int    `json:"parent_id" binding:"required" db:"parent_id"`
	TypesObjectsId string  `json:"types_objects_id" binding:"required" db:"types_objects_id"`
}
