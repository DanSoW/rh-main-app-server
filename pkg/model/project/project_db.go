package project

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

/* Расширенная модель данных (полная информация о проекте для чтения) */
type ProjectDbModel struct {
	Uuid        string             `json:"uuid"`
	Data        ProjectDataDbModel `json:"data"`
	WorkersId   int                `json:"workers_id"`
	CompaniesId int                `json:"companies_id"`
}

/* Модель данных для парсинга data из из структуры ProjectDbModel*/
type ProjectDataDbModel struct {
	Logo        string `json:"logo" db:"logo"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`
}

/* Переопределение метода для получения структуры из JSON-строки */
func (pdm *ProjectDbModel) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		json.Unmarshal(v, &pdm)
		return nil
	case string:
		json.Unmarshal([]byte(v), &pdm)
		return nil
	default:
		return errors.New(fmt.Sprintf("Неподдерживаемый тип: %T", v))
	}
}

/* Переопределение метода для получения JSON-строки из структуры */
func (pdm *ProjectDbModel) Value() (driver.Value, error) {
	return json.Marshal(&pdm)
}
