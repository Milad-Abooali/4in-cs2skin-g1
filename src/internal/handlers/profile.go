package handlers

import (
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/grpcclient"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/memory"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"google.golang.org/protobuf/types/known/structpb"
	"strings"
)

func GetUserEmail(data map[string]interface{}) (string, bool) {
	var (
		errR models.HandlerError
	)

	// Check Token
	if val, exists := data["token"]; exists {
		if val == "" {
			errR.Type = "TOKEN_EMPTY"
			errR.Code = 1031
			return "", false
		}
	} else {
		errR.Type = "TOKEN_EXPECTED"
		errR.Code = 1017
		return "", false
	}

	// Get Email
	token, ok := memory.GetToken(data["token"].(string))
	if !ok {
		errR.Type = "TOKEN_NOT_FOUND"
		errR.Code = 1032
		return "", false
	}

	ok = memory.ValidateToken(data["token"].(string))
	if !ok {
		errR.Type = "TOKEN_NOT_FOUND"
		errR.Code = 1032
		return "", false
	}

	// Success - register
	return token.Email, true
}

func GetUserId(email string) (int64, bool) {

	// Sanitize and build query
	query := fmt.Sprintf(
		`SELECT id FROM users u WHERE email='%s' `,
		strings.ToLower(email),
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		return 0, false
	}

	// Extract gRPC struct
	dataDB := res.Data.GetFields()
	// DB result rows count
	exist := dataDB["count"].GetNumberValue()
	if exist == 0 {
		return 0, false
	}
	// DB result rows get fields
	userFields := dataDB["rows"].GetListValue().GetValues()[0].GetStructValue().GetFields()["id"].GetNumberValue()

	return int64(userFields), true

}

func JWT2UserID(data map[string]interface{}) (int64, bool) {

	var (
		errR models.HandlerError
	)

	// Check Token
	if val, exists := data["token"]; exists {
		if val == "" {
			errR.Type = "TOKEN_EMPTY"
			errR.Code = 1031
			return 0, false
		}
	} else {
		errR.Type = "TOKEN_EXPECTED"
		errR.Code = 1017
		return 0, false
	}

	// Get Email
	token, ok := memory.GetToken(data["token"].(string))
	if !ok {
		errR.Type = "TOKEN_NOT_FOUND"
		errR.Code = 1032
		return 0, false
	}

	ok = memory.ValidateToken(data["token"].(string))
	if !ok {
		errR.Type = "TOKEN_NOT_FOUND"
		errR.Code = 1032
		return 0, false
	}

	email := token.Email

	// Sanitize and build query
	query := fmt.Sprintf(
		`SELECT id FROM users u WHERE email='%s' `,
		strings.ToLower(email),
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		return 0, false
	}

	// Extract gRPC struct
	dataDB := res.Data.GetFields()
	// DB result rows count
	exist := dataDB["count"].GetNumberValue()
	if exist == 0 {
		return 0, false
	}
	// DB result rows get fields
	userFields := dataDB["rows"].GetListValue().GetValues()[0].GetStructValue().GetFields()["id"].GetNumberValue()

	return int64(userFields), true

}

func GetProfile(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
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
		errR.Type = "TOKEN_EXPECTED"
		errR.Code = 1017
		return resR, errR
	}

	// Get Email
	email, ok := GetUserEmail(data)
	if !ok {
		errR.Type = "TOKEN_NOT_FOUND"
		errR.Code = 1032
		return resR, errR
	}

	// Sanitize and build query
	query := fmt.Sprintf(
		`SELECT *,'******' password FROM users WHERE email='%s' LIMIT 1`,
		strings.ToLower(email),
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "PROFILE_GRPC_ERROR"
		errR.Code = 1033
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
		errR.Type = "USER_NOT_FOUND"
		errR.Code = 1040
		return resR, errR
	}
	// DB result rows get fields
	userFields := dataDB["rows"].GetListValue().GetValues()[0].GetStructValue().GetFields()

	// Check avatar existence
	userID := int64(userFields["id"].GetNumberValue())
	avatar, avatarErr := AvatarExists(userID)
	if avatarErr == false {
		userFields["avatar"] = structpb.NewStringValue(avatar)
	} else {
		userFields["avatar"] = structpb.NewStringValue("")
	}

	// Success - Return Profile
	resR.Type = "getProfile"
	resR.Data = userFields
	return resR, errR
}

func UpdateProfile(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
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

	// Check Display Name
	if val, exists := data["display_name"]; exists {
		if val == "" {
			errR.Type = "display_name_is_empty"
			errR.Code = 207
			return resR, errR
		}
	} else {
		errR.Type = "display_name_expected"
		errR.Code = 206
		return resR, errR
	}
	// Check First Name
	if val, exists := data["first_name"]; exists {
		if val == "" {
			data["first_name"] = "-"
		}
	} else {
		data["first_name"] = "-"
	}
	// Check Last Name
	if val, exists := data["last_name"]; exists {
		if val == "" {
			data["last_name"] = "-"
		}
	} else {
		data["last_name"] = "-"
	}
	// Check Steam ID
	if val, exists := data["steam_id"]; exists {
		if val == "" {
			data["steam_id"] = "-"
		}
	} else {
		data["steam_id"] = "-"
	}
	// Check Google ID
	if val, exists := data["google_id"]; exists {
		if val == "" {
			data["google_id"] = "-"
		}
	} else {
		data["google_id"] = "-"
	}
	// Check Discord ID
	if val, exists := data["discord_id"]; exists {
		if val == "" {
			data["discord_id"] = "-"
		}
	} else {
		data["discord_id"] = "-"
	}

	// Sanitize and build query
	query := fmt.Sprintf(
		`UPDATE users SET 
                 display_name='%s',
                 first_name='%s',
                 last_name='%s',
                 steam_id='%s',
                 google_id='%s',
                 discord_id='%s'
             WHERE email='%s' LIMIT 1`,
		data["display_name"].(string),
		data["first_name"].(string),
		data["last_name"].(string),
		data["steam_id"].(string),
		data["google_id"].(string),
		data["discord_id"].(string),
		strings.ToLower(email),
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "PROFILE_GRPC_ERROR"
		errR.Code = 1033
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}

	// Extract gRPC struct
	dataDB := res.Data.GetFields()

	// DB result rows count
	exist := dataDB["rows_affected"].GetNumberValue()
	if exist == 0 {
		errR.Type = "USER_NOT_FOUND"
		errR.Code = 1035
		return resR, errR
	}

	// Success - Return Profile
	resR.Type = "updateProfile"
	resR.Data = ""
	return resR, errR
}

func GetMetadata(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Get Email
	email, ok := GetUserEmail(data)
	if !ok {
		errR.Type = "INVALID_CREDENTIALS"
		errR.Code = 208
		return resR, errR
	}

	// Check App Key
	if val, exists := data["appBin"]; exists {
		if val == "" {
			errR.Type = "BIN_EMPTY"
			errR.Code = 1050
			return resR, errR
		}
	} else {
		errR.Type = "BIN_EXPECTED"
		errR.Code = 1051
		return resR, errR
	}

	// Sanitize and build query
	query := fmt.Sprintf(
		`SELECT m.* 
				FROM users_meta m 
				LEFT JOIN  users u ON u.id=m.user_id 
				WHERE u.email='%s' 
				AND (m.appBin='%s' OR m.type IN ('public','protected') ) `,
		strings.ToLower(email),
		strings.ToLower(data["appBin"].(string)),
	)

	// Check App Key
	if targetType, exists := data["type"]; exists {
		if targetType != "" {
			query = fmt.Sprintf(
				`SELECT m.* 
				FROM users_meta m 
				LEFT JOIN  users u ON u.id=m.user_id 
				WHERE u.email='%s' 
				AND m.appBin='%s'
				AND m.type='%s' `,
				strings.ToLower(email),
				strings.ToLower(data["appBin"].(string)),
				targetType.(string),
			)
		}
	}

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "PROFILE_GRPC_ERROR"
		errR.Code = 1033
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
		errR.Type = "METADATA_NOT_FOUND"
		errR.Code = 1041
		return resR, errR
	}
	// DB result rows get fields
	userFields := dataDB["rows"].GetListValue().GetValues()

	// Success - Return Metadata
	resR.Type = "getMetadata"
	resR.Data = userFields
	return resR, errR
}

func UpdateMetadata(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Get Email
	email, ok := GetUserEmail(data)
	if !ok {
		errR.Type = "INVALID_CREDENTIALS"
		errR.Code = 208
		return resR, errR
	}
	userID, OK := GetUserId(email)
	if !OK {
		errR.Type = "INVALID_CREDENTIALS"
		errR.Code = 208
		return resR, errR
	}

	// Check App Key
	if val, exists := data["appBin"]; exists {
		if val == "" {
			errR.Type = "BIN_EMPTY"
			errR.Code = 1050
			return resR, errR
		}
	} else {
		errR.Type = "BIN_EXPECTED"
		errR.Code = 1051
		return resR, errR
	}

	// Check Type
	if val, exists := data["type"]; exists {
		if val == "" {
			errR.Type = "TYPE_EMPTY"
			errR.Code = 1052
			return resR, errR
		}
	} else {
		errR.Type = "TYPE_EXPECTED"
		errR.Code = 1053
		return resR, errR
	}

	// Check Value
	if _, exists := data["value"]; !exists {
		errR.Type = "VALUE_EXPECTED"
		errR.Code = 1055
		return resR, errR
	}

	switch data["type"] {
	case "public", "private", "protected":
		// continue
	default:
		errR.Type = "TYPE_NOT_VALID"
		errR.Code = 1054
		return resR, errR
	}

	// Sanitize and build query
	query := fmt.Sprintf(
		`REPLACE INTO users_meta (user_id, appBin, type, value) VALUES (%d, '%s', '%s', '%s');`,
		userID,
		data["appBin"].(string),
		data["type"].(string),
		data["value"].(string),
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "PROFILE_GRPC_ERROR"
		errR.Code = 1033
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}

	// Extract gRPC struct
	dataDB := res.Data.GetFields()

	// DB result rows count
	exist := dataDB["inserted_id"].GetNumberValue()
	if exist == 0 {
		errR.Type = "DB_DATA"
		errR.Code = 1070
		return resR, errR
	}

	// Success - Return Profile
	resR.Type = "updateMetadata"
	resR.Data = ""
	return resR, errR
}
