package handlers

import (
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/memory"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"log"
	"time"
)

func Extend(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
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

	val, ok := data["expire_in"].(float64)
	if !ok {
		log.Println("expire_in is not a float64")
		errR.Type = "EXPIRE_IN_INVALID"
		errR.Code = 1034
		return resR, errR
	}
	expireIn := int(val)
	duration := time.Duration(expireIn) * time.Minute
	// Save JWT in memory for session tracking
	memory.ExtendToken(data["token"].(string), duration)

	// Success
	resR.Type = "tokenRenew"
	resR.Data = ""
	return resR, errR
}
