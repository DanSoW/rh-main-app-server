package service

import (
	adminModel "main-server/pkg/model/admin"
	companyModel "main-server/pkg/model/company"
	emailModel "main-server/pkg/model/email"
	excelModel "main-server/pkg/model/excel"
	projectModel "main-server/pkg/model/project"
	rbacModel "main-server/pkg/model/rbac"
	userModel "main-server/pkg/model/user"
	infoModel "main-server/pkg/module/excel_analysis/model"
	repository "main-server/pkg/repository"

	"github.com/gin-gonic/gin"
)

type Authorization interface {
	CreateUser(user userModel.UserRegisterModel) (userModel.UserAuthDataModel, error)
	UploadProfileImage(c *gin.Context, filepath string) (bool, error)
	LoginUser(user userModel.UserLoginModel) (userModel.UserAuthDataModel, error)
	LoginUserOAuth2(code string) (userModel.UserAuthDataModel, error)
	Refresh(data userModel.TokenLogoutDataModel, refreshToken string) (userModel.UserAuthDataModel, error)
	Logout(tokens userModel.TokenLogoutDataModel) (bool, error)
	Activate(link string) (bool, error)
	RecoveryPassword(email string) (bool, error)
	ResetPassword(data userModel.ResetPasswordModel) (bool, error)
}

type Token interface {
	ParseToken(token, signingKey string) (userModel.TokenOutputParse, error)
	ParseTokenWithoutValid(token, signingKey string) (userModel.TokenOutputParse, error)
	ParseResetToken(pToken, signingKey string) (userModel.ResetTokenOutputParse, error)
}

type AuthType interface {

	// CRUD
	Get(column string, value interface{}, check bool) (*userModel.AuthTypeModel, error)
}

type User interface {
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

type Domain interface {

	// CRUD
	Get(column string, value interface{}, check bool) (*rbacModel.DomainModel, error)
}

type Role interface {
	HasRole(usersId, domainsId int, roleValue string) (bool, error)

	// CRUD
	Get(column string, value interface{}, check bool) (*rbacModel.RoleModel, error)
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

type ServiceMain interface {
	SendEmail(user *userModel.UserIdentityModel, body *emailModel.MessageInputModel) (bool, error)
}

type ExcelAnalysis interface {
	GetHeaderInfoDocument(document excelModel.DocumentIdModel) (infoModel.HeaderInfoModel, error)
}

type Object interface {
	AddResource(resource *rbacModel.ResourceModel) (*rbacModel.ObjectDbModel, error)

	// CRUD
	Get(column string, value interface{}, check bool) (*rbacModel.ObjectDbModel, error)
}

type Service struct {
	Authorization
	Token
	User
	Admin
	Domain
	Role
	Project
	Company
	ServiceMain
	ExcelAnalysis
	Object
}

func NewService(repos *repository.Repository) *Service {
	tokenService := NewTokenService(repos.Role, repos.User, repos.AuthType)

	return &Service{
		Token:         tokenService,
		Authorization: NewAuthService(repos.Authorization, *tokenService),
		User:          NewUserService(repos.User),
		Admin:         NewAdminService(repos.Admin),
		Domain:        NewDomainService(repos.Domain),
		Role:          NewRoleService(repos.Role),
		Project:       NewProjectService(repos.Project),
		Company:       NewCompanyService(repos.Company),
		ServiceMain:   NewServiceMainService(repos.ServiceMain),
		ExcelAnalysis: NewExcelAnalysisService(repos.ExcelAnalysis),
		Object:        NewObjectService(repos.Object),
	}
}
