package middlewares

import (
	"encoding/json"
	"strings"
	"time"

	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/common/platformevent"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type bodyDumpMiddleware struct {
	db       *gorm.DB
	producer platformevent.PlatformEventInterface
}

func NewBodyDumpMiddleware(db *gorm.DB, producer platformevent.PlatformEventInterface) *bodyDumpMiddleware {
	return &bodyDumpMiddleware{
		db:       db,
		producer: producer,
	}
}
func (m *bodyDumpMiddleware) BodyDumpConfig() middleware.BodyDumpConfig {
	return middleware.BodyDumpConfig{
		Skipper: func(ctx echo.Context) bool {
			if strings.EqualFold(ctx.Request().RequestURI, "/swagger/index.html") {
				return true
			}
			if strings.EqualFold(ctx.Request().RequestURI, "/swagger/swagger-ui-standalone-preset.js") {
				return true
			}
			if strings.EqualFold(ctx.Request().RequestURI, "/swagger/swagger-ui.css") {
				return true
			}
			if strings.EqualFold(ctx.Request().RequestURI, "/swagger/favicon-32x32.png") {
				return true
			}
			if strings.EqualFold(ctx.Request().RequestURI, "/swagger/doc.json") {
				return true
			}
			if strings.EqualFold(ctx.Request().RequestURI, "/swagger/swagger-ui-bundle.js") {
				return true
			}
			return false
		},
		Handler: func(e echo.Context, reqBody []byte, resBody []byte) {

			var (
				isSave                                        bool
				isSaveTrxKPMError                             bool
				isSaveTrxKPMStatus                            bool
				prospectID, assetCode, branchID, referralCode string
				trxKPMError                                   entity.TrxKPMError
				trxKPMStatus                                  entity.TrxKPMStatus
				trxError                                      entity.TrxPrincipleError
				trxStepOne                                    entity.TrxPrincipleStepOne
				kpmId, step                                   int
			)
			if e.Response().Status != 200 {

				kpmId, prospectID, assetCode, branchID, referralCode = GetPrinciplePayload(reqBody)

				switch e.Request().URL.Path {
				case "/api/v3/kmb/verify-asset":
					result := m.db.Raw("SELECT KpmID FROM trx_principle_error WITH (nolock) WHERE KpmID = ? AND step = 1 AND created_at >= DATEADD (HOUR , -1 , GETDATE())", kpmId).Scan(&trxError)
					if result.RowsAffected < 3 {
						isSave = true
						step = 1
					}
				case "/api/v3/kmb/verify-pemohon":
					m.db.Raw("SELECT KPMID FROM trx_principle_step_one WITH (nolock) WHERE ProspectID = ?", prospectID).Scan(&trxStepOne)
					kpmId = trxStepOne.KPMID
					result := m.db.Raw("SELECT KpmID FROM trx_principle_error WITH (nolock) WHERE KpmID = ? AND step = 2 AND created_at >= DATEADD (HOUR , -1 , GETDATE())", kpmId).Scan(&trxError)
					if result.RowsAffected < 3 {
						isSave = true
						step = 2
					}
				case "/api/v3/kmb/verify-pembiayaan ":
					m.db.Raw("SELECT KPMID FROM trx_principle_step_one WITH (nolock) WHERE ProspectID = ?", prospectID).Scan(&trxStepOne)
					kpmId = trxStepOne.KPMID
					result := m.db.Raw("SELECT KpmID FROM trx_principle_error WITH (nolock) WHERE KpmID = ? AND step = 2 AND created_at >= DATEADD (HOUR , -1 , GETDATE())", kpmId).Scan(&trxError)
					if result.RowsAffected < 3 {
						isSave = true
						step = 3
					}
				case "/api/v3/kmb/submission-2wilen":
					result := m.db.Raw("SELECT KpmID FROM trx_kpm_error WITH (nolock) WHERE KpmID = ? AND created_at >= DATEADD (HOUR , -1 , GETDATE())", kpmId).Scan(&trxKPMError)
					if result.RowsAffected < 3 {
						isSaveTrxKPMError = true
					}
					resultTrxKpmStatus := m.db.Raw("SELECT ProspectID FROM trx_kpm_status WITH (nolock) WHERE ProspectID = ? AND decision = 'KPM-ERROR'", prospectID).Scan(&trxKPMStatus)
					if e.Response().Status == 504 && resultTrxKpmStatus.RowsAffected == 0 {
						isSaveTrxKPMStatus = true
					}
				}
			}

			if isSave {
				m.db.Create(&entity.TrxPrincipleError{
					ProspectID: prospectID,
					KpmId:      kpmId,
					Step:       step,
					CreatedAt:  time.Now(),
				})
			}

			if isSaveTrxKPMError {
				m.db.Create(&entity.TrxKPMError{
					ProspectID: prospectID,
					KpmId:      kpmId,
					CreatedAt:  time.Now(),
				})
			}

			if isSaveTrxKPMStatus {
				m.db.Create(&entity.TrxKPMStatus{
					ProspectID: prospectID,
					Decision:   constant.STATUS_KPM_ERROR_2WILEN,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				})

				m.producer.PublishEvent(e.Request().Context(), UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_2WILEN, constant.KEY_PREFIX_UPDATE_TRANSACTION_PRINCIPLE, prospectID, utils.StructToMap(request.Update2wPrincipleTransaction{
					OrderID:                    prospectID,
					KpmID:                      kpmId,
					Source:                     3,
					StatusCode:                 constant.STATUS_KPM_ERROR_2WILEN,
					ProductName:                assetCode,
					BranchCode:                 branchID,
					AssetTypeCode:              constant.KPM_ASSET_TYPE_CODE_MOTOR,
					ReferralCode:               referralCode,
					Is2wPrincipleApprovalOrder: true,
				}), 0)
			}

		},
	}
}

func GetPrinciplePayload(payload []byte) (kpmID int, prospectID, assetCode, branchID, referralCode string) {

	type KpmID struct {
		KPMId        int    `json:"kpm_id"`
		ProspectID   string `json:"prospect_id"`
		AssetCode    string `json:"asset_code"`
		BranchID     string `json:"branch_id"`
		ReferralCode string `json:"referral_code"`
	}

	var data KpmID

	_ = json.Unmarshal(payload, &data)

	return data.KPMId, data.ProspectID, data.AssetCode, data.BranchID, data.ReferralCode
}
