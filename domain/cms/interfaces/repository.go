package interfaces

import (
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

type Repository interface {
	GetSpIndustryTypeMaster() (data []entity.SpIndustryTypeMaster, err error)
	GetCustomerPhoto(prospectID string) (photo []entity.DataPhoto, err error)
	GetReasonPrescreening(req request.ReqReasonPrescreening, pagination interface{}) (reason []entity.ReasonMessage, rowTotal int, err error)
	GetCancelReason(pagination interface{}) (reason []entity.CancelReason, rowTotal int, err error)
	GetApprovalReason(req request.ReqApprovalReason, pagination interface{}) (reason []entity.ApprovalReason, rowTotal int, err error)
	GetSurveyorData(prospectID string) (surveyor []entity.TrxSurveyor, err error)
	GetInquiryPrescreening(req request.ReqInquiryPrescreening, pagination interface{}) (data []entity.InquiryPrescreening, rowTotal int, err error)
	GetTrxStatus(prospectID string) (status entity.TrxStatus, err error)
	SavePrescreening(prescreening entity.TrxPrescreening, detail entity.TrxDetail, status entity.TrxStatus) (err error)
	SaveLogOrchestrator(header, request, response interface{}, path, method, prospectID string, requestID string) (err error)
	GetInquiryCa(req request.ReqInquiryCa, pagination interface{}) (data []entity.InquiryCa, rowTotal int, err error)
	GetHistoryApproval(prospectID string) (history []entity.HistoryApproval, err error)
	GetInternalRecord(prospectID string) (record []entity.TrxInternalRecord, err error)
	SaveDraftData(draft entity.TrxDraftCaDecision) (err error)
	GetLimitApproval(ntf float64) (limit entity.MappingLimitApprovalScheme, err error)
	GetInquirySearch(req request.ReqSearchInquiry, pagination interface{}) (data []entity.InquirySearch, rowTotal int, err error)
	GetAkkk(prospectID string) (data entity.Akkk, err error)
	GetHistoryProcess(prospectID string) (detail []entity.HistoryProcess, err error)
	ProcessTransaction(trxCaDecision entity.TrxCaDecision, trxHistoryApproval entity.TrxHistoryApprovalScheme, trxStatus entity.TrxStatus, trxDetail entity.TrxDetail) (err error)
	ProcessReturnOrder(prospectID string, trxStatus entity.TrxStatus, trxDetail entity.TrxDetail) (err error)
	ProcessRecalculateOrder(prospectID string, trxStatus entity.TrxStatus, trxDetail entity.TrxDetail, trxHistoryApproval entity.TrxHistoryApprovalScheme) (err error)
	GetInquiryApproval(req request.ReqInquiryApproval, pagination interface{}) (data []entity.InquiryCa, rowTotal int, err error)
	SubmitApproval(req request.ReqSubmitApproval, trxStatus entity.TrxStatus, trxDetail entity.TrxDetail, trxRecalculate entity.TrxRecalculate, approval response.RespApprovalScheme) (err error)
	GetAFMobilePhone(prospectID string) (data entity.AFMobilePhone, err error)
	GetRegionBranch(userId string) (data []entity.RegionBranch, err error)
	GetMappingClusterBranch(req request.ReqListMappingCluster, pagination interface{}) (data []entity.MappingClusterBranch, rowTotal int, err error)
}
