package repository

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type WrapperPostgres struct {
	db *sqlx.DB
}

func NewWrapperPostgres(db *sqlx.DB) *WrapperPostgres {
	return &WrapperPostgres{
		db: db,
	}
}

func (r *WrapperPostgres) Get(table, column string, value interface{}, check bool) (*interface{}, error) {
	var model []interface{}
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1 LIMIT 1", table, column)

	var err error

	switch value.(type) {
	case int:
		err = r.db.Select(&model, query, value.(int))
		break
	case string:
		err = r.db.Select(&model, query, value.(string))
		break
	}

	if len(model) <= 0 {
		if check {
			return nil, errors.New(fmt.Sprintf("Ошибка: экземпляра таблицы %s по запросу %s:%s не найдено!", table, column, value))
		}

		return nil, nil
	}

	return &model[len(model)-1], err
}
