package repository

import (
	//middlewareConstant "main-server/pkg/constant/middleware"
	"encoding/json"
	"errors"
	"fmt"
	actionConstant "main-server/pkg/constant/action"
	objectConstant "main-server/pkg/constant/object"
	tableConstant "main-server/pkg/constant/table"
	projectModel "main-server/pkg/model/project"
	rbacModel "main-server/pkg/model/rbac"
	"strconv"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
)

type ProjectPostgres struct {
	db       *sqlx.DB
	enforcer *casbin.Enforcer
}

/* Function for create new struct of AdminPostgres */
func NewProjectPostgres(db *sqlx.DB, enforcer *casbin.Enforcer) *ProjectPostgres {
	return &ProjectPostgres{
		db:       db,
		enforcer: enforcer,
	}
}

func (r *ProjectPostgres) CreateProject(userId, domainId int, data projectModel.ProjectModel) (projectModel.ProjectModel, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return projectModel.ProjectModel{}, err
	}

	query := fmt.Sprintf("SELECT id FROM %s WHERE uuid=$1", tableConstant.COMPANIES_TABLE)
	var companyId int

	row := r.db.QueryRow(query, data.Uuid)
	if err := row.Scan(&companyId); err != nil {
		tx.Rollback()
		return projectModel.ProjectModel{}, err
	}

	query = fmt.Sprintf("INSERT INTO %s (uuid, data, created_at, updated_at, users_id, companies_id) values ($1, $2, $3, $4, $5, $6) RETURNING uuid", tableConstant.PROJECTS_TABLE)

	dataJson, err := json.Marshal(projectModel.ProjectDbModel{
		Title:       data.Title,
		Description: data.Description,
		Managers:    data.Managers,
	})

	if err != nil {
		tx.Rollback()
		return projectModel.ProjectModel{}, err
	}

	currentDate := time.Now()
	projectUuid := uuid.NewV4()

	var uuid string

	row = tx.QueryRow(query, projectUuid, dataJson, currentDate, currentDate, userId, companyId)
	if err := row.Scan(&uuid); err != nil {
		tx.Rollback()
		return projectModel.ProjectModel{}, err
	}

	if err != nil {
		tx.Rollback()
		return projectModel.ProjectModel{}, err
	}

	// Adding information about a resource
	var typesObjects rbacModel.TypesObjectsModel

	query = fmt.Sprintf("SELECT * FROM %s WHERE value=$1", tableConstant.TYPES_OBJECTS_TABLE)

	err = r.db.Get(&typesObjects, query, objectConstant.TYPE_PROJECT)
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectModel{}, err
	}

	query = fmt.Sprintf("INSERT INTO %s (value, types_objects_id) values ($1, $2)", tableConstant.OBJECTS_TABLE)

	_, err = tx.Exec(query, data.Uuid, typesObjects.Id)
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectModel{}, err
	}

	var userIdStr string = strconv.Itoa(userId)
	var domainIdStr string = strconv.Itoa(domainId)

	// Update current user policy for current article
	_, err = r.enforcer.AddPolicies([][]string{
		{userIdStr, domainIdStr, uuid, actionConstant.DELETE},
		{userIdStr, domainIdStr, uuid, actionConstant.MODIFY},
		{userIdStr, domainIdStr, uuid, actionConstant.READ},
		{userIdStr, domainIdStr, uuid, actionConstant.ADMINISTRATION},
	})

	if err != nil {
		tx.Rollback()
		return projectModel.ProjectModel{}, err
	}

	// Get ids for user managers
	query = fmt.Sprintf("SELECT id FROM %s WHERE email=$1 LIMIT 1", tableConstant.USERS_TABLE)

	for _, element := range data.Managers {
		// @(idea):
		// Можно улучшить данный фрагмент кода добавив сюда определение функции транзакции,
		// которая бы выполнялась в текущей транзакции

		var userManagerId int
		row := r.db.QueryRow(query, element.Email)
		if err := row.Scan(&userManagerId); err != nil {
			continue
		}

		userManagerIdStr := strconv.Itoa(userManagerId)

		_, err = r.enforcer.AddPolicies([][]string{
			{userManagerIdStr, domainIdStr, uuid, actionConstant.READ},
			{userManagerIdStr, domainIdStr, uuid, actionConstant.MANAGEMENT},
		})
	}

	// Save results all operation into a tables
	if err := tx.Commit(); err != nil {
		// Rejected
		r.enforcer.RemovePolicies([][]string{
			{userIdStr, domainIdStr, uuid, actionConstant.DELETE},
			{userIdStr, domainIdStr, uuid, actionConstant.MODIFY},
			{userIdStr, domainIdStr, uuid, actionConstant.READ},
			{userIdStr, domainIdStr, uuid, actionConstant.ADMINISTRATION},
		})

		for _, element := range data.Managers {
			var userManagerId int
			row := r.db.QueryRow(query, element.Email)
			if err := row.Scan(&userManagerId); err != nil {
				continue
			}

			userManagerIdStr := strconv.Itoa(userManagerId)

			_, err = r.enforcer.RemovePolicies([][]string{
				{userManagerIdStr, domainIdStr, uuid, actionConstant.READ},
				{userManagerIdStr, domainIdStr, uuid, actionConstant.MANAGEMENT},
			})
		}

		tx.Rollback()
		return projectModel.ProjectModel{}, err
	}

	return projectModel.ProjectModel{
		Logo:        data.Logo,
		Title:       data.Title,
		Description: data.Description,
		Managers:    data.Managers,
		Uuid:        uuid,
	}, nil
}

/* Add logo project */
func (r *ProjectPostgres) AddLogoProject(userId, domainId int, data projectModel.ProjectLogoModel) (projectModel.ProjectLogoModel, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return projectModel.ProjectLogoModel{}, err
	}

	userIdStr := strconv.Itoa(userId)
	domainIdStr := strconv.Itoa(domainId)

	// Check access for user
	access, err := r.enforcer.Enforce(userIdStr, domainIdStr, data.Uuid, actionConstant.MODIFY)
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectLogoModel{}, err
	}

	if !access {
		tx.Rollback()
		return projectModel.ProjectLogoModel{}, errors.New("Ошибка! Нет доступа!")
	}

	// Update logo for project
	query := fmt.Sprintf(`UPDATE %s tl SET data = jsonb_set(data, '{logo}', to_jsonb($1::text), true) WHERE tl.uuid = $2`, tableConstant.PROJECTS_TABLE)

	_, err = r.db.Exec(query, data.Filepath, data.Uuid)
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectLogoModel{}, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return projectModel.ProjectLogoModel{}, err
	}

	return data, nil
}
