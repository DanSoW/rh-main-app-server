package admin

import (
	pathConstant "main-server/pkg/constant/path"
	utilContext "main-server/pkg/handler/util"
	adminModel "main-server/pkg/model/admin"
	httpModel "main-server/pkg/model/http"
	userModel "main-server/pkg/model/user"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

// @Summary GetAllUsers
// @Tags admin
// @Description Получение списка всех пользователей находящихся в системе
// @ID admin-user-get-all
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Токен доступа для текущего пользователя" example(Bearer access_token)
// @Success 200 {object} adminModel.UsersResponseModel "data"
// @Failure 400,404 {object} ResponseMessage
// @Failure 500 {object} ResponseMessage
// @Failure default {object} ResponseMessage
// @Router /admin/user/get/all [post]
func (h *AdminHandler) getAllUsers(c *gin.Context) {
	var data adminModel.UsersResponseModel

	data, err := h.services.Admin.GetAllUsers(c)

	if err != nil {
		utilContext.NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, data)
}

// @Summary CreateCompany
// @Tags admin
// @Description Создание новой компании (доступно только администратору)
// @ID admin-company-create
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Токен доступа для текущего пользователя" example(Bearer access_token)
// @Param logo formData string true "Логотип компании"
// @Param title formData string true "Название компании"
// @Param description formData string true "Описание компании"
// @Param email_company formData string true "Адрес электронной почты компании"
// @Param email_admin formData string true "Адрес электронной почты главного администратора "
// @Param phone formData string true "Номер телефона компании"
// @Param link formData string true "Ссылка на сайт компании"
// @Success 200 {object} adminModel.CompanyModel "data"
// @Failure 400,404 {object} ResponseMessage
// @Failure 500 {object} ResponseMessage
// @Failure default {object} ResponseMessage
// @Router /admin/company/create [post]
func (h *AdminHandler) createCompany(c *gin.Context) {
	form, err := c.MultipartForm()

	if err != nil {
		utilContext.NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	file := form.File["logo"]
	fileImg := file[len(file)-1]

	title := c.PostForm("title")
	description := c.PostForm("description")
	emailCompany := c.PostForm("email_company")
	phone := c.PostForm("phone")
	link := c.PostForm("link")
	emailAdmin := c.PostForm("email_admin")

	var data adminModel.CompanyModel

	newFilename := uuid.NewV4().String()
	filepath := pathConstant.PUBLIC_COMPANY + newFilename

	data, err = h.services.Admin.CreateCompany(c, adminModel.CompanyModel{
		Logo:         filepath,
		Title:        title,
		Description:  description,
		Phone:        phone,
		Link:         link,
		EmailCompany: emailCompany,
		EmailAdmin:   emailAdmin,
	})

	if err != nil {
		form.RemoveAll()
		utilContext.NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.SaveUploadedFile(fileImg, filepath)
	c.JSON(http.StatusOK, data)
}

// @Summary SystemUserAddAccess
// @Tags admin
// @Description Добавление доступа какому-либо пользователю
// @ID admin-system-user-add-access
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Токен доступа для текущего пользователя" example(Bearer access_token)
// @Param input body adminModel.SystemPermissionModel true "credentials"
// @Success 200 {object} httpModel.ResponseValue "data"
// @Failure 400,404 {object} ResponseMessage
// @Failure 500 {object} ResponseMessage
// @Failure default {object} ResponseMessage
// @Router /admin/system/user/add/access [post]
func (h *AdminHandler) systemUserAddAccess(c *gin.Context) {
	// Получение пользовательских данных обработанных с помощью цепочки middleware
	userId, userUuid, domainId, err := utilContext.GetContextUserInfo(c)
	if err != nil {
		utilContext.NewErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}

	// Получение данных из запроса
	var input adminModel.SystemPermissionModel
	if err := c.Bind(&input); err != nil {
		utilContext.NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Отправка данных на слой сервисов
	data, err := h.services.Admin.SystemAddManager(&userModel.UserIdentityModel{
		UserId:   userId,
		UserUuid: userUuid,
		DomainId: domainId,
	}, input)

	if err != nil {
		utilContext.NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Возвращение результата
	c.JSON(http.StatusOK, httpModel.ResponseValue{
		Value: data,
	})
}
