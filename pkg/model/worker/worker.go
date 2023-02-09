package worker

import (
	companyModel "main-server/pkg/model/company"
	projectModel "main-server/pkg/model/project"
)

type WorkerModel struct {
	Worker   *WorkerDbExModel               `json:"worker"`
	Company  *companyModel.CompanyDbExModel `json:"company"`
	Projects *projectModel.ProjectDbDataEx  `json:"projects"`
}
