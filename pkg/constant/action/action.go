package action

const (
	// Action CRUD
	CREATE = "create"
	MODIFY = "modify"
	DELETE = "delete"
	READ   = "read"

	// Action system
	ADMINISTRATION = "administration"
	MANAGEMENT     = "management"
)

func GetSlice() []string {
	return []string{
		CREATE,
		MODIFY,
		DELETE,
		READ,
		ADMINISTRATION,
		MANAGEMENT,
	}
}
