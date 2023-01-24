package excel

type DocumentIdModel struct {
	DocumentId string `json:"document_id" binding:"required"`
}
