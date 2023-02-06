package project

import (
	"time"
)

/* Модель данных о проекте */
type ProjectCreateModel struct {
	Logo        *string          `json:"logo"`
	CompanyUuid string           `json:"uuid" binding:"required"`
	Title       string           `json:"title" binding:"required"`
	Description string           `json:"description" binding:"required"`
	Manager     ManagerInfoModel `json:"manager" binding:"required"`
}

/* Модель данных для обновления проекта */
type ProjectUpdateModel struct {
	Uuid        string           `json:"uuid" binding:"required"`
	Title       string           `json:"title" binding:"required"`
	Description string           `json:"description" binding:"required"`
	Manager     ManagerInfoModel `json:"manager" binding:"required"`
}

type ProjectImgModel struct {
	Filepath string `json:"filepath" binding:"required" db:"filepath"`
	Uuid     string `json:"uuid" binding:"required"`
}

type ProjectDataModel struct {
	Logo        *string          `json:"logo" db:"logo"`
	Title       string           `json:"title" binding:"required" db:"title"`
	Description string           `json:"description" binding:"required" db:"description"`
	Manager     ManagerInfoModel `json:"manager" binding:"required" db:"manager"`
}

type ProjectLowInfoModel struct {
	Uuid      string    `json:"uuid" db:"uuid"`
	Data      string    `json:"data" db:"data"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type ProjectDbDataEx struct {
	Uuid      string           `json:"uuid" db:"uuid"`
	Data      ProjectDataModel `json:"data" db:"data"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
}

type ProjectUuidModel struct {
	Uuid string `json:"uuid" binding:"required"`
}

type ProjectCountModel struct {
	Uuid  string `json:"uuid" binding:"required"`
	Limit int    `json:"limit" binding:"required"`
	Count int    `json:"count"`
}

type ProjectAnyCountModel struct {
	Projects []ProjectDbDataEx `json:"projects" binding:"required"`
	Count    int               `json:"count" binding:"required"`
}
