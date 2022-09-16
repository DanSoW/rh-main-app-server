package repository

import (
	//middlewareConstant "main-server/pkg/constant/middleware"
	"encoding/json"
	"fmt"
	actionConstant "main-server/pkg/constant/action"
	middlewareConstant "main-server/pkg/constant/middleware"
	objectConstant "main-server/pkg/constant/object"
	tableConstant "main-server/pkg/constant/table"
	adminModel "main-server/pkg/model/admin"
	rbacModel "main-server/pkg/model/rbac"
	"strconv"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
)

type AdminPostgres struct {
	db       *sqlx.DB
	enforcer *casbin.Enforcer
	domain   *DomainPostgres
}

/* Function for create new struct of AdminPostgres */
func NewAdminPostgres(db *sqlx.DB, enforcer *casbin.Enforcer, domain *DomainPostgres) *AdminPostgres {
	return &AdminPostgres{
		db:       db,
		enforcer: enforcer,
		domain:   domain,
	}
}

/* Method for get all users, when location in system */
func (r *AdminPostgres) GetAllUsers(c *gin.Context) (adminModel.UsersResponseModel, error) {
	/*usersId, _ := c.Get(middlewareConstant.USER_CTX)
	domainsId, _ := c.Get(middlewareConstant.DOMAINS_ID)*/

	// Select all users without ban
	query := fmt.Sprintf(`SELECT email FROM %s`, tableConstant.USERS_TABLE)
	var users []adminModel.UserResponseModel

	if err := r.db.Select(&users, query); err != nil {
		return adminModel.UsersResponseModel{}, err
	}

	return adminModel.UsersResponseModel{
		Users: &users,
	}, nil
}

func (r *AdminPostgres) CreateCompany(c *gin.Context, data adminModel.CompanyModel) (adminModel.CompanyModel, error) {
	usersId, _ := c.Get(middlewareConstant.USER_CTX)
	domainsId, _ := c.Get(middlewareConstant.DOMAINS_ID)

	// Wrapping all operations into a transaction
	tx, err := r.db.Begin()
	if err != nil {
		return adminModel.CompanyModel{}, err
	}

	query := fmt.Sprintf("INSERT INTO %s (uuid, data, created_at, updated_at, users_id) values ($1, $2, $3, $4, $5)", tableConstant.COMPANIES_TABLE)

	dataJson, err := json.Marshal(data)
	if err != nil {
		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	currentDate := time.Now()
	companyUuid := uuid.NewV4()

	_, err = tx.Exec(query, companyUuid, dataJson, currentDate, currentDate, usersId)

	if err != nil {
		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	// Adding information about a resource
	var typesObjects rbacModel.TypesObjectsModel

	query = fmt.Sprintf("SELECT * FROM %s WHERE value=$1", tableConstant.TYPES_OBJECTS_TABLE)

	err = r.db.Get(&typesObjects, query, objectConstant.TYPE_COMPANY)
	if err != nil {
		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	query = fmt.Sprintf("INSERT INTO %s (value, types_objects_id) values ($1, $2)", tableConstant.OBJECTS_TABLE)

	_, err = tx.Exec(query, companyUuid, typesObjects.Id)
	if err != nil {
		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	var userId string = strconv.Itoa(usersId.(int))
	var domainId string = strconv.Itoa(domainsId.(int))

	// Update current user policy for current article

	// For creator company
	_, err = r.enforcer.AddPolicies([][]string{
		{userId, domainId, companyUuid.String(), actionConstant.DELETE},
		{userId, domainId, companyUuid.String(), actionConstant.MODIFY},
		{userId, domainId, companyUuid.String(), actionConstant.READ},
		{userId, domainId, companyUuid.String(), actionConstant.ADMINISTRATION},
	})

	if err != nil {
		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	// Get id for user admin
	query = fmt.Sprintf("SELECT id FROM %s WHERE email=$1 LIMIT 1", tableConstant.USERS_TABLE)
	var userAdminId int

	row := r.db.QueryRow(query, data.EmailAdmin)
	if err := row.Scan(&userAdminId); err != nil {
		r.enforcer.RemovePolicies([][]string{
			{userId, domainId, companyUuid.String(), actionConstant.DELETE},
			{userId, domainId, companyUuid.String(), actionConstant.MODIFY},
			{userId, domainId, companyUuid.String(), actionConstant.READ},
			{userId, domainId, companyUuid.String(), actionConstant.ADMINISTRATION},
		})

		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	userId = strconv.Itoa(userAdminId)

	// For admin company
	_, err = r.enforcer.AddPolicies([][]string{
		{userId, domainId, companyUuid.String(), actionConstant.MODIFY},
		{userId, domainId, companyUuid.String(), actionConstant.READ},
		{userId, domainId, companyUuid.String(), actionConstant.ADMINISTRATION},
	})

	// Save results all operation into a tables
	if err := tx.Commit(); err != nil {
		r.enforcer.RemovePolicies([][]string{
			{userId, domainId, companyUuid.String(), actionConstant.MODIFY},
			{userId, domainId, companyUuid.String(), actionConstant.READ},
			{userId, domainId, companyUuid.String(), actionConstant.ADMINISTRATION},
		})

		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	return data, nil
}
