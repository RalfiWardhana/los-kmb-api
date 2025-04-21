package platformauth

import (
	"errors"
	"los-kmb-api/shared/constant"

	"os"
	"strings"

	"github.com/KB-FMF/platform-library/auth"
)

func PlatformVerify(token string) (err error) {

	env := os.Getenv("APP_ENV")

	if strings.Contains(strings.ToLower(env), "production") {
		env = auth.ENV_PRODUCTION
	} else if strings.Contains(strings.ToLower(env), "staging") {
		env = auth.ENV_STAGING
	} else {
		env = auth.ENV_DEVELOPMENT
	}

	auth := auth.New(env)

	resp, authErr := auth.Validation(token, "los-kmb-api")

	if authErr != nil {
		return errors.New(constant.ERROR_UNAUTHORIZED + " - " + "invalid")
	}

	if resp.Data["status"] != "active" {
		return errors.New(constant.ERROR_UNAUTHORIZED + " - " + "expired")
	}

	return
}
