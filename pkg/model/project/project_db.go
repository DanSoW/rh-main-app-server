package project

/* Модель данных проекта (в БД) */
type ProjectDbModel struct {
	Id          int    `json:"id" db:"id"`
	Uuid        string `json:"uuid" db:"uuid"`
	Data        string `json:"data" db:"data"`
	WorkersId   int    `json:"workers_id" db:"workers_id"`
	CompaniesId int    `json:"companies_id" db:"companies_id"`
}

/* Модель данных для парсинга data из из структуры ProjectDbModel*/
type ProjectDataDbModel struct {
	Logo        string `json:"logo" db:"logo"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`
}

/* Расширенная модель данных (полная информация о проекте для чтения) */
type ProjectExModel struct {
	Uuid        string             `json:"uuid"`
	Data        ProjectDataDbModel `json:"data"`
	WorkersId   int                `json:"workers_id"`
	CompaniesId int                `json:"companies_id"`
}
