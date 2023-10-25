package interfaces

import (
	"context"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

type Usecase interface {
	GetInquiryPrescreening(ctx context.Context, req request.ReqInquiryPrescreening, pagination interface{}) (data []entity.InquiryData, rowTotal int, err error)
	GetReasonPrescreening(ctx context.Context, req request.ReqReasonPrescreening, pagination interface{}) (data []entity.ReasonMessage, rowTotal int, err error)
	ReviewPrescreening(ctx context.Context, req request.ReqReviewPrescreening) (data response.ReviewPrescreening, err error)
	GetInquiryCa(ctx context.Context, req request.ReqInquiryCa, pagination interface{}) (data []entity.InquiryDataCa, rowTotal int, err error)
	SaveAsDraft(ctx context.Context, req request.ReqSaveAsDraft) (data response.CAResponse, err error)
	SubmitDecision(ctx context.Context, req request.ReqSubmitDecision) (data response.CAResponse, err error)
	GetSearchInquiry(ctx context.Context, req request.ReqSearchInquiry, pagination interface{}) (data []entity.InquiryDataSearch, rowTotal int, err error)
	CancelOrder(ctx context.Context, req request.ReqCancelOrder) (data response.CancelResponse, err error)
}
