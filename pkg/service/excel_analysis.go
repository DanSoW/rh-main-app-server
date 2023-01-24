package service

import (
	excelModel "main-server/pkg/model/excel"
	infoModel "main-server/pkg/module/excel_analysis/model"
	repository "main-server/pkg/repository"
)

type ExcelAnalysisService struct {
	repo repository.ExcelAnalysis
}

func NewExcelAnalysisService(repo repository.ExcelAnalysis) *ExcelAnalysisService {
	return &ExcelAnalysisService{
		repo: repo,
	}
}

func (s *ExcelAnalysisService) GetHeaderInfoDocument(document excelModel.DocumentIdModel) (infoModel.HeaderInfoModel, error) {
	return s.repo.GetHeaderInfoDocument(document)
}
