package handlers

import (
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/memory"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
)

func Logout(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Token
	val, exists := data["token"]
	if !exists {
		errR.Type = "TOKEN_MISSING"
		errR.Code = 1030
		return resR, errR
	}
	tokenStr, ok := val.(string)
	if !ok || tokenStr == "" {
		errR.Type = "TOKEN_EMPTY"
		errR.Code = 1031
		return resR, errR
	}

	// Validate Token in memory
	_, ok = memory.GetToken(tokenStr)
	if !ok {
		errR.Type = "TOKEN_NOT_FOUND"
		errR.Code = 1032
		return resR, errR
	}

	// Delete Token
	memory.DeleteToken(tokenStr)

	// Success
	resR.Type = "logout"
	resR.Data = ""
	return resR, errR
}
