package interfaces

import (
	"context"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"mime/multipart"
)

type Usecase interface {
	GetInquiryPrescreening(ctx context.Context, req request.ReqInquiryPrescreening, pagination interface{}) (data []entity.InquiryData, rowTotal int, err error)
	GetReasonPrescreening(ctx context.Context, req request.ReqReasonPrescreening, pagination interface{}) (data []entity.ReasonMessage, rowTotal int, err error)
	ReviewPrescreening(ctx context.Context, req request.ReqReviewPrescreening) (data response.ReviewPrescreening, err error)
	GetInquiryCa(ctx context.Context, req request.ReqInquiryCa, pagination interface{}) (data []entity.InquiryDataCa, rowTotal int, err error)
	SaveAsDraft(ctx context.Context, req request.ReqSaveAsDraft) (data response.CAResponse, err error)
	SubmitDecision(ctx context.Context, req request.ReqSubmitDecision) (data response.CAResponse, err error)
	GetSearchInquiry(ctx context.Context, req request.ReqSearchInquiry, pagination interface{}) (data []entity.InquiryDataSearch, rowTotal int, err error)
	GetAkkk(prospectID string) (data entity.Akkk, err error)
	SubmitNE(ctx context.Context, req request.MetricsNE) (data interface{}, err error)
	GetInquiryNE(ctx context.Context, req request.ReqInquiryNE, pagination interface{}) (data []entity.InquiryDataNE, rowTotal int, err error)
	GetInquiryNEDetail(ctx context.Context, prospectID string) (data request.MetricsNE, err error)
	CancelOrder(ctx context.Context, req request.ReqCancelOrder) (data response.CancelResponse, err error)
	GetCancelReason(ctx context.Context, pagination interface{}) (data []entity.CancelReason, rowTotal int, err error)
	ReturnOrder(ctx context.Context, req request.ReqReturnOrder) (data response.ReturnResponse, err error)
	RecalculateOrder(ctx context.Context, req request.ReqRecalculateOrder, accessToken string) (data response.RecalculateResponse, err error)
	GetInquiryApproval(ctx context.Context, req request.ReqInquiryApproval, pagination interface{}) (data []entity.InquiryDataApproval, rowTotal int, err error)
	GetApprovalReason(ctx context.Context, req request.ReqApprovalReason, pagination interface{}) (data []entity.ApprovalReason, rowTotal int, err error)
	SubmitApproval(ctx context.Context, req request.ReqSubmitApproval) (data response.ApprovalResponse, err error)
	GetInquiryMappingCluster(req request.ReqListMappingCluster, pagination interface{}) (data []entity.InquiryMappingCluster, rowTotal int, err error)
	GenerateExcelMappingCluster() (genName, fileName string, err error)
	UpdateMappingCluster(req request.ReqUploadMappingCluster, file multipart.File) (err error)
	GetMappingClusterBranch(req request.ReqListMappingClusterBranch) (data []entity.ConfinsBranch, err error)
	GetMappingClusterChangeLog(pagination interface{}) (data []entity.MappingClusterChangeLog, rowTotal int, err error)
	GenerateFormAKKK(ctx context.Context, req request.RequestGenerateFormAKKK, accessToken string) (data interface{}, err error)
}
