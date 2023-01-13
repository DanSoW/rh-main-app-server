package excel_analysis

import (
	"context"
	"fmt"
	"io/ioutil"
	"reflect"

	model "main-server/pkg/module/excel_analysis/model"

	"github.com/xuri/excelize/v2"
	"golang.org/x/oauth2/google"
	"gopkg.in/Iwark/spreadsheet.v2"
)

/* Тип предиката для сравнения каждой ячейки таблицы с определённым значением */
type PredicatType func(value string, row, column int) bool

type IExAnalysisKernel interface {
	CopyTo(filePath string) (bool, error)
	GetSheet() (*spreadsheet.Sheet, error)
	GetHeaderInfo() (model.HeaderInfoModel, error)
	GetIndexByValue(value string, sheet *spreadsheet.Sheet) (model.IndexCellModel, error)
	GetIndexNextRow(index model.IndexCellModel, sheet *spreadsheet.Sheet, predicat PredicatType) (model.IndexCellModel, error)
	GetIndexNextRowOffset(index model.IndexCellModel, sheet *spreadsheet.Sheet, offset int, predicat PredicatType) (model.IndexCellModel, error)
	GetLengthCells(index model.IndexCellModel, sheet *spreadsheet.Sheet, predicat PredicatType) int
	GetValueCells(data *model.HeaderInfoModel, sheet *spreadsheet.Sheet, index model.IndexCellModel, place string) model.IndexCellModel
}

type ExAnalysisKernel struct {
	ClientSecretPath string
	DocumentId       string
}

/* Функция создания нового экземпляра ExAnalysisKernel */
func NewExAnalysisKernel(clientSecretPath, documentId string) *ExAnalysisKernel {
	return &ExAnalysisKernel{
		ClientSecretPath: clientSecretPath,
		DocumentId:       documentId,
	}
}

/* Получение длины ячеек начиная с определённого индекса */
func (k *ExAnalysisKernel) GetLengthCells(index model.IndexCellModel, sheet *spreadsheet.Sheet, predicat PredicatType) int {
	if sheet == nil {
		var err error
		if sheet, err = k.GetSheet(); err != nil {
			return 0
		}
	}

	var result int
	result = 0

	for i := index.Row; i < len(sheet.Rows); i++ {
		cell := sheet.Rows[i][index.Column]
		if predicat(cell.Value, i, index.Column) {
			result++
		} else {
			break
		}
	}

	return result
}

/* Получение подробной информации о ячейки в таблице по её значению */
func (k *ExAnalysisKernel) GetIndexByValue(value string, sheet *spreadsheet.Sheet) (model.IndexCellModel, error) {
	if sheet == nil {
		var err error
		if sheet, err = k.GetSheet(); err != nil {
			return model.IndexCellModel{}, err
		}
	}

	var result model.IndexCellModel

	for rInd, row := range sheet.Rows {
		for cInd, cell := range row {
			if cell.Value == value {
				result = model.IndexCellModel{
					Pos:    cell.Pos(),
					Row:    rInd,
					Column: cInd,
					Value:  cell.Value,
				}
				break
			}
		}
	}

	return result, nil
}

/* Получение следующей строки, которая удовлетворяет некоторому условию, определённому в предикате */
func (k *ExAnalysisKernel) GetIndexNextRowOffset(index model.IndexCellModel, sheet *spreadsheet.Sheet, offset int, predicat PredicatType) (model.IndexCellModel, error) {
	if sheet == nil {
		var err error
		if sheet, err = k.GetSheet(); err != nil {
			return model.IndexCellModel{}, err
		}
	}

	var result model.IndexCellModel
	for i := (index.Row + offset); i < len(sheet.Rows); i++ {
		cell := sheet.Rows[i][index.Column]
		if predicat(cell.Value, i, index.Column) {
			result = model.IndexCellModel{
				Pos:    cell.Pos(),
				Row:    i,
				Column: index.Column,
				Value:  cell.Value,
			}
			break
		}
	}

	return result, nil
}

/* Получение следующей строки, которая удовлетворяет некоторому условию, определённому в предикате */
func (k *ExAnalysisKernel) GetIndexNextRow(index model.IndexCellModel, sheet *spreadsheet.Sheet, predicat PredicatType) (model.IndexCellModel, error) {
	if sheet == nil {
		var err error
		if sheet, err = k.GetSheet(); err != nil {
			return model.IndexCellModel{}, err
		}
	}

	var result model.IndexCellModel
	for i := index.Row; i < len(sheet.Rows); i++ {
		cell := sheet.Rows[i][index.Column]
		if predicat(cell.Value, i, index.Column) {
			result = model.IndexCellModel{
				Pos:    cell.Pos(),
				Row:    i,
				Column: index.Column,
				Value:  cell.Value,
			}
			break
		}
	}

	return result, nil
}

/* Получение указателя на объект листа */
func (k *ExAnalysisKernel) GetSheet() (*spreadsheet.Sheet, error) {
	// Чтение данных из файла
	data, err := ioutil.ReadFile(k.ClientSecretPath)
	if err != nil {
		return nil, err
	}

	// Получение конфигурации JWT из чтинанных ранее данных
	conf, err := google.JWTConfigFromJSON(data, spreadsheet.Scope)
	if err != nil {
		return nil, err
	}

	// Создание нового клиента работающего в фоновом режиме
	client := conf.Client(context.Background())

	// Создание нового сервиса spreadsheet с ранее объявленным клиентом
	service2 := spreadsheet.NewServiceWithClient(client)

	// Получение удалённого доступа к таблице
	spreadsheet, err := service2.FetchSpreadsheet(k.DocumentId)
	if err != nil {
		return nil, err
	}

	// Получение листа с индексом 0 (первого листа)
	sheet, err := spreadsheet.SheetByIndex(0)
	if err != nil {
		return nil, err
	}

	return sheet, nil
}

/* Копирование данных из удалённой таблицы в локальную */
func (k *ExAnalysisKernel) CopyTo(filePath string) (bool, error) {
	sheet, err := k.GetSheet()
	if err != nil {
		return false, err
	}

	// Создание нового файла Excel
	f := excelize.NewFile()

	// Копирование данных из одной таблицы в другую
	for _, row := range sheet.Rows {
		for _, cell := range row {
			f.SetCellValue("Sheet1", cell.Pos(), cell.Value)
		}
	}

	// Сохранение нового файла
	if err := f.SaveAs(filePath); err != nil {
		fmt.Println(err)
	}

	return true, nil
}

/* Получение информации о ячейках и её загрузка в структуру */
func (k *ExAnalysisKernel) GetValueCells(data *model.HeaderInfoModel, sheet *spreadsheet.Sheet, index model.IndexCellModel, place string) model.IndexCellModel {
	nextIndex, _ := k.GetIndexNextRowOffset(index, sheet, 1, func(value string, row, column int) bool { return (len(value) > 0) })
	nextIndex_RC := model.IndexCellModel{
		Row:    nextIndex.Row,
		Column: (nextIndex.Column + 1),
	}

	indexLen := k.GetLengthCells(nextIndex_RC, sheet, func(value string, row, column int) bool {
		if len(value) <= 0 {
			return false
		}

		if row == nextIndex.Row {
			return true
		}

		return (len(sheet.Rows[row][column-1].Value) <= 0)
	})

	ps := reflect.ValueOf(data)
	s := ps.Elem()
	var reflectValue reflect.Value

	if s.Kind() == reflect.Struct {
		reflectValue = s.FieldByName(place)

		if !(reflectValue.IsValid() && reflectValue.CanSet()) {
			return model.IndexCellModel{}
		}
	} else {
		return model.IndexCellModel{}
	}

	for i := 0; i < indexLen; i++ {
		nextIndex_RC.Row += i
		next, _ := k.GetIndexNextRow(nextIndex_RC, sheet, func(value string, row, column int) bool {
			if len(value) <= 0 {
				return false
			}

			if row == nextIndex.Row {
				return true
			}

			return (len(sheet.Rows[row][column-1].Value) <= 0)
		})

		if reflectValue.Kind() == reflect.String {
			reflectValue.SetString(next.Value)
		} else if reflectValue.Kind() == reflect.Slice {
			reflectValue.Set(reflect.Append(reflectValue, reflect.ValueOf(next.Value)))
		}
	}

	return nextIndex
}

func (k *ExAnalysisKernel) GetHeaderInfo() (model.HeaderInfoModel, error) {
	sheet, err := k.GetSheet()
	if err != nil {
		return model.HeaderInfoModel{}, err
	}

	var headerInfo model.HeaderInfoModel

	// Определение индекса ячейки с указанием на основную информацию
	index, _ := k.GetIndexByValue("Основная информация", sheet)
	var nextIndex model.IndexCellModel
	nextIndex, _ = k.GetIndexNextRowOffset(index, sheet, 1, func(value string, row, column int) bool { return (len(value) > 0) })

	headerInfo.Title = nextIndex.Value

	// Заполнение полей структуры (начиная с адреса)
	t := reflect.TypeOf(headerInfo)

	for i := 1; i < t.NumField(); i++ {
		nextIndex = k.GetValueCells(&headerInfo, sheet, nextIndex, t.Field(i).Name)
	}

	return headerInfo, nil
}
