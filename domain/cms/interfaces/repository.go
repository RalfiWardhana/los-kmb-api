package interfaces

import (
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
)

type Repository interface {
	GetSpIndustryTypeMaster() (data []entity.SpIndustryTypeMaster, err error)
	GetCustomerPhoto(prospectID string) (photo []entity.DataPhoto, err error)
	GetReasonPrescreening(req request.ReqReasonPrescreening, pagination interface{}) (reason []entity.ReasonMessage, rowTotal int, err error)
	GetSurveyorData(prospectID string) (surveyor []entity.TrxSurveyor, err error)
	GetInquiryPrescreening(req request.ReqInquiryPrescreening, pagination interface{}) (data []entity.InquiryPrescreening, rowTotal int, err error)
	GetStatusPrescreening(prospectID string) (status entity.TrxStatus, err error)
	SavePrescreening(prescreening entity.TrxPrescreening, detail entity.TrxDetail, status entity.TrxStatus) (err error)
	SaveLogOrchestrator(header, request, response interface{}, path, method, prospectID string, requestID string) (err error)
	GetInquiryCa(req request.ReqInquiryCa, pagination interface{}) (data []entity.InquiryCa, rowTotal int, err error)
	GetHistoryApproval(prospectID string) (history []entity.TrxHistoryApprovalScheme, err error)
	GetInternalRecord(prospectID string) (record []entity.TrxInternalRecord, err error)
	SaveDraftData(draft entity.TrxDraftCaDecision) (err error)
}
