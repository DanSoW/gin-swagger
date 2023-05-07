package email

/* Структура описывающая модель для работы с STMP (фундаментальное тело сообщения) */
type Mail struct {
	Sender  string
	To      []string
	Subject string
	Body    string
}

/* Структура, описывающая полное содержимое сообщения пользователя */
type MessageInputModel struct {
	UuidReceivers []string `json:"uuid_receiver" binding:"required"` // Получатели сообщения
	Subject       string   `json:"subject" binding:"required"`       // Тема сообщения
	Message       string   `json:"message" binding:"required"`       // Тело сообщения
}

/* Структура, описывающая полное содержимое сообщения пользователя */
type MessageOutputModel struct {
	Sender  string `json:"sender" binding:"required"`  // Отправитель сообщения
	Subject string `json:"subject" binding:"required"` // Тема сообщения
	Message string `json:"message" binding:"required"` // Тело сообщения
}
