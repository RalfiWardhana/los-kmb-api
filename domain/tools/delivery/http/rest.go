package http

import (
	"errors"
	"los-kmb-api/middlewares"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"

	"github.com/labstack/echo/v4"
)

type handlerTools struct {
	Json        common.JSON
	customUtils utils.UtilsInterface
}

type RequestEncryption struct {
	Encrypt []string `json:"encrypt" example:"hello world"`
	Decrypt []string `json:"decrypt" example:"6gs+t7lBQTYM5SPuqJTNonWLjvmmmc9FaWIj"`
}

func ToolsHandler(kmbroute *echo.Group, json common.JSON, middlewares *middlewares.AccessMiddleware, customUtils utils.UtilsInterface) {
	handler := handlerTools{
		Json:        json,
		customUtils: customUtils,
	}
	kmbroute.POST("/encrypt-decrypt", handler.EncryptDecrypt, middlewares.AccessMiddleware())
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
		encrypted, errR := c.customUtils.PlatformEncryptText(v)
		if errR != nil {
			err = errors.New(constant.ERROR_BAD_REQUEST + " - Encryption Error")
			return c.Json.InternalServerErrorCustomV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Encrypt-Decrypt", err)
		}
		data[v] = encrypted
	}
	for _, v := range req.Decrypt {
		decrypted, errR := c.customUtils.PlatformDecryptText(v)
		if errR != nil {
			err = errors.New(constant.ERROR_BAD_REQUEST + " - Decryption Error")
			return c.Json.InternalServerErrorCustomV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Encrypt-Decrypt", err)
		}
		data[v] = decrypted
	}

	return c.Json.SuccessV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Encrypt-Decrypt", req, data)
}
