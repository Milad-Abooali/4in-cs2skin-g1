package handlers

import (
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/memory"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/utils"
	"time"
)

func Renew(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
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

	// Delete JWT in memory for session tracking
	oldToken, ok := memory.GetToken(data["token"].(string))
	if !ok {
		errR.Type = "TOKEN_NOT_FOUND"
		errR.Code = 1032
		return resR, errR
	}

	memory.DeleteToken(data["token"].(string))

	expireIn := 60
	duration := time.Duration(expireIn) * time.Minute
	token, err := utils.GenerateJWT(oldToken.Email, duration)
	if err != nil {
		errR.Type = "TOKEN_GENERATION_FAILED"
		errR.Code = 209
		return resR, errR
	}

	// Save JWT in memory for session tracking
	memory.SetToken(token, oldToken.Email, duration)

	// Success
	resR.Type = "tokenRenew"
	resR.Data = token
	return resR, errR
}
