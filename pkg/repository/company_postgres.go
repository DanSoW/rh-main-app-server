package repository

import (
	"encoding/json"
	"fmt"
	roleConstant "main-server/pkg/constant/role"
	tableConstant "main-server/pkg/constant/table"
	companyModel "main-server/pkg/model/company"
	rbacModel "main-server/pkg/model/rbac"
	utilSlice "main-server/pkg/util"
	"sort"

	"github.com/casbin/casbin/v2"
	"github.com/jmoiron/sqlx"
)

type CompanyPostgres struct {
	db       *sqlx.DB
	enforcer *casbin.Enforcer
	role     *RolePostgres
}

/* Function for create new struct of CompanyPostgres */
func NewCompanyPostgres(db *sqlx.DB, enforcer *casbin.Enforcer, role *RolePostgres) *CompanyPostgres {
	return &CompanyPostgres{
		db:       db,
		enforcer: enforcer,
		role:     role,
	}
}

func (r *CompanyPostgres) GetManagers(userId, domainId int, data companyModel.ManagerCountModel) (companyModel.ManagerAnyCountModel, error) {
	query := fmt.Sprintf("SELECT id FROM %s WHERE uuid=$1", tableConstant.COMPANIES_TABLE)
	var companyId int

	row := r.db.QueryRow(query, data.Uuid)
	if err := row.Scan(&companyId); err != nil {
		return companyModel.ManagerAnyCountModel{}, err
	}

	var managers []companyModel.ManagerDbDataEx
	sum := (data.Count + data.Limit)

	query = fmt.Sprintf(`
	SELECT DISTINCT tl.uuid, tr.data, tw.created_at FROM %s tl
	JOIN %s tr ON tr.users_id = tl.id
	JOIN %s tw ON tw.users_id = tl.id
	JOIN %s twp ON twp.workers_id = tw.id
	WHERE tw.data = $1 AND tw.companies_id = $2
	`,
		tableConstant.USERS_TABLE, tableConstant.USERS_DATA_TABLE,
		tableConstant.WORKERS_TABLE, tableConstant.WORKERS_PROJECTS_TABLE,
	)

	role, err := r.role.GetRole("value", roleConstant.ROLE_BUILDER_MANAGER)
	if err != nil {
		return companyModel.ManagerAnyCountModel{}, err
	}

	roleData := rbacModel.RoleDataModel{Role: role.Uuid}
	roleDataStr, err := json.Marshal(roleData)
	if err != nil {
		return companyModel.ManagerAnyCountModel{}, err
	}

	err = r.db.Select(&managers, query, roleDataStr, companyId)
	if err != nil {
		return companyModel.ManagerAnyCountModel{}, err
	}

	var managersEx []companyModel.ManagerDataEx

	for _, element := range managers {
		var managerData companyModel.ManagerUserData
		err = json.Unmarshal([]byte(element.Data), &managerData)

		if err != nil {
			return companyModel.ManagerAnyCountModel{}, err
		}

		managersEx = append(managersEx, companyModel.ManagerDataEx{
			Uuid:      element.Uuid,
			Data:      managerData,
			CreatedAt: element.CreatedAt,
		})
	}

	// Remove duplicates
	managersEx = utilSlice.RemoveDuplicate(managersEx)

	sort.SliceStable(managersEx, func(i, j int) bool {
		return managersEx[i].CreatedAt.After(managersEx[j].CreatedAt)
	})

	if data.Count >= len(managersEx) {
		return companyModel.ManagerAnyCountModel{}, nil
	}

	if sum >= len(managersEx) {
		sum -= (sum - len(managersEx))
	}

	return companyModel.ManagerAnyCountModel{
		Managers: managersEx[data.Count:sum],
		Count:    (sum - data.Count),
	}, nil
}
