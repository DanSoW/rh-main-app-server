package repository

import (
	//middlewareConstant "main-server/pkg/constant/middleware"
	"encoding/json"
	"fmt"
	actionConstant "main-server/pkg/constant/action"
	middlewareConstant "main-server/pkg/constant/middleware"
	objectConstant "main-server/pkg/constant/object"
	roleConstant "main-server/pkg/constant/role"
	tableConstant "main-server/pkg/constant/table"
	adminModel "main-server/pkg/model/admin"
	rbacModel "main-server/pkg/model/rbac"
	userModel "main-server/pkg/model/user"
	"main-server/pkg/model/worker"
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
	role     *RolePostgres
}

/* Function for create new struct of AdminPostgres */
func NewAdminPostgres(
	db *sqlx.DB,
	enforcer *casbin.Enforcer,
	domain *DomainPostgres,
	role *RolePostgres,
) *AdminPostgres {
	return &AdminPostgres{
		db:       db,
		enforcer: enforcer,
		domain:   domain,
		role:     role,
	}
}

/* Метод для получения информации обо всех пользователях */
func (r *AdminPostgres) GetAllUsers(c *gin.Context) (adminModel.UsersResponseModel, error) {
	// Получение email-адресов всех пользователей
	query := fmt.Sprintf(`SELECT email FROM %s`, tableConstant.USERS_TABLE)
	var users []adminModel.UserResponseModel

	if err := r.db.Select(&users, query); err != nil {
		return adminModel.UsersResponseModel{}, err
	}

	return adminModel.UsersResponseModel{
		Users: &users,
	}, nil
}

/* Функция создания новой компании (доступно админу) */
func (r *AdminPostgres) CreateCompany(c *gin.Context, data adminModel.CompanyModel) (adminModel.CompanyModel, error) {
	// Получение основных данных из контекста
	usersId, _ := c.Get(middlewareConstant.USER_CTX)
	domainsId, _ := c.Get(middlewareConstant.DOMAINS_ID)

	// Обёртка всех операций в одну транзакцию в рамках текущей функции
	tx, err := r.db.Begin()
	if err != nil {
		return adminModel.CompanyModel{}, err
	}

	query := fmt.Sprintf("INSERT INTO %s (uuid, data, created_at, updated_at, users_id) values ($1, $2, $3, $4, $5) RETURNING id", tableConstant.COMPANIES_TABLE)

	/* Добавление новой компании */
	dataJson, err := json.Marshal(data)
	if err != nil {
		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	var companyId int
	currentDate := time.Now()
	companyUuid := uuid.NewV4()

	row := tx.QueryRow(query, companyUuid, dataJson, currentDate, currentDate, usersId)
	if err := row.Scan(&companyId); err != nil {
		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	if err != nil {
		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	/* Добавление информации о новом объекте (объект в данном случае - это компания) */
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

	/* Добавление информации о пользователе в текущей компании */
	// Осуществление поиска пользователя по его email-адресу
	var user userModel.UserModel
	query = fmt.Sprintf("SELECT * FROM %s tl WHERE tl.email=$1", tableConstant.USERS_TABLE)
	err = r.db.Get(&user, query, data.EmailAdmin)
	if err != nil {
		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	// Получение роли администратора-застройщика, которая будет выдана пользователю
	roleAdmin, err := r.role.GetRole("value", roleConstant.ROLE_BUILDER_ADMIN)
	if err != nil {
		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	roleAdminJson, err := json.Marshal(worker.WorkerModel{
		Role: roleAdmin.Uuid,
	})

	if err != nil {
		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	// Добавление нового участника в компании
	query = fmt.Sprintf(`
		INSERT INTO %s (uuid, data, created_at, updated_at, users_id, companies_id) 
		values ($1, $2, $3, $4, $5, $6) RETURNING id`,
		tableConstant.WORKERS_TABLE,
	)
	var workerId int
	row = tx.QueryRow(query, uuid.NewV4().String(), roleAdminJson, currentDate, currentDate, user.Id, companyId)
	if err := row.Scan(&workerId); err != nil {
		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	var userId string = strconv.Itoa(usersId.(int))
	var domainId string = strconv.Itoa(domainsId.(int))
	var roleAdminId string = strconv.Itoa(roleAdmin.Id)

	// Получение модели GPSubjectModel по строке
	gpsm, err := rbacModel.NewGPSubjectModel(fmt.Sprintf("%s%s%s", roleAdminId, rbacModel.Separator, companyUuid.String()))
	if err != nil {
		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	// Обновление политик управления доступа для текущего пользователя (администратор, менеджер или "выше")
	_, err = r.enforcer.AddPolicies([][]string{
		{userId, domainId, companyUuid.String(), actionConstant.DELETE},
		{userId, domainId, companyUuid.String(), actionConstant.MODIFY},
		{userId, domainId, companyUuid.String(), actionConstant.READ},
	})

	if err != nil {
		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	// Добавление пользователя в группу администраторов в контексте данной компании
	_, err = r.enforcer.AddRoleForUserInDomain(userId, gpsm.ToString(), domainId)

	if err != nil {
		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	// Get id for user admin
	query = fmt.Sprintf("SELECT id FROM %s WHERE email=$1 LIMIT 1", tableConstant.USERS_TABLE)
	var userAdminId int

	row = r.db.QueryRow(query, data.EmailAdmin)
	if err := row.Scan(&userAdminId); err != nil {
		r.enforcer.RemovePolicies([][]string{
			{userId, domainId, companyUuid.String(), actionConstant.DELETE},
			{userId, domainId, companyUuid.String(), actionConstant.MODIFY},
			{userId, domainId, companyUuid.String(), actionConstant.READ},
		})

		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	userId = strconv.Itoa(userAdminId)

	_, err = r.enforcer.AddPolicies([][]string{
		{userId, domainId, companyUuid.String(), actionConstant.MODIFY},
		{userId, domainId, companyUuid.String(), actionConstant.READ},
		{userId, domainId, companyUuid.String(), actionConstant.DELETE},
	})

	// Save results all operation into a tables
	if err := tx.Commit(); err != nil {
		r.enforcer.RemovePolicies([][]string{
			{userId, domainId, companyUuid.String(), actionConstant.MODIFY},
			{userId, domainId, companyUuid.String(), actionConstant.READ},
		})

		tx.Rollback()
		return adminModel.CompanyModel{}, err
	}

	return data, nil
}
