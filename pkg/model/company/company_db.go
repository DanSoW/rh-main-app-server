package company

import "time"

/*
 * Модели, использующиеся для взаимодействия с таблицей cb_companies
 */

/* Основная модель */
type CompanyDbModel struct {
	Id        int       `json:"id" db:"id"`
	Uuid      string    `json:"uuid" db:"uuid"`
	Data      string    `json:"data" db:"data"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	UsersId   int       `json:"users_id" db:"users_id"`
}

/* Модель для атрибута data из структуры CompanyDbModel */
type CompanyDataModel struct {
	Logo         string `json:"logo" binding:"required"`
	Title        string `json:"title" binding:"required"`
	Description  string `json:"description" binding:"required"`
	Phone        string `json:"phone" binding:"required"`
	Link         string `json:"link" binding:"required"`
	EmailCompany string `json:"email_company" binding:"required"`
	EmailAdmin   string `json:"email_admin" binding:"required"`
}

/* Расширенная модель CompanyDbModel */
type CompanyDbExModel struct {
	Id        int            `json:"id"`
	Uuid      string         `json:"uuid"`
	Data      CompanyDbModel `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	UsersId   int            `json:"users_id"`
}
