package handlers

import (
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/grpcclient"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/memory"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"strings"
)

func PassReset(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check password
	if val, exists := data["pass"]; exists {
		if val == "" {
			errR.Type = "PASSWORD_EMPTY"
			errR.Code = 1008
			return resR, errR
		}
	} else {
		errR.Type = "PASSWORD_MISSING"
		errR.Code = 1007
		return resR, errR
	}

	// Check token
	if val, exists := data["token"]; exists {
		if val == "" {
			errR.Type = "TOKEN_EXPECTED"
			errR.Code = 1017
			return resR, errR
		}
	} else {
		errR.Type = "TOKEN_EXPECTED"
		errR.Code = 1017
		return resR, errR
	}

	// Validate token in memory
	email, ok := memory.GetEmailByRecoveryToken(data["token"].(string))
	if !ok {
		errR.Type = "TOKEN_NOT_FOUND"
		errR.Code = 1018
		return resR, errR
	}

	// Prepare update query
	query := fmt.Sprintf(
		"UPDATE users SET `password`=MD5('%s') WHERE email='%s'",
		data["pass"].(string),
		strings.ToLower(email),
	)

	// Invalidate token
	memory.DeleteRecoveryToken(email)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "PASSWORD_RESET_GRPC_ERROR"
		errR.Code = 1019
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}

	// Check update result
	dataDB := res.Data.GetFields()
	if dataDB["rows_affected"].GetNumberValue() == 0 {
		errR.Type = "PASSWORD_NOT_CHANGED"
		errR.Code = 213
		return resR, errR
	}

	// Success
	resR.Type = "passReset"
	resR.Data = ""
	return resR, errR
}
