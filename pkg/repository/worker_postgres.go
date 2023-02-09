package repository

import (
	"errors"
	"fmt"
	tableConstant "main-server/pkg/constant/table"
	workerModel "main-server/pkg/model/worker"

	"github.com/jmoiron/sqlx"
)

type WorkerPostgres struct {
	db      *sqlx.DB
	company *CompanyPostgres
}

func NewWorkerPostgres(db *sqlx.DB, company *CompanyPostgres) *WorkerPostgres {
	return &WorkerPostgres{
		db:      db,
		company: company,
	}
}

/* Получение экземлпяра объекта таблицы */
func (r *WorkerPostgres) Get(column string, value interface{}, check bool) (*workerModel.WorkerModel, error) {
	var worker workerModel.WorkerModel
	workerEx, err := r.GetEx(column, value, check)
	if err != nil {
		return nil, err
	}

	worker.Worker = workerEx
	companies, err := r.company.GetByWorker(workerEx.Id, false)
	if err != nil {
		return nil, err
	}

	// Ситуация, когда worker есть, а его компании нет
	if companies == nil || len(companies) <= 0 {
		return nil, errors.New("Ошибка: компания worker'a не присутствует в базе данных!")
	}
	worker.Company = &companies[len(companies)-1]

	return &worker, nil
}

/* Получение расширенного экземпляра объекта таблицы */
func (r *WorkerPostgres) GetEx(column string, value interface{}, check bool) (*workerModel.WorkerDbExModel, error) {
	var worker []workerModel.WorkerDbExModel
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1 LIMIT 1", tableConstant.CB_WORKERS, column)

	var err error

	switch value.(type) {
	case int:
		err = r.db.Select(&worker, query, value.(int))
		break
	case string:
		err = r.db.Select(&worker, query, value.(string))
		break
	}

	if len(worker) <= 0 {
		if check {
			return nil, errors.New(fmt.Sprintf("Ошибка: экземпляра объекта worker по запросу %s:%s не найдено!", column, value))
		}

		return nil, nil
	}

	return &worker[len(worker)-1], err
}
