package middlewares

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"los-kmb-api/models/entity"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type bodyDumpMiddleware struct {
	db *gorm.DB
}

func NewBodyDumpMiddleware(db *gorm.DB) *bodyDumpMiddleware {
	return &bodyDumpMiddleware{db: db}
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
				isSave            bool
				isSaveTrxKPMError bool
				prospectID        string
				trxKPMError       entity.TrxKPMError
				trxError          entity.TrxPrincipleError
				trxStepOne        entity.TrxPrincipleStepOne
				kpmId, step       int
			)
			if e.Response().Status != 200 {

				kpmId, prospectID = GetPrinciplePayload(reqBody)

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

				if e.Response().Status == http.StatusGatewayTimeout {
					id := utils.GenerateUUID()
					m.db.Create(&entity.TrxKPMStatus{
						ID:         id,
						ProspectID: prospectID,
						Decision:   constant.STATUS_KPM_ERROR_2WILEN,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					})
				}
			}

		},
	}
}

func GetPrinciplePayload(payload []byte) (kpmID int, prospectID string) {

	type KpmID struct {
		KPMId      int    `json:"kpm_id"`
		ProspectID string `json:"prospect_id"`
	}

	var data KpmID

	_ = json.Unmarshal(payload, &data)

	return data.KPMId, data.ProspectID
}
