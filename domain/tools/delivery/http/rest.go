package http

import (
	"errors"
	"los-kmb-api/middlewares"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/common/platformevent"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"

	"github.com/labstack/echo/v4"
)

type handlerTools struct {
	Json     common.JSON
	producer platformevent.PlatformEventInterface
}

type RequestEncryption struct {
	Encrypt []string `json:"encrypt" example:"hello world"`
	Decrypt []string `json:"decrypt" example:"6gs+t7lBQTYM5SPuqJTNonWLjvmmmc9FaWIj"`
}

type Encryption struct {
	Encrypt string `json:"encrypt" example:"hello world"`
}

type Decryption struct {
	Decrypt string `json:"decrypt" example:"hello world"`
}

func ToolsHandler(kmbroute *echo.Group, json common.JSON, middlewares *middlewares.AccessMiddleware, producer platformevent.PlatformEventInterface) {
	handler := handlerTools{
		Json:     json,
		producer: producer,
	}
	kmbroute.POST("/encrypt-decrypt", handler.EncryptDecrypt, middlewares.AccessMiddleware())
	kmbroute.POST("/encrypt", handler.Encrypt, middlewares.AccessMiddleware())
	kmbroute.POST("/decrypt", handler.Decrypt, middlewares.AccessMiddleware())
	kmbroute.POST("/produce/update-customer/:prospect_id", handler.UpdateCustomer, middlewares.AccessMiddleware())
}

// Encrypt Decrypt Tools godoc
// @Description Encrypt Decrypt
// @Tags Tools
// @Produce json
// @Param body body RequestEncryption true "Body payload"
// @Success 200 {object} response.ApiResponse{data=map[string]string}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/encrypt-decrypt [post]
func (c *handlerTools) EncryptDecrypt(ctx echo.Context) (err error) {
	var req RequestEncryption
	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Encrypt-Decrypt", err)
	}
	data := make(map[string]string)
	for _, v := range req.Encrypt {
		encrypted, errR := utils.PlatformEncryptText(v)
		if errR != nil {
			err = errors.New(constant.ERROR_BAD_REQUEST + " - Encryption Error")
			return c.Json.InternalServerErrorCustomV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Encrypt-Decrypt", err)
		}
		data[v] = encrypted
	}
	for _, v := range req.Decrypt {
		decrypted, errR := utils.PlatformDecryptText(v)
		if errR != nil {
			err = errors.New(constant.ERROR_BAD_REQUEST + " - Decryption Error")
			return c.Json.InternalServerErrorCustomV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Encrypt-Decrypt", err)
		}
		data[v] = decrypted
	}

	return c.Json.SuccessV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Encrypt-Decrypt", req, data)
}

// Encrypt Decrypt Tools godoc
// @Description Encrypt Decrypt
// @Tags Tools
// @Produce json
// @Param body body Encryption true "Body payload"
// @Success 200 {object} response.ApiResponse{data=map[string]string}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/encrypt [post]
func (c *handlerTools) Encrypt(ctx echo.Context) (err error) {
	var req Encryption
	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Encrypt", err)
	}
	encrypted, errR := utils.PlatformEncryptText(req.Encrypt)
	if errR != nil {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Encryption Error")
		return c.Json.InternalServerErrorCustomV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Encrypt", err)
	}

	data := Encryption{
		Encrypt: encrypted,
	}

	return c.Json.SuccessV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Encrypt", req, data)
}

// Encrypt Decrypt Tools godoc
// @Description Encrypt Decrypt
// @Tags Tools
// @Produce json
// @Param body body Decryption true "Body payload"
// @Success 200 {object} response.ApiResponse{data=map[string]string}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/decrypt [post]
func (c *handlerTools) Decrypt(ctx echo.Context) (err error) {
	var req Decryption
	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Decrypt", err)
	}
	decrypted, errR := utils.PlatformDecryptText(req.Decrypt)
	if errR != nil {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Decryption Error")
		return c.Json.InternalServerErrorCustomV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Decrypt", err)
	}

	data := Decryption{
		Decrypt: decrypted,
	}

	return c.Json.SuccessV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Decrypt", req, data)
}

// Produce Messages Update Customer Tools godoc
// @Description Api Produce Messages Update Customer
// @Tags Update Customer
// @Produce json
// @Param prospect_id path string true "Prospect ID"
// @Success 200 {object} response.ApiResponse{data=response.ReasonMessageRow}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/produce/update-customer/{prospect_id} [post]
func (c *handlerTools) UpdateCustomer(ctx echo.Context) (err error) {
	var (
		ctxJson error
	)

	prospectID := ctx.Param("prospect_id")
	if prospectID == "" {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - ProspectID does not exist")
		ctxJson, _ = c.Json.BadRequestErrorBindV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS KMB - Produce Messages - Error", prospectID, err)
		return ctxJson
	}

	req := map[string]interface{}{
		"prospect_id": prospectID,
	}

	err = c.producer.PublishEvent(ctx.Request().Context(), middlewares.UserInfoData.AccessToken, constant.TOPIC_INSERT_CUSTOMER, constant.KEY_PREFIX_UPDATE_CUSTOMER, prospectID, req, 0)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Failed to produce messages with error: " + err.Error())
		ctxJson, _ = c.Json.ServerSideErrorV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS KMB - Produce Messages - Error", prospectID, err)
		return ctxJson
	}

	return c.Json.SuccessV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS KMB - Produce Messages - Success", prospectID, nil)

}
