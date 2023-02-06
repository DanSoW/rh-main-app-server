package worker

import "time"

type WorkerModel struct {
	Role string `json:"role" db:"role"`
}

type WorkerDbModel struct {
	Id          int       `json:"id" db:"id"`
	Uuid        string    `json:"uuid" db:"uuid"`
	Data        string    `json:"data" db:"data"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	UsersId     int       `json:"users_id" db:"users_id"`
	CompaniesId int       `json:"companies_id" db:"companies_id"`
}
