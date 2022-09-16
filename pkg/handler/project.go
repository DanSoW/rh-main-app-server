package handler

import (
	middlewareConstant "main-server/pkg/constant/middleware"
	projectModel "main-server/pkg/model/project"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

// @Summary CreateProject
// @Tags project
// @Description Get all users, which are located in system
// @ID get-all-users
// @Accept  json
// @Produce  json
// @Success 200 {object} adminModel.UsersResponseModel "data"
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /company/project/create [post]
func (h *Handler) createProject(c *gin.Context) {
	var input projectModel.ProjectModel

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	userId, _ := c.Get(middlewareConstant.USER_CTX)
	domainId, _ := c.Get(middlewareConstant.DOMAINS_ID)

	data, err := h.services.Project.CreateProject(userId.(int), domainId.(int), input)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, data)
}

// @Summary AddLogoProject
// @Tags project
// @Description Add logo to project
// @ID add-logo-project
// @Accept  json
// @Produce  json
// @Success 200 {object} adminModel.UsersResponseModel "data"
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /company/project/add/logo [post]
func (h *Handler) addLogoProject(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	file := form.File["logo"]
	fileImg := file[len(file)-1]
	projectUuid := c.PostForm("uuid")

	newFilename := uuid.NewV4().String()
	filepath := "public/project/" + newFilename

	userId, _ := c.Get(middlewareConstant.USER_CTX)
	domainId, _ := c.Get(middlewareConstant.DOMAINS_ID)

	data, err := h.services.Project.AddLogoProject(
		userId.(int),
		domainId.(int),
		projectModel.ProjectLogoModel{
			Filepath: filepath,
			Uuid:     projectUuid,
		},
	)

	if err != nil {
		form.RemoveAll()
		newErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}

	c.SaveUploadedFile(fileImg, filepath)
	c.JSON(http.StatusOK, data)
}
