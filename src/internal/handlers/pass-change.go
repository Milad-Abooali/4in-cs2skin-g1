package handlers

import (
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/grpcclient"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"strings"
)

func PassChange(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Get Email
	email, ok := GetUserEmail(data)
	if !ok {
		errR.Type = "TOKEN_NOT_FOUND"
		errR.Code = 1032
		return resR, errR
	}

	if _, exists := data["pass"]; exists {
		if data["pass"] == data["new_pass"] {
			errR.Type = "PASSWORD_SAME"
			errR.Code = 1036
			return resR, errR
		}
	} else {
		errR.Type = "PASSWORD_MISSING"
		errR.Code = 1007
		return resR, errR
	}

	// Check Pass New
	if val, exists := data["new_pass"]; exists {
		strVal, ok := val.(string)
		if !ok {
			errR.Type = "PASSWORD_MISSING"
			errR.Code = 1007
			return resR, errR
		}

		if strVal == "" {
			errR.Type = "PASSWORD_EMPTY"
			errR.Code = 1008
			return resR, errR
		}

		if len(strVal) < 8 || !hasNumber(strVal) || !hasLetter(strVal) {
			errR.Type = "PASSWORD_TOO_WEAK"
			errR.Code = 1009
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

	query := fmt.Sprintf(
		`SELECT display_name FROM users WHERE email='%s' AND password=MD5('%s') LIMIT 1`,
		strings.ToLower(email),
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

	// Prepare update query
	query = fmt.Sprintf(
		"UPDATE users SET `password`=MD5('%s') WHERE email='%s' AND `password`=MD5('%s')",
		data["new_pass"].(string),
		strings.ToLower(email),
		data["pass"].(string),
	)

	// gRPC Call
	res, err = grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "PASSWORD_RESET_GRPC_ERROR"
		errR.Code = 1019
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}

	// Check update result
	dataDB = res.Data.GetFields()
	if dataDB["rows_affected"].GetNumberValue() == 0 {
		errR.Type = "PASSWORD_NOT_CHANGED"
		errR.Code = 213
		return resR, errR
	}

	// Success
	resR.Type = "passChange"
	resR.Data = ""
	return resR, errR
}
