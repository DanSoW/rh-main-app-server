package repository

import (
	"fmt"
	tableConstants "main-server/pkg/constant/table"
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
func (r *RolePostgres) GetRole(column, value interface{}) (rbacModel.RoleModel, error) {
	var user rbacModel.RoleModel
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1", tableConstants.AC_ROLES, column.(string))

	var err error

	switch value.(type) {
	case int:
		err = r.db.Get(&user, query, value.(int))
		break
	case string:
		err = r.db.Get(&user, query, value.(string))
		break
	}

	return user, err
}

/* Проверка присутствия у пользователя определённой роли (принадлежность к группе пользователей) */
func (r *RolePostgres) HasRole(usersId, domainsId int, roleValue string) (bool, error) {
	data, err := r.GetRole("value", roleValue)

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
