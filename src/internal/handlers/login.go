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

func Login(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Email
	if val, exists := data["email"]; exists {
		if val == "" {
			errR.Type = "EMAIL_MISSING"
			errR.Code = 1005
			return resR, errR
		}
	} else {
		errR.Type = "EMAIL_EMPTY"
		errR.Code = 1006
		return resR, errR
	}
	// Check Pass
	if val, exists := data["pass"]; exists {
		if val == "" {
			errR.Type = "PASSWORD_MISSING"
			errR.Code = 1007
			return resR, errR
		}
	} else {
		errR.Type = "PASSWORD_EMPTY"
		errR.Code = 1008
		return resR, errR
	}
	// Check Expire Time
	if val, exists := data["expire_in"]; exists {
		if val == "" {
			data["expire_in"] = 60
		}
	} else {
		data["expire_in"] = 60
	}

	// Sanitize and build query
	query := fmt.Sprintf(
		`SELECT display_name FROM users WHERE email='%s' AND password=MD5('%s') LIMIT 1`,
		strings.ToLower(data["email"].(string)),
		data["pass"].(string),
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "LOGIN_GRPC_ERROR"
		errR.Code = 1017
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}

	// Extract gRPC struct
	dataDB := res.Data.GetFields()
	// DB result rows count
	exist := dataDB["count"].GetNumberValue()
	if exist == 0 {
		errR.Type = "INVALID_CREDENTIALS"
		errR.Code = 208
		return resR, errR
	}
	// DB result rows get fields
	userFields := dataDB["rows"].GetListValue().GetValues()[0].GetStructValue().GetFields()
	displayName := ""
	if val, ok := userFields["display_name"]; ok {
		displayName = val.GetStringValue()
	}

	// Generate JWT token
	val, ok := data["expire_in"].(float64)
	if !ok {
		val = 60
	}
	expireIn := int(val)
	duration := time.Duration(expireIn) * time.Minute
	token, err := utils.GenerateJWT(strings.ToLower(data["email"].(string)), duration)
	if err != nil {
		errR.Type = "TOKEN_GENERATION_FAILED"
		errR.Code = 209
		return resR, errR
	}

	// Save JWT in memory for session tracking
	memory.SetToken(token, strings.ToLower(data["email"].(string)), duration)

	// Success
	resR.Type = "login"
	resR.Data = map[string]interface{}{
		"display_name": displayName,
		"token":        token,
	}
	return resR, errR
}
