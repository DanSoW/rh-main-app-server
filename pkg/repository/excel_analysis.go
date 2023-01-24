package repository

import (
	excelModel "main-server/pkg/model/excel"
	"main-server/pkg/module/excel_analysis"
	infoModel "main-server/pkg/module/excel_analysis/model"

	"github.com/spf13/viper"
)

type ExcelAnalysisRepository struct{}

/* Создание нового экземпляра объекта ExcelAnalysis*/
func NewExcelAnalysis() *ExcelAnalysisRepository {
	return &ExcelAnalysisRepository{}
}

/* Получение основной информации из Excel-документа */
func (r *ExcelAnalysisRepository) GetHeaderInfoDocument(document excelModel.DocumentIdModel) (infoModel.HeaderInfoModel, error) {
	exKernel := excel_analysis.NewExAnalysisKernel(viper.GetString("paths.client_secret"), document.DocumentId)

	// Создание локальной копии таблицы
	exKernel.CopyTo("Result.xlsx")

	// Возвращение основной информации из Excel-таблицы
	return exKernel.GetHeaderInfo()
}
