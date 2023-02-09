package repository

import (
	"errors"
	"fmt"
	tableConstant "main-server/pkg/constant/table"
	rbacModel "main-server/pkg/model/rbac"

	"github.com/jmoiron/sqlx"
)

type TypeObjectPostgres struct {
	db *sqlx.DB
}

func NewTypeObjectPostgres(
	db *sqlx.DB,
) *TypeObjectPostgres {
	return &TypeObjectPostgres{
		db: db,
	}
}

func (r *TypeObjectPostgres) Get(column string, value interface{}, check bool) (*rbacModel.TypeObjectDbModel, error) {
	var types []rbacModel.TypeObjectDbModel
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1", tableConstant.AC_TYPES_OBJECTS, column)

	var err error

	switch value.(type) {
	case int:
		err = r.db.Select(&types, query, value.(int))
		break
	case string:
		err = r.db.Select(&types, query, value.(string))
		break
	}

	if len(types) <= 0 {
		if check {
			return nil, errors.New(fmt.Sprintf("Ошибка: типа объекта по запросу %s:%s не найдено!", column, value))
		}

		return nil, nil
	}

	return &types[len(types)-1], err
}
