package repository

import (
	"fmt"
	tableConstant "main-server/pkg/constant/table"
	rbacModel "main-server/pkg/model/rbac"

	"github.com/jmoiron/sqlx"
)

type ObjectPostgres struct {
	db           *sqlx.DB
	acTypeObject *TypeObjectPostgres
}

func NewObjectPostgres(
	db *sqlx.DB,
	acTypeObject *TypeObjectPostgres,
) *ObjectPostgres {
	return &ObjectPostgres{
		db:           db,
		acTypeObject: acTypeObject,
	}
}

/* Метод получения информационного ресурса из системы */
func (r *ObjectPostgres) GetObject(column, value interface{}) (*rbacModel.ObjectDbModel, error) {
	var object rbacModel.ObjectDbModel
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1", tableConstant.AC_OBJECTS, column.(string))

	var err error

	switch value.(type) {
	case int:
		err = r.db.Get(&object, query, value.(int))
		break
	case string:
		err = r.db.Get(&object, query, value.(string))
		break
	}

	return &object, err
}

/* Метод добавления нового информационного ресурса в систему */
func (r *ObjectPostgres) AddResource(resource *rbacModel.ResourceModel) (*rbacModel.ObjectDbModel, error) {
	typeObject, err := r.acTypeObject.GetTypeObject("value", resource.TypeResource)
	if err != nil {
		return nil, err
	}

	var parentObject *rbacModel.ObjectDbModel
	parentObject = nil

	if resource.ParentUuid != nil {
		parentObject, err = r.GetObject("value", resource.ParentUuid)
		if err != nil {
			return nil, err
		}
	}

	var parentId *int
	parentId = nil

	if parentObject != nil {
		parentId = &parentObject.Id
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (value, description, parent_id, types_objects_id) 
		values ($1, $2, $3, $4) RETURNING id`,
		tableConstant.AC_OBJECTS,
	)

	var objectId int
	row := tx.QueryRow(query, resource.Resource.ResourceUuid, resource.Resource.Description, parentId, typeObject.Id)
	if err := row.Scan(&objectId); err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	var object rbacModel.ObjectDbModel
	object.Id = objectId
	object.Value = resource.Resource.ResourceUuid
	object.Description = resource.Resource.Description
	object.ParentId = parentId
	object.TypesObjectsId = typeObject.Id

	return &object, nil
}
