package interfaces

import "los-kmb-api/models/dto"

type Repository interface {
	GetAuth(trx dto.AuthModel) (auth dto.AuthJoinTable, err error)
}
