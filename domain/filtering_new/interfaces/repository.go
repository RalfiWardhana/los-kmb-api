package interfaces

import "los-kmb-api/models/entity"

type Repository interface {
	DummyDataPbk(query string) (data entity.DummyPBK, err error)
	SaveDupcheckResult(data entity.FilteringKMB, trxDetailBiro []entity.TrxDetailBiro) (err error)
	GetFilteringByID(prospectID string) (row int, err error)
}
