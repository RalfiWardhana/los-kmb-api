package interfaces

import (
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
)

type Repository interface {
	GetSpIndustryTypeMaster() (data []entity.SpIndustryTypeMaster, err error)
	GetCustomerPhoto(prospectID string) (photo []entity.TrxCustomerPhoto, err error)
	GetSurveyorData(prospectID string) (surveyor []entity.TrxSurveyor, err error)
	GetInquiryPrescreening(req request.ReqInquiryPrescreening, pagination interface{}) (data []entity.InquiryPrescreening, rowTotal int, err error)
}
