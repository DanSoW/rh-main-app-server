package repository

import (
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

func (r *TypeObjectPostgres) GetTypeObject(column, value interface{}) (rbacModel.TypeObjectDbModel, error) {
	var typeObject rbacModel.TypeObjectDbModel
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1", tableConstant.AC_TYPES_OBJECTS, column.(string))

	var err error

	switch value.(type) {
	case int:
		err = r.db.Get(&typeObject, query, value.(int))
		break
	case string:
		err = r.db.Get(&typeObject, query, value.(string))
		break
	}

	return typeObject, err
}
