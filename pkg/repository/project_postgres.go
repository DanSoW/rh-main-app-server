package repository

import (
	//middlewareConstant "main-server/pkg/constant/middleware"
	"encoding/json"
	"errors"
	"fmt"
	actionConstant "main-server/pkg/constant/action"
	objectConstant "main-server/pkg/constant/object"
	roleConstant "main-server/pkg/constant/role"
	tableConstant "main-server/pkg/constant/table"
	projectModel "main-server/pkg/model/project"
	"main-server/pkg/model/rbac"
	rbacModel "main-server/pkg/model/rbac"
	userModel "main-server/pkg/model/user"
	workerModel "main-server/pkg/model/worker"
	"sort"

	"strconv"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
)

type ProjectPostgres struct {
	db       *sqlx.DB
	enforcer *casbin.Enforcer
	role     *RolePostgres
	user     *UserPostgres
	object   *ObjectPostgres
	company  *CompanyPostgres
}

/* Функция создания нового экземпляра структуры ProjectPostgres */
func NewProjectPostgres(
	db *sqlx.DB,
	enforcer *casbin.Enforcer,
	role *RolePostgres,
	user *UserPostgres,
	object *ObjectPostgres,
	company *CompanyPostgres,
) *ProjectPostgres {
	return &ProjectPostgres{
		db:       db,
		enforcer: enforcer,
		role:     role,
		user:     user,
		object:   object,
		company:  company,
	}
}

/* Функция создания нового проекта */
func (r *ProjectPostgres) CreateProject(userId, domainId int, data projectModel.ProjectCreateModel) (projectModel.ProjectCreateModel, error) {
	// Начало транзакции
	tx, err := r.db.Begin()
	if err != nil {
		return projectModel.ProjectCreateModel{}, err
	}

	// Получение информации о компании
	company, err := r.company.GetEx("uuid", data.CompanyUuid)
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectCreateModel{}, err
	}

	/* Процесс определения менеджера, который будет менеджером для данного проекта */
	// Определение пользователя в системе
	manager, err := r.user.GetUser("email", data.Manager.Email)
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectCreateModel{}, err
	}

	// Определяем, присутствует ли пользователь в других компаниях
	query := fmt.Sprintf("SELECT * FROM %s WHERE companies_id != $1")
	var workers []workerModel.WorkerDbModel
	err = r.db.Select(&workers, query, company.Id)
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectCreateModel{}, err
	}

	if len(workers) > 0 {
		tx.Rollback()
		return projectModel.ProjectCreateModel{}, errors.New("Ошибка: данный пользователь уже является менеджеров в другой компании")
	}

	// Проверка на существование пользователя в системе
	query = fmt.Sprintf("SELECT * FROM %s WHERE companies_id=$1 and users_id=$2", tableConstant.CB_WORKERS)
	err = r.db.Select(&workers, query, company.Id, manager.Id)
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectCreateModel{}, err
	}

	// Идентификатор работника (worker'a)
	var workerId int

	// Если пользователь ещё не был добавлен в компанию (ещё обычный пользователь)
	if len(workers) <= 0 {
		// Добавление новой записи работника в систему
		query = fmt.Sprintf(`
				INSERT INTO %s (uuid, data, created_at, updated_at, users_id, companies_id) 
				values ($1, $2, $3, $4, $5, $6) RETURNING id`,
			tableConstant.CB_WORKERS,
		)

		row := tx.QueryRow(query, uuid.NewV4().String(), "", time.Now(), time.Now(), manager.Id, company.Id)
		if err := row.Scan(&workerId); err != nil {
			tx.Rollback()
			return projectModel.ProjectCreateModel{}, err
		}
	} else {
		// Получение id работника
		workerId = workers[len(workers)-1].Id
	}

	/* Процесс добавления информации о проекте в компанию*/
	// Добавление информации о проекте
	query = fmt.Sprintf("INSERT INTO %s (uuid, data, created_at, updated_at, workers_id, companies_id) values ($1, $2, $3, $4, $5, $6);", tableConstant.CB_PROJECTS)

	dataJson, err := json.Marshal(projectModel.ProjectDataDbModel{
		Logo:        *data.Logo,
		Title:       data.Title,
		Description: data.Description,
	})

	if err != nil {
		tx.Rollback()
		return projectModel.ProjectCreateModel{}, err
	}

	// Добавление информации о проекте в БД
	projectUuid := uuid.NewV4()
	row := tx.QueryRow(query, projectUuid, dataJson, time.Now(), time.Now(), workerId, company.Id)
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectCreateModel{}, err
	}

	parentObject, err := r.object.GetObject("value", company.Uuid)
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectCreateModel{}, err
	}

	var resource rbacModel.ResourceModel
	resource.ParentUuid = &parentObject.Value
	resource.TypeResource = objectConstant.PROJECT
	resource.Resource.ResourceUuid = projectUuid.String()
	resource.Resource.Description = fmt.Sprintf("Проект компании %s", company.Data.Title)

	role, err := r.role.GetRole("value", roleConstant.ROLE_BUILDER_MANAGER)
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectCreateModel{}, err
	}

	_, err = r.object.AddResource(&resource)
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectCreateModel{}, err
	}

	// Модель для представления группы и информационного ресурса, в рамках которого есть определённая группа
	var gpsm rbac.GPSubjectModel
	gpsm.RoleId = role.Id
	gpsm.ObjectUuid = data.CompanyUuid

	// Добавление пользователю прав для данного проекта
	_, err = r.enforcer.AddPolicies([][]string{
		{strconv.Itoa(manager.Id), strconv.Itoa(domainId), projectUuid.String(), actionConstant.DELETE},
		{strconv.Itoa(manager.Id), strconv.Itoa(domainId), projectUuid.String(), actionConstant.MODIFY},
		{strconv.Itoa(manager.Id), strconv.Itoa(domainId), projectUuid.String(), actionConstant.READ},
	})
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectCreateModel{}, err
	}

	// Добавление пользователя в группу менеджеров в рамку данной компании
	_, err = r.enforcer.AddRoleForUserInDomain(strconv.Itoa(manager.Id), gpsm.ToString(), strconv.Itoa(domainId))
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectCreateModel{}, err
	}

	// Save results all operation into a tables
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return projectModel.ProjectCreateModel{}, err
	}

	return projectModel.ProjectCreateModel{
		Logo:        data.Logo,
		Title:       data.Title,
		Description: data.Description,
		Manager:     data.Manager,
		CompanyUuid: projectUuid.String(),
	}, nil
}

/* Add logo project */
func (r *ProjectPostgres) ProjectUpdateImage(userId, domainId int, data projectModel.ProjectImgModel) (projectModel.ProjectImgModel, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return projectModel.ProjectImgModel{}, err
	}

	userIdStr := strconv.Itoa(userId)
	domainIdStr := strconv.Itoa(domainId)

	// Check access for user
	access, err := r.enforcer.Enforce(userIdStr, domainIdStr, data.Uuid, actionConstant.MODIFY)
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectImgModel{}, err
	}

	if !access {
		tx.Rollback()
		return projectModel.ProjectImgModel{}, errors.New("Ошибка! Нет доступа!")
	}

	// Update logo for project
	query := fmt.Sprintf(`UPDATE %s tl SET data = jsonb_set(data, '{logo}', to_jsonb($1::text), true) WHERE tl.uuid = $2`, tableConstant.CB_PROJECTS)

	_, err = r.db.Exec(query, data.Filepath, data.Uuid)
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectImgModel{}, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectImgModel{}, err
	}

	return data, nil
}

/* Get information about one project */
func (r *ProjectPostgres) GetProject(userId, domainId int, data projectModel.ProjectUuidModel) (projectModel.ProjectLowInfoModel, error) {
	var project projectModel.ProjectLowInfoModel

	query := fmt.Sprintf("SELECT uuid, data, created_at FROM %s tl WHERE tl.uuid = $1 LIMIT 1", tableConstant.CB_PROJECTS)

	err := r.db.Get(&project, query, data.Uuid)
	if err != nil {
		return projectModel.ProjectLowInfoModel{}, err
	}

	return project, nil
}

/* Получение информации обо всех проектах */
func (r *ProjectPostgres) GetProjects(userId, domainId int, data projectModel.ProjectCountModel) (projectModel.ProjectAnyCountModel, error) {
	query := fmt.Sprintf("SELECT id FROM %s WHERE uuid=$1", tableConstant.CB_COMPANIES)
	var companyId int

	row := r.db.QueryRow(query, data.Uuid)
	if err := row.Scan(&companyId); err != nil {
		return projectModel.ProjectAnyCountModel{}, err
	}

	var projects []projectModel.ProjectLowInfoModel
	sum := (data.Count + data.Limit)

	query = fmt.Sprintf("SELECT uuid, data, created_at FROM %s tl WHERE tl.companies_id = $1", tableConstant.CB_PROJECTS)
	err := r.db.Select(&projects, query, companyId)
	if err != nil {
		return projectModel.ProjectAnyCountModel{}, err
	}

	var projectsEx []projectModel.ProjectDbDataEx
	for _, element := range projects {
		var projectData projectModel.ProjectDataModel
		err = json.Unmarshal([]byte(element.Data), &projectData)

		if err != nil {
			return projectModel.ProjectAnyCountModel{}, err
		}

		projectsEx = append(projectsEx, projectModel.ProjectDbDataEx{
			Uuid:      element.Uuid,
			Data:      projectData,
			CreatedAt: element.CreatedAt,
		})
	}

	sort.SliceStable(projectsEx, func(i, j int) bool {
		return projectsEx[i].CreatedAt.After(projectsEx[j].CreatedAt)
	})

	if data.Count >= len(projectsEx) {
		return projectModel.ProjectAnyCountModel{}, nil
	}

	if sum >= len(projectsEx) {
		sum -= (sum - len(projectsEx))
	}

	return projectModel.ProjectAnyCountModel{
		Projects: projectsEx[data.Count:sum],
		Count:    (sum - data.Count),
	}, nil
}

/* Get information about any count managers for current project */
func (r *ProjectPostgres) GetProjectManagers(userId, domainId int, data projectModel.ProjectCountModel) (projectModel.ProjectAnyCountModel, error) {
	query := fmt.Sprintf("SELECT id FROM %s WHERE uuid=$1", tableConstant.CB_COMPANIES)
	var companyId int

	row := r.db.QueryRow(query, data.Uuid)
	if err := row.Scan(&companyId); err != nil {
		return projectModel.ProjectAnyCountModel{}, err
	}

	var projects []projectModel.ProjectLowInfoModel
	sum := (data.Count + data.Limit)

	query = fmt.Sprintf("SELECT uuid, data, created_at FROM %s tl WHERE tl.companies_id = $1 LIMIT $2", tableConstant.CB_PROJECTS)
	err := r.db.Select(&projects, query, companyId, sum)
	if err != nil {
		return projectModel.ProjectAnyCountModel{}, err
	}

	var projectsEx []projectModel.ProjectDbDataEx
	for _, element := range projects {
		var projectData projectModel.ProjectDataModel
		err = json.Unmarshal([]byte(element.Data), &projectData)

		if err != nil {
			return projectModel.ProjectAnyCountModel{}, err
		}

		projectsEx = append(projectsEx, projectModel.ProjectDbDataEx{
			Uuid:      element.Uuid,
			Data:      projectData,
			CreatedAt: element.CreatedAt,
		})
	}

	sort.SliceStable(projectsEx, func(i, j int) bool {
		return projectsEx[i].CreatedAt.After(projectsEx[j].CreatedAt)
	})

	if data.Count >= len(projectsEx) {
		return projectModel.ProjectAnyCountModel{}, nil
	}

	if sum >= len(projectsEx) {
		sum -= (sum - len(projectsEx))
	}

	return projectModel.ProjectAnyCountModel{
		Projects: projectsEx[data.Count:sum],
		Count:    (sum - data.Count),
	}, nil
}

func (r *ProjectPostgres) ProjectUpdate(user userModel.UserIdentityModel, data projectModel.ProjectUpdateModel) (projectModel.ProjectUpdateModel, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return projectModel.ProjectUpdateModel{}, err
	}

	userIdStr := strconv.Itoa(user.UserId)
	domainIdStr := strconv.Itoa(user.DomainId)

	// Check access for user
	access, err := r.enforcer.Enforce(userIdStr, domainIdStr, data.Uuid, actionConstant.MODIFY)
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectUpdateModel{}, err
	}

	if !access {
		tx.Rollback()
		return projectModel.ProjectUpdateModel{}, errors.New("Ошибка! Нет доступа!")
	}

	var projectInfo []projectModel.ProjectLowInfoModel
	query := fmt.Sprintf("SELECT uuid, data, created_at FROM %s WHERE uuid=$1 LIMIT 1", tableConstant.CB_PROJECTS)

	err = r.db.Select(&projectInfo, query, data.Uuid)
	if err != nil {
		return projectModel.ProjectUpdateModel{}, err
	}

	var projectData projectModel.ProjectDataModel
	err = json.Unmarshal([]byte(projectInfo[0].Data), &projectData)

	projectData.Description = data.Description
	projectData.Title = data.Title
	// projectData.Managers = data.Managers (need make)

	projectDataJson, err := json.Marshal(projectData)
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectUpdateModel{}, err
	}

	query = fmt.Sprintf("UPDATE %s tl SET data=$1 WHERE tl.uuid=$2", tableConstant.CB_PROJECTS)

	_, err = tx.Exec(query, projectDataJson, data.Uuid)
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectUpdateModel{}, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectUpdateModel{}, err
	}

	return data, nil
}
