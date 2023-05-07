package rbac

type ResourceModel struct {
	Resource     ResourceInfoModel `json:"resource" binding:"required"`
	ParentUuid   *string           `json:"parent_uuid" binding:"required"`
	TypeResource string            `json:"type_resource" binding:"required"`
}

type ResourceInfoModel struct {
	ResourceUuid string `json:"resource_uuid" binding:"required"`
	Description  string `json:"description" binding:"required"`
}
