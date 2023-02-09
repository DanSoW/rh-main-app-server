package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	actionConstant "main-server/pkg/constant/action"
	tableConstant "main-server/pkg/constant/table"
	companyModel "main-server/pkg/model/company"
	projectModel "main-server/pkg/model/project"
	userModel "main-server/pkg/model/user"
	"sort"
	"strconv"

	"github.com/casbin/casbin/v2"
	"github.com/jmoiron/sqlx"
)

type CompanyPostgres struct {
	db       *sqlx.DB
	enforcer *casbin.Enforcer
	role     *RolePostgres
	user     *UserPostgres
	wrapper  *WrapperPostgres
}

/* Function for create new struct of CompanyPostgres */
func NewCompanyPostgres(
	db *sqlx.DB,
	enforcer *casbin.Enforcer,
	role *RolePostgres,
	user *UserPostgres,
	wrapper *WrapperPostgres,
) *CompanyPostgres {
	return &CompanyPostgres{
		db:       db,
		enforcer: enforcer,
		role:     role,
		user:     user,
		wrapper:  wrapper,
	}
}

/* Получение списка менеджеров компании */
func (r *CompanyPostgres) GetManagers(userId, domainId int, data companyModel.ManagerCountModel) (companyModel.ManagerAnyCountModel, error) {
	query := fmt.Sprintf("SELECT id FROM %s WHERE uuid=$1", tableConstant.CB_COMPANIES)

	// Идентификатор компании
	var companyId int

	row := r.db.QueryRow(query, data.Uuid)
	if err := row.Scan(&companyId); err != nil {
		return companyModel.ManagerAnyCountModel{}, err
	}

	var managers []companyModel.ManagerDbDataEx
	sum := (data.Count + data.Limit)

	query = fmt.Sprintf(`
			SELECT DISTINCT u.uuid, u.email, ud.data, w.created_at FROM %s u
			JOIN %s ud ON ud.users_id = u.id
			JOIN %s w ON w.users_id = u.id
			WHERE w.companies_id = $1
		`,
		tableConstant.U_USERS, tableConstant.U_USERS_DATA, tableConstant.CB_WORKERS,
	)

	err := r.db.Select(&managers, query, companyId)
	if err != nil {
		return companyModel.ManagerAnyCountModel{}, err
	}

	// Если количество полученных менеджеров больше или равно текущему числу
	// получаемых менеджеров, то возвращаем пустую структуру
	if data.Count >= len(managers) {
		return companyModel.ManagerAnyCountModel{}, nil
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
			Email:     element.Email,
			Data:      managerData,
			CreatedAt: element.CreatedAt,
		})
	}

	// Удаление дубликатов (реализация должна быть в SQL)
	// managersEx = utilSlice.RemoveDuplicate(managersEx)

	// Сортировка списка менеджеров, по дате создания
	sort.SliceStable(managersEx, func(i, j int) bool {
		return managersEx[i].CreatedAt.After(managersEx[j].CreatedAt)
	})

	if sum >= len(managersEx) {
		sum -= (sum - len(managersEx))
	}

	return companyModel.ManagerAnyCountModel{
		Managers: managersEx[data.Count:sum],
		Count:    (sum - data.Count),
	}, nil
}

/* Method for update image for current company */
func (r *CompanyPostgres) CompanyUpdateImage(user userModel.UserIdentityModel, data companyModel.CompanyImageModel) (companyModel.CompanyImageModel, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return companyModel.CompanyImageModel{}, err
	}

	userIdStr := strconv.Itoa(user.UserId)
	domainIdStr := strconv.Itoa(user.DomainId)

	// Check access for user
	access, err := r.enforcer.Enforce(userIdStr, domainIdStr, data.Uuid, actionConstant.MODIFY)
	if err != nil {
		tx.Rollback()
		return companyModel.CompanyImageModel{}, err
	}

	if !access {
		tx.Rollback()
		return companyModel.CompanyImageModel{}, errors.New("Ошибка! Нет доступа!")
	}

	// Update logo for project
	query := fmt.Sprintf(`UPDATE %s tl SET data = jsonb_set(data, '{logo}', to_jsonb($1::text), true) WHERE tl.uuid = $2`, tableConstant.CB_COMPANIES)

	_, err = r.db.Exec(query, data.Filepath, data.Uuid)
	if err != nil {
		tx.Rollback()
		return companyModel.CompanyImageModel{}, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return companyModel.CompanyImageModel{}, err
	}

	return data, nil
}

/* Method for update info for current company */
func (r *CompanyPostgres) CompanyUpdate(user userModel.UserIdentityModel, data companyModel.CompanyUpdateModel) (companyModel.CompanyUpdateModel, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return companyModel.CompanyUpdateModel{}, err
	}

	userIdStr := strconv.Itoa(user.UserId)
	domainIdStr := strconv.Itoa(user.DomainId)

	// Check access for user
	access, err := r.enforcer.Enforce(userIdStr, domainIdStr, data.Uuid, actionConstant.MODIFY)
	if err != nil {
		tx.Rollback()
		return companyModel.CompanyUpdateModel{}, err
	}

	if !access {
		tx.Rollback()
		return companyModel.CompanyUpdateModel{}, errors.New("Ошибка! Нет доступа!")
	}

	var companyInfo []companyModel.CompanyInfoModel
	query := fmt.Sprintf("SELECT uuid, data, created_at FROM %s WHERE uuid=$1 LIMIT 1", tableConstant.CB_COMPANIES)

	err = r.db.Select(&companyInfo, query, data.Uuid)
	if err != nil {
		return companyModel.CompanyUpdateModel{}, err
	}

	var companyData companyModel.CompanyModel
	err = json.Unmarshal([]byte(companyInfo[0].Data), &companyData)

	companyData.Description = data.Description
	companyData.EmailCompany = data.EmailCompany
	companyData.Link = data.Link
	companyData.Phone = data.Phone
	companyData.Title = data.Title

	companyDataJson, err := json.Marshal(companyData)
	if err != nil {
		tx.Rollback()
		return companyModel.CompanyUpdateModel{}, err
	}

	query = fmt.Sprintf("UPDATE %s tl SET data=$1 WHERE tl.uuid=$2", tableConstant.CB_COMPANIES)

	_, err = tx.Exec(query, companyDataJson, data.Uuid)
	if err != nil {
		tx.Rollback()
		return companyModel.CompanyUpdateModel{}, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return companyModel.CompanyUpdateModel{}, err
	}

	return data, nil
}

type SQLResultGetManager struct {
	Id int `json:"id" binding:"required" db:"id"`
}

type SQLResultListProject struct {
	Uuid string `json:"uuid" binding:"required" db:"uuid"`
	Data string `json:"data" binding:"required" db:"data"`
}

/* Метод структуры для получения информации о менеджере */
func (r *CompanyPostgres) GetManager(user userModel.UserIdentityModel, data companyModel.ManagerUuidModel) (companyModel.ManagerCompanyModel, error) {

	// Информация о менеджере как о пользователе
	manager, err := r.user.Get("uuid", data.ManagerUuid, true)
	if err != nil {
		return companyModel.ManagerCompanyModel{}, err
	}

	// Информация о всех проектах, в которых участвует менеджер
	var sqlResultProject []SQLResultListProject

	query := fmt.Sprintf(
		`SELECT p.uuid, p.data FROM %s w
		INNER JOIN %s p ON p.workers_id = w.id
		WHERE w.users_id = $1;`,
		tableConstant.CB_WORKERS,
		tableConstant.CB_PROJECTS,
	)

	if err := r.db.Select(&sqlResultProject, query, manager.Id); err != nil {
		return companyModel.ManagerCompanyModel{}, err
	}

	var projects []companyModel.ManagerProjectInfoModel

	for _, element := range sqlResultProject {
		var data projectModel.ProjectDataModel
		err := json.Unmarshal([]byte(element.Data), &data)
		if err != nil {
			return companyModel.ManagerCompanyModel{}, err
		}

		projects = append(projects, companyModel.ManagerProjectInfoModel{
			Uuid:        element.Uuid,
			Logo:        data.Logo,
			Title:       data.Title,
			Description: data.Description,
		})
	}

	return companyModel.ManagerCompanyModel{
		Projects: projects,
	}, nil
}

func (r *CompanyPostgres) Get(column string, value interface{}, check bool) (*companyModel.CompanyDbModel, error) {
	var companies []companyModel.CompanyDbModel
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1", tableConstant.CB_COMPANIES, column)

	var err error

	switch value.(type) {
	case int:
		err = r.db.Select(&companies, query, value.(int))
		break
	case string:
		err = r.db.Select(&companies, query, value.(string))
		break
	}

	if len(companies) <= 0 {
		if check {
			return nil, errors.New(fmt.Sprintf("Ошибка: компании по запросу %s:%s не найдено!", column, value))
		}

		return nil, nil
	}

	return &companies[len(companies)-1], err
}

func (r *CompanyPostgres) GetEx(column string, value interface{}, check bool) (*companyModel.CompanyDbExModel, error) {
	var companies []companyModel.CompanyDbModel
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1", tableConstant.CB_COMPANIES, column)

	var err error

	switch value.(type) {
	case int:
		err = r.db.Select(&companies, query, value.(int))
		break
	case string:
		err = r.db.Select(&companies, query, value.(string))
		break
	}

	if len(companies) <= 0 {
		if check {
			return nil, errors.New(fmt.Sprintf("Ошибка: компании по запросу %s:%s не найдено!", column, value))
		}

		return nil, nil
	}

	var data companyModel.CompanyDataModel
	company := companies[len(companies)-1]

	err = json.Unmarshal([]byte(company.Data), &data)
	if err != nil {
		return nil, err
	}

	return &companyModel.CompanyDbExModel{
		Id:        company.Id,
		Uuid:      company.Uuid,
		Data:      data,
		CreatedAt: company.CreatedAt,
		UpdatedAt: company.UpdatedAt,
		UsersId:   company.UsersId,
	}, err
}

func (r *CompanyPostgres) GetByWorker(id int, check bool) ([]companyModel.CompanyDbExModel, error) {
	var companies []companyModel.CompanyDbExModel
	query := fmt.Sprintf(
		`SELECT c.* FROM %s c
		INNER JOIN %s w ON c.id = w.companies_id
		WHERE w.id = $1;`,
		tableConstant.CB_COMPANIES,
		tableConstant.CB_WORKERS,
	)

	var err error

	err = r.db.Select(&companies, query, id)

	if len(companies) <= 0 {
		if check {
			return nil, errors.New(fmt.Sprintf("Ошибка: компаний по запросу id:%d не найдено!", id))
		}

		return nil, nil
	}

	return companies, err
}
