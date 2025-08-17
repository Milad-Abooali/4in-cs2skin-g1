package handlers

import (
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/grpcclient"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/memory"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/utils"
	"strings"
	"time"
)

func PassRecovery(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Email
	if val, exists := data["email"]; exists {
		if val == "" {
			errR.Type = "EMAIL_EMPTY"
			errR.Code = 1006
			return resR, errR
		}
	} else {
		errR.Type = "EMAIL_MISSING"
		errR.Code = 1005
		return resR, errR
	}

	email := strings.ToLower(data["email"].(string))

	// Build query
	query := fmt.Sprintf(`SELECT id FROM users WHERE email='%s' LIMIT 1`, email)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "PASSWORD_RECOVERY_GRPC_ERROR"
		errR.Code = 1018
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}

	// Parse gRPC response
	dataDB := res.Data.GetFields()
	if dataDB["count"].GetNumberValue() == 0 {
		errR.Type = "PASSWORD_RECOVERY_DB_ERROR"
		errR.Code = 1019
		return resR, errR
	}

	userFields := dataDB["rows"].GetListValue().GetValues()[0].GetStructValue().GetFields()
	id := 0
	if val, ok := userFields["id"]; ok {
		id = int(val.GetNumberValue())
	}

	// Generate JWT token - 1 Day
	expireIn := 1440
	duration := time.Duration(expireIn) * time.Minute
	token, err := utils.GenerateJWT(email, duration)
	if err != nil {
		errR.Type = "TOKEN_GENERATION_FAILED"
		errR.Code = 209
		return resR, errR
	}

	// Store recovery token in memory
	memory.SetRecoveryToken(token, email, duration)

	// Success
	resR.Type = "passRecovery"
	resR.Data = id
	return resR, errR
}
