package repository

import (
	"errors"
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
func (r *ObjectPostgres) Get(column string, value interface{}, check bool) (*rbacModel.ObjectDbModel, error) {
	var objects []rbacModel.ObjectDbModel
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1", tableConstant.AC_OBJECTS, column)

	var err error

	switch value.(type) {
	case int:
		err = r.db.Select(&objects, query, value.(int))
		break
	case string:
		err = r.db.Select(&objects, query, value.(string))
		break
	}

	if len(objects) <= 0 {
		if check {
			return nil, errors.New(fmt.Sprintf("Ошибка: объекта по запросу %s:%s не найдено!", column, value))
		}

		return nil, nil
	}

	return &objects[len(objects)-1], err
}

/* Метод добавления нового информационного ресурса в систему */
func (r *ObjectPostgres) AddResource(resource *rbacModel.ResourceModel) (*rbacModel.ObjectDbModel, error) {
	// Получение типа объекта
	typeObject, err := r.acTypeObject.Get("value", resource.TypeResource, true)
	if err != nil {
		return nil, err
	}

	var parentObject *rbacModel.ObjectDbModel
	parentObject = nil

	if resource.ParentUuid != nil {
		// Получение родительского объекта
		parentObject, err = r.Get("value", resource.ParentUuid, true)
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
