package handler

import (
	adminModel "main-server/pkg/model/admin"
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
// @Success 200 {object} adminModel.UsersResponseModel "data"
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /admin/user/get/all [post]
func (h *Handler) getAllUsers(c *gin.Context) {
	var data adminModel.UsersResponseModel

	data, err := h.services.Admin.GetAllUsers(c)

	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
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
// @Param logo query string true "logo"
// @Param title query string true "title"
// @Param description query string true "description"
// @Param email_company query string true "email_company"
// @Param email_admin query string true "email_admin"
// @Param phone query string true "phone"
// @Param link query string true "link"
// @Success 200 {object} adminModel.CompanyModel "data"
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /admin/company/create [post]
func (h *Handler) createCompany(c *gin.Context) {
	form, err := c.MultipartForm()

	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
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
	filepath := "public/company/" + newFilename

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
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.SaveUploadedFile(fileImg, filepath)
	c.JSON(http.StatusOK, data)
}
