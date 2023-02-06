package repository

import (
	adminModel "main-server/pkg/model/admin"
	companyModel "main-server/pkg/model/company"
	emailModel "main-server/pkg/model/email"
	excelModel "main-server/pkg/model/excel"
	projectModel "main-server/pkg/model/project"
	rbacModel "main-server/pkg/model/rbac"
	userModel "main-server/pkg/model/user"
	infoModel "main-server/pkg/module/excel_analysis/model"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
)

type Authorization interface {
	CreateUser(user userModel.UserRegisterModel) (userModel.UserAuthDataModel, error)
	UploadProfileImage(c *gin.Context, filepath string) (bool, error)
	LoginUser(user userModel.UserLoginModel) (userModel.UserAuthDataModel, error)
	LoginUserOAuth2(code string) (userModel.UserAuthDataModel, error)
	CreateUserOAuth2(user userModel.UserRegisterOAuth2Model, token *oauth2.Token) (userModel.UserAuthDataModel, error)
	Refresh(data userModel.TokenLogoutDataModel, refreshToken string, token userModel.TokenOutputParse) (userModel.UserAuthDataModel, error)
	Logout(tokens userModel.TokenLogoutDataModel) (bool, error)
	Activate(link string) (bool, error)
	GetUser(column, value string) (userModel.UserModel, error)
	GetRole(column, value string) (rbacModel.RoleModel, error)
	RecoveryPassword(email string) (bool, error)
	ResetPassword(data userModel.ResetPasswordModel, token userModel.ResetTokenOutputParse) (bool, error)
}

type Role interface {
	GetRole(column, value interface{}) (rbacModel.RoleModel, error)
	HasRole(usersId, domainsId int, roleValue string) (bool, error)
}

type Domain interface {
	GetDomain(column, value interface{}) (rbacModel.DomainModel, error)
}

type User interface {
	GetUser(column, value interface{}) (userModel.UserModel, error)
	GetProfile(c *gin.Context) (userModel.UserProfileModel, error)
	UpdateProfile(c *gin.Context, data userModel.UserProfileUpdateDataModel) (userModel.UserJSONBModel, error)
	GetUserCompany(userId, domainId int) (companyModel.CompanyInfoModelEx, error)
	AccessCheck(userId, domainId int, value rbacModel.RoleValueModel) (bool, error)
	GetAllRoles(user userModel.UserIdentityModel) (*userModel.UserRoleModel, error)
}

type Admin interface {
	GetAllUsers(c *gin.Context) (adminModel.UsersResponseModel, error)
	CreateCompany(c *gin.Context, data adminModel.CompanyModel) (adminModel.CompanyModel, error)
	SystemAddManager(user *userModel.UserIdentityModel, data adminModel.SystemPermissionModel) (bool, error)
}

type AuthType interface {
	GetAuthType(column, value interface{}) (userModel.AuthTypeModel, error)
}

type Project interface {
	CreateProject(userId, domainId int, data projectModel.ProjectCreateModel) (projectModel.ProjectCreateModel, error)
	ProjectUpdate(user userModel.UserIdentityModel, data projectModel.ProjectUpdateModel) (projectModel.ProjectUpdateModel, error)
	ProjectUpdateImage(userId, domainId int, data projectModel.ProjectImgModel) (projectModel.ProjectImgModel, error)
	GetProject(userId, domainId int, data projectModel.ProjectUuidModel) (projectModel.ProjectLowInfoModel, error)
	GetProjects(userId, domainId int, data projectModel.ProjectCountModel) (projectModel.ProjectAnyCountModel, error)
}

type Company interface {
	GetManagers(userId, domainId int, data companyModel.ManagerCountModel) (companyModel.ManagerAnyCountModel, error)
	GetManager(user userModel.UserIdentityModel, data companyModel.ManagerUuidModel) (companyModel.ManagerCompanyModel, error)
	CompanyUpdateImage(user userModel.UserIdentityModel, data companyModel.CompanyImageModel) (companyModel.CompanyImageModel, error)
	CompanyUpdate(user userModel.UserIdentityModel, data companyModel.CompanyUpdateModel) (companyModel.CompanyUpdateModel, error)
}

type Wrapper interface {
	GetOne(table string, column, value interface{}) (interface{}, error)
}

type ServiceMain interface {
	SendEmail(*userModel.UserIdentityModel, *emailModel.MessageInputModel) (bool, error)
}

type ExcelAnalysis interface {
	GetHeaderInfoDocument(document excelModel.DocumentIdModel) (infoModel.HeaderInfoModel, error)
}

/* Интерфейс репозитория для таблицы ac_objects */
type Object interface {
	GetObject(column, value interface{}) (*rbacModel.ObjectDbModel, error)
	AddResource(resource *rbacModel.ResourceModel) (*rbacModel.ObjectDbModel, error)
}

/* Интерфейс репозитория для таблицы ac_types_objects */
type TypeObject interface {
	GetTypeObject(column, value interface{}) (rbacModel.TypeObjectDbModel, error)
}

type Repository struct {
	Authorization
	Role
	Domain
	User
	Admin
	AuthType
	Project
	Company
	Wrapper
	ServiceMain
	ExcelAnalysis
	Object
	TypeObject
}

func NewRepository(db *sqlx.DB, enforcer *casbin.Enforcer) *Repository {
	wrapper := NewWrapperPostgres(db)

	domain := NewDomainPostgres(db)
	role := NewRolePostgres(db, enforcer)
	user := NewUserPostgres(db, enforcer, domain, role)
	admin := NewAdminPostgres(db, enforcer, domain, role, user)
	project := NewProjectPostgres(db, enforcer, role, user)
	company := NewCompanyPostgres(db, enforcer, role, user, wrapper)
	serviceMain := NewServiceMainRepository(db, enforcer, user)
	acTypeObject := NewTypeObjectPostgres(db)

	return &Repository{
		Authorization: NewAuthPostgres(db, enforcer, *user),
		Role:          role,
		Domain:        domain,
		User:          user,
		Admin:         admin,
		AuthType:      NewAuthTypePostgres(db),
		Project:       project,
		Company:       company,
		Wrapper:       wrapper,
		ServiceMain:   serviceMain,
		ExcelAnalysis: NewExcelAnalysis(),
		Object:        NewObjectPostgres(db, acTypeObject),
		TypeObject:    acTypeObject,
	}
}
