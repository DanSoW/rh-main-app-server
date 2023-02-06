package admin

import (
	"errors"
	"fmt"
	actionConstant "main-server/pkg/constant/action"
	tableConstant "main-server/pkg/constant/table"
	rbacModel "main-server/pkg/model/rbac"
	util "main-server/pkg/util"

	"github.com/jmoiron/sqlx"
)

/* Модель данных информации о менеджере*/
type SystemPermissionModel struct {
	Email          string            `json:"email" binding:"required"`
	RoleUuid       *string           `json:"role_uuid"`
	PermissionList []PermissionModel `json:"permission_list"`
}

/* Модель распространения ограничений на конкретный объект (или привелегий) */
type PermissionModel struct {
	ObjectUuid string   `json:"object_uuid" binding:"required"`
	ActionList []string `json:"action_list" binding:"required"`
}

func (smm *SystemPermissionModel) Check(db *sqlx.DB) (bool, error) {
	slice := actionConstant.GetSlice()

	query := fmt.Sprintf(`SELECT value FROM %s`, tableConstant.AC_OBJECTS)
	if len(smm.PermissionList) > 0 {
		for _, item := range smm.PermissionList {
			var objects []rbacModel.ObjectDbModel

			if err := db.Select(&objects, query, item.ObjectUuid); err != nil {
				return false, err
			}

			if len(objects) <= 0 {
				return false, errors.New(fmt.Sprintf("Ошибка: объекта с ID = %s нет в базе данных", item.ObjectUuid))
			}

			if len(item.ActionList) > 0 {
				for _, subItem := range item.ActionList {
					if exists, _ := util.InArray(subItem, slice); !exists {
						return false, errors.New(fmt.Sprintf("Ошибка: элемента со значением %s нет в доступных действиях", subItem))
					}
				}
			}
		}
	}

	return true, nil
}
