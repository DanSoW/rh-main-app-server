package company

import "time"

/* Models for local */
type ManagerCountModel struct {
	Uuid  string `json:"uuid" binding:"required"`
	Limit int    `json:"limit" binding:"required"`
	Count int    `json:"count"`
}

type ManagerAnyCountModel struct {
	Managers []ManagerDataEx `json:"managers" binding:"required"`
	Count    int             `json:"count" binding:"required"`
}

type ManagerDataEx struct {
	Uuid      string          `json:"uuid" binding:"required"`
	Data      ManagerUserData `json:"data" binding:"required"`
	CreatedAt time.Time       `jsong:"created_at" binding:"required" db:"created_at"`
}

type ManagerUserData struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Nickname   string `json:"nickname"`
	Patronymic string `json:"patronymic"`
	Position   string `json:"position"`
	Avatar     string `json:"avatar"`
}

/* Models from Database */
type ManagerAnyCountDbModel struct {
	Managers []ManagerDbDataEx `json:"managers" binding:"required" db:"managers"`
}

type ManagerDbDataEx struct {
	Uuid      string    `json:"uuid" binding:"required" db:"uuid"`
	Data      string    `json:"data" binding:"required" db:"data"`
	CreatedAt time.Time `jsong:"created_at" binding:"required" db:"created_at"`
}
