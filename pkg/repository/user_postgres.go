package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	actionConstant "main-server/pkg/constant/action"
	middlewareConstant "main-server/pkg/constant/middleware"
	objectConstant "main-server/pkg/constant/object"
	tableConstant "main-server/pkg/constant/table"
	companyModel "main-server/pkg/model/company"
	rbacModel "main-server/pkg/model/rbac"
	userModel "main-server/pkg/model/user"
	"strconv"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type UserPostgres struct {
	db       *sqlx.DB
	enforcer *casbin.Enforcer
	domain   *DomainPostgres
	role     *RolePostgres
}

/*
* Функция создания экземпляра сервиса
 */
func NewUserPostgres(
	db *sqlx.DB, enforcer *casbin.Enforcer,
	domain *DomainPostgres, role *RolePostgres,
) *UserPostgres {
	return &UserPostgres{
		db:       db,
		enforcer: enforcer,
		domain:   domain,
		role:     role,
	}
}

func (r *UserPostgres) Get(column string, value interface{}, check bool) (*userModel.UserModel, error) {
	var users []userModel.UserModel
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1", tableConstant.U_USERS, column)

	var err error

	switch value.(type) {
	case int:
		err = r.db.Select(&users, query, value.(int))
		break
	case string:
		err = r.db.Select(&users, query, value.(string))
		break
	}

	if len(users) <= 0 {
		if check {
			return nil, errors.New(fmt.Sprintf("Ошибка: пользователя по запросу %s:%s не найдено!", column, value))
		}

		return nil, nil
	}

	return &users[len(users)-1], err
}

func (r *UserPostgres) GetProfile(c *gin.Context) (userModel.UserProfileModel, error) {
	usersId, _ := c.Get(middlewareConstant.USER_CTX)

	var profile userModel.UserProfileModel
	var email userModel.UserEmailModel

	query := fmt.Sprintf("SELECT data FROM %s tl WHERE tl.users_id = $1 LIMIT 1",
		tableConstant.U_USERS_DATA,
	)

	err := r.db.Get(&profile, query, usersId)
	if err != nil {
		return userModel.UserProfileModel{}, err
	}

	query = fmt.Sprintf("SELECT email FROM %s tl WHERE tl.id = $1 LIMIT 1", tableConstant.U_USERS)

	err = r.db.Get(&email, query, usersId)
	if err != nil {
		return userModel.UserProfileModel{}, err
	}

	return userModel.UserProfileModel{
		Email: email.Email,
		Data:  profile.Data,
	}, nil
}

func (r *UserPostgres) UpdateProfile(c *gin.Context, data userModel.UserProfileUpdateDataModel) (userModel.UserJSONBModel, error) {
	usersId, _ := c.Get(middlewareConstant.USER_CTX)

	userJsonb, err := json.Marshal(data)
	if err != nil {
		return userModel.UserJSONBModel{}, err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return userModel.UserJSONBModel{}, err
	}

	query := fmt.Sprintf("UPDATE %s tl SET data=$1 WHERE tl.users_id = $2", tableConstant.U_USERS_DATA)

	// Update data about user profile
	_, err = tx.Exec(query, userJsonb, usersId)
	if err != nil {
		tx.Rollback()
		return userModel.UserJSONBModel{}, err
	}

	query = fmt.Sprintf("SELECT data FROM %s tl WHERE users_id=$1 LIMIT 1", tableConstant.U_USERS_DATA)
	var userData []userModel.UserDataModel

	err = r.db.Select(&userData, query, usersId)

	if err != nil {
		tx.Rollback()
		return userModel.UserJSONBModel{}, err
	}

	if len(userData) <= 0 {
		tx.Rollback()
		return userModel.UserJSONBModel{}, errors.New("Данных у пользователя нет")
	}

	var dataFromJson userModel.UserJSONBModel
	err = json.Unmarshal([]byte(userData[0].Data), &dataFromJson)

	if err != nil {
		tx.Rollback()
		return userModel.UserJSONBModel{}, err
	}

	// Change password
	if data.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*data.Password), viper.GetInt("crypt.cost"))
		if err != nil {
			tx.Rollback()
			return userModel.UserJSONBModel{}, err
		}

		query := fmt.Sprintf("UPDATE %s SET password=$1 WHERE id=$2", tableConstant.U_USERS)
		_, err = r.db.Exec(query, string(hashedPassword), usersId)

		if err != nil {
			tx.Rollback()
			return userModel.UserJSONBModel{}, err
		}
	}

	err = tx.Commit()

	if err != nil {
		tx.Rollback()
		return userModel.UserJSONBModel{}, err
	}

	return dataFromJson, nil
}

/* Get information about company for current user */
func (r *UserPostgres) GetUserCompany(userId, domainId int) (companyModel.CompanyInfoModelEx, error) {
	query := fmt.Sprintf(`
			SELECT DISTINCT tl.value, trule.v3
			FROM %s tl
			JOIN %s tr ON tr.id = tl.types_objects_id
			JOIN %s trule ON trule.v2 = tl.value 
			WHERE tr.value = $1 
					AND trule.v0 = $2 
					AND trule.ptype = 'p'
					AND (trule.v3 = $3 OR trule.v3 = $4)
	`, tableConstant.AC_OBJECTS, tableConstant.AC_TYPES_OBJECTS, tableConstant.AC_RULES)

	var companies []companyModel.CompanyRuleModelEx
	err := r.db.Select(&companies, query,
		objectConstant.COMPANY, userId,
		actionConstant.ADMINISTRATION, actionConstant.MANAGEMENT,
	)

	if err != nil {
		return companyModel.CompanyInfoModelEx{}, err
	}

	if len(companies) <= 0 {
		return companyModel.CompanyInfoModelEx{}, nil
	}

	company := companies[len(companies)-1]
	policies := r.enforcer.GetFilteredPolicy(0, strconv.Itoa(userId), strconv.Itoa(domainId), company.Value)

	rules := lo.Map(policies, func(x []string, _ int) string {
		return x[len(x)-1]
	})

	var companyInfo companyModel.CompanyInfoModel
	query = fmt.Sprintf(`SELECT uuid, data, created_at FROM %s tl WHERE tl.uuid = $1 LIMIT 1`, tableConstant.CB_COMPANIES)

	err = r.db.Get(&companyInfo, query, company.Value)
	if err != nil {
		return companyModel.CompanyInfoModelEx{}, err
	}

	var companyDataEx companyModel.CompanyModel
	err = json.Unmarshal([]byte(companyInfo.Data), &companyDataEx)
	if err != nil {
		return companyModel.CompanyInfoModelEx{}, err
	}

	return companyModel.CompanyInfoModelEx{
		Uuid:      companyInfo.Uuid,
		Data:      companyDataEx,
		CreatedAt: companyInfo.CreatedAt,
		Rules:     rules,
	}, nil
}

/* Проверка доступа пользователя */
func (r *UserPostgres) AccessCheck(userId, domainId int, value rbacModel.RoleValueModel) (bool, error) {
	role, err := r.role.Get("value", value.Value, true)
	if err != nil {
		return false, err
	}

	result, err := r.enforcer.HasRoleForUser(strconv.Itoa(userId), strconv.Itoa(role.Id), strconv.Itoa(domainId))
	if err != nil {
		return false, err
	}

	return result, nil
}

/* Метод для получения информации о всех ролях пользователя (функциональные модули пользователя) */
func (r *UserPostgres) GetAllRoles(user userModel.UserIdentityModel) (*userModel.UserRoleModel, error) {
	query := fmt.Sprintf(`
			SELECT DISTINCT tl.value 
			FROM %s tr
			JOIN %s tl ON (tl.id = tr.v1::integer AND tr.ptype = 'g')
			WHERE
				tr.v2 = $1 AND 
				tr.v0 = $2
	`, tableConstant.AC_RULES, tableConstant.AC_ROLES)

	var roles []string

	err := r.db.Select(&roles, query, user.DomainId, user.UserId)
	if err != nil {
		return nil, err
	}

	return &userModel.UserRoleModel{
		Roles: roles,
	}, nil
}
