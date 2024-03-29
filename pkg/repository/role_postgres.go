package repository

import (
	"errors"
	"fmt"
	tableConstant "main-server/pkg/constant/table"
	rbacModel "main-server/pkg/model/rbac"
	"strconv"

	"github.com/casbin/casbin/v2"
	"github.com/jmoiron/sqlx"
)

type RolePostgres struct {
	db       *sqlx.DB
	enforcer *casbin.Enforcer
}

/* Создание нового экземпляра структуры RolePostgres */
func NewRolePostgres(db *sqlx.DB, enforcer *casbin.Enforcer) *RolePostgres {
	return &RolePostgres{
		db:       db,
		enforcer: enforcer,
	}
}

/* Получение определённой роли пользователя */
func (r *RolePostgres) Get(column string, value interface{}, check bool) (*rbacModel.RoleModel, error) {
	var roles []rbacModel.RoleModel
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1", tableConstant.AC_ROLES, column)

	var err error

	switch value.(type) {
	case int:
		err = r.db.Select(&roles, query, value.(int))
		break
	case string:
		err = r.db.Select(&roles, query, value.(string))
		break
	}

	if len(roles) <= 0 {
		if check {
			return nil, errors.New(fmt.Sprintf("Ошибка: роли по запросу %s:%s не найдено!", column, value))
		}

		return nil, nil
	}

	return &roles[len(roles)-1], err
}

/* Проверка присутствия у пользователя определённой роли (принадлежность к группе пользователей) */
func (r *RolePostgres) HasRole(usersId, domainsId int, roleValue string) (bool, error) {
	data, err := r.Get("value", roleValue, true)
	if err != nil {
		return false, err
	}

	r.enforcer.LoadPolicy()
	has, err := r.enforcer.HasRoleForUser(
		strconv.Itoa(usersId),
		strconv.Itoa(data.Id),
		strconv.Itoa(domainsId),
	)

	return has, err
}
