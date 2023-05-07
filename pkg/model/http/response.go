package http

type ResponseMessage struct {
	Message string `json:"message" binding:"required"`
}

type ResponseStatus struct {
	Status string `json:"status"`
}

type ResponseValue struct {
	Value bool `json:"value"`
}
