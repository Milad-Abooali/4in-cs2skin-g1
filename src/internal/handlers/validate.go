package handlers

import (
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/memory"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
)

func Validate(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Token
	if val, exists := data["token"]; exists {
		if val == "" {
			errR.Type = "TOKEN_EMPTY"
			errR.Code = 1031
			return resR, errR
		}
	} else {
		errR.Type = "TOKEN_MISSING"
		errR.Code = 1030
		return resR, errR
	}

	// Get Email
	_, ok := memory.GetToken(data["token"].(string))
	if !ok {
		errR.Type = "TOKEN_NOT_FOUND"
		errR.Code = 1032
		return resR, errR
	}

	ok = memory.ValidateToken(data["token"].(string))
	if !ok {
		errR.Type = "TOKEN_EXPIRED"
		errR.Code = 1033
		return resR, errR
	}

	// Success
	resR.Type = "tokenValidate"
	resR.Data = ""
	return resR, errR
}
