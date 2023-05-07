package resource

/* Модель для представления изображений */
type ImageModel struct {
	Filename string `json:"filename" binding:"required"`
	Filepath string `json:"filepath" binding:"required"`
}
