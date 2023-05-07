package util

func GeneratePolicies(userId, domainId, objectId string, actions []string) [][]string {
	var policies [][]string
	policies = make([][]string, len(actions))

	for _, item := range actions {
		policies = append(policies, []string{userId, domainId, objectId, item})
	}

	return policies
}
