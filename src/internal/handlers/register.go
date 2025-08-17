package handlers

import (
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/grpcclient"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/memory"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/utils"
	"strings"
	"time"
	"unicode"
)

func hasNumber(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func Register(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
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

	// Check Display Name
	if val, exists := data["d_name"]; exists {
		if val == "" {
			errR.Type = "DISPLAY_NAME_EMPTY"
			errR.Code = 1012
			return resR, errR
		}
	} else {
		errR.Type = "DISPLAY_NAME_MISSING"
		errR.Code = 1011
		return resR, errR
	}
	// Check First Name
	if val, exists := data["f_name"]; exists {
		if val == "" {
			data["f_name"] = "-"
		}
	} else {
		data["f_name"] = "-"
	}
	// Check Last Name
	if val, exists := data["l_name"]; exists {
		if val == "" {
			data["l_name"] = "-"
		}
	} else {
		data["l_name"] = "-"
	}

	// gRPC Call Check exist
	existsQuery := fmt.Sprintf(`SELECT id FROM users WHERE email = "%s" LIMIT 1`, strings.ToLower(data["email"].(string)))
	existsRes, err := grpcclient.SendQuery(existsQuery)
	if err == nil && existsRes != nil && existsRes.Status == "ok" {
		// Try to get ID from first row
		rows := existsRes.Data.Fields["rows"].GetListValue().Values
		if len(rows) > 0 {
			firstRow := rows[0].GetStructValue()
			idField, ok := firstRow.Fields["id"]
			if ok && idField.GetNumberValue() > 0 {
				errR.Type = "EMAIL_ALREADY_EXISTS"
				errR.Code = 1015
				return resR, errR
			}
		}
	}

	// Sanitize and build query
	query := fmt.Sprintf(
		`INSERT INTO users (email, password, first_name, last_name, display_name) 
				VALUES ('%s', MD5('%s'), '%s', '%s', '%s')`,
		strings.ToLower(data["email"].(string)),
		data["pass"].(string),
		data["f_name"].(string),
		data["l_name"].(string),
		data["d_name"].(string),
	)

	// gRPC Call Insert User
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "REGISTER_GRPC_ERROR"
		errR.Code = 1013
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}

	// Extract inserted_id from nested struct
	dataDB := res.Data.GetFields()
	id := int(dataDB["inserted_id"].GetNumberValue())
	if id < 1 {
		errR.Type = "REGISTER_DB_ERROR"
		errR.Code = 1014
		return resR, errR
	}

	// Generate JWT token
	val, ok := data["expire_in"].(float64)
	if !ok {
		val = 60
	}
	expireIn := int(val)
	duration := time.Duration(expireIn) * time.Minute
	email := strings.ToLower(data["email"].(string))

	token, err := utils.GenerateJWT(email, duration)
	if err != nil {
		errR.Type = "token_generation_failed"
		errR.Code = 209
		return resR, errR
	}

	// Save JWT in memory
	memory.SetToken(token, email, duration)

	// Success
	resR.Type = "register"
	resR.Data = map[string]interface{}{
		"user_id":      id,
		"display_name": data["d_name"].(string),
		"token":        token,
	}
	return resR, errR
}
