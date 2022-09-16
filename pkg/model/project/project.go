package project

/* Model data for response users model */
type ProjectModel struct {
	Logo        *string          `json:"logo"`
	Uuid        string           `json:"uuid" binding:"required"`
	Title       string           `json:"title" binding:"required"`
	Description string           `json:"description" binding:"required"`
	Managers    []UserEmailModel `json:"managers" binding:"required"`
}

type ProjectDbModel struct {
	Logo        *string          `json:"logo" db:"logo"`
	Title       string           `json:"title" binding:"required" db:"title"`
	Description string           `json:"description" binding:"required" db:"description"`
	Managers    []UserEmailModel `json:"managers" binding:"required" db:"managers"`
}

type ProjectLogoModel struct {
	Filepath string `json:"filepath" binding:"required" db:"filepath"`
	Uuid     string `json:"uuid" binding:"required"`
}
