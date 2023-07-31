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
	Json common.JSON
}

func ToolsHandler(kmbroute *echo.Group, json common.JSON, middlewares *middlewares.AccessMiddleware) {
	handler := handlerTools{
		Json: json,
	}
	kmbroute.POST("/encrypt-decrypt", handler.EncryptDecrypt, middlewares.AccessMiddleware())
}

func (c *handlerTools) EncryptDecrypt(ctx echo.Context) (err error) {
	type RequestEncryption struct {
		Encrypt []string `json:"encrypt"`
		Decrypt []string `json:"decrypt"`
	}
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
