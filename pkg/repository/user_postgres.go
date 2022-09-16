package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	middlewareConstants "main-server/pkg/constant/middleware"
	tableConstants "main-server/pkg/constant/table"
	userModel "main-server/pkg/model/user"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type UserPostgres struct {
	db       *sqlx.DB
	enforcer *casbin.Enforcer
	domain   *DomainPostgres
}

/*
* Функция создания экземпляра сервиса
 */
func NewUserPostgres(db *sqlx.DB, enforcer *casbin.Enforcer, domain *DomainPostgres) *UserPostgres {
	return &UserPostgres{
		db:       db,
		enforcer: enforcer,
		domain:   domain,
	}
}

func (r *UserPostgres) GetUser(column, value interface{}) (userModel.UserModel, error) {
	var user userModel.UserModel
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1", tableConstants.USERS_TABLE, column.(string))

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

func (r *UserPostgres) GetProfile(c *gin.Context) (userModel.UserProfileModel, error) {
	usersId, _ := c.Get(middlewareConstants.USER_CTX)

	var profile userModel.UserProfileModel
	var email userModel.UserEmailModel

	query := fmt.Sprintf("SELECT data FROM %s tl WHERE tl.users_id = $1 LIMIT 1",
		tableConstants.USERS_DATA_TABLE,
	)

	err := r.db.Get(&profile, query, usersId)
	if err != nil {
		return userModel.UserProfileModel{}, err
	}

	query = fmt.Sprintf("SELECT email FROM %s tl WHERE tl.id = $1 LIMIT 1", tableConstants.USERS_TABLE)

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
	usersId, _ := c.Get(middlewareConstants.USER_CTX)

	userJsonb, err := json.Marshal(data)
	if err != nil {
		return userModel.UserJSONBModel{}, err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return userModel.UserJSONBModel{}, err
	}

	query := fmt.Sprintf("UPDATE %s tl SET data=$1 WHERE tl.users_id = $2", tableConstants.USERS_DATA_TABLE)

	// Update data about user profile
	_, err = tx.Exec(query, userJsonb, usersId)
	if err != nil {
		tx.Rollback()
		return userModel.UserJSONBModel{}, err
	}

	query = fmt.Sprintf("SELECT data FROM %s tl WHERE users_id=$1 LIMIT 1", tableConstants.USERS_DATA_TABLE)
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

		query := fmt.Sprintf("UPDATE %s SET password=$1 WHERE id=$2", tableConstants.USERS_TABLE)
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
