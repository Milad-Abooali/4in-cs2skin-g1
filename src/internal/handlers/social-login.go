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

func SocialLogin(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Step 1: Check required fields
	provider, ok := data["provider"].(string)
	if !ok || provider == "" {
		errR.Type = "provider_expected"
		errR.Code = 1016
		return resR, errR
	}
	token, ok := data["token"].(string)
	if !ok || token == "" {
		errR.Type = "token_expected"
		errR.Code = 1017
		return resR, errR
	}

	// Step 2: Get expire_in (optional)
	expireIn := 60
	if val, ok := data["expire_in"].(float64); ok && val > 0 {
		expireIn = int(val)
	}

	// Step 3: Handle each provider
	switch provider {

	case "steam":
		prefix := "https://steamcommunity.com/openid/id/"
		if !strings.HasPrefix(token, prefix) {
			errR.Type = "invalid_steam_token"
			errR.Code = 1020
			return resR, errR
		}
		socialID := strings.TrimPrefix(token, prefix)

		checkQuery := fmt.Sprintf(`SELECT id, display_name FROM users WHERE steam_id = "%s" LIMIT 1`, socialID)
		checkRes, err := grpcclient.SendQuery(checkQuery)
		if err == nil && checkRes != nil && checkRes.Status == "ok" {
			dataDB := checkRes.Data.GetFields()
			if dataDB["count"].GetNumberValue() > 0 {
				userFields := dataDB["rows"].GetListValue().GetValues()[0].GetStructValue().GetFields()
				displayName := userFields["display_name"].GetStringValue()
				userID := int(userFields["id"].GetNumberValue())
				return generateSocialLoginResponse(socialID, displayName, userID, expireIn)
			}
		}

		displayName := "SteamUser"
		if meta, ok := data["meta"].(map[string]interface{}); ok {
			if dn, ok := meta["display_name"].(string); ok && dn != "" {
				displayName = dn
			}
		}

		insertQuery := fmt.Sprintf(`INSERT INTO users (steam_id, display_name) VALUES ("%s", "%s")`, socialID, displayName)
		insertRes, err := grpcclient.SendQuery(insertQuery)
		if err != nil || insertRes == nil || insertRes.Status != "ok" {
			errR.Type = "social_register_failed"
			errR.Code = 1021
			if insertRes != nil {
				errR.Data = insertRes.Error
			}
			return resR, errR
		}
		userID := int(insertRes.Data.GetFields()["inserted_id"].GetNumberValue())
		if userID < 1 {
			errR.Type = "social_register_db_err"
			errR.Code = 1022
			return resR, errR
		}
		return generateSocialLoginResponse(socialID, displayName, userID, expireIn)

	case "google":
		socialID, displayName, err := utils.VerifyGoogleToken(token)
		if err != nil || socialID == "" {
			errR.Type = "invalid_google_token"
			errR.Code = 1024
			return resR, errR
		}

		checkQuery := fmt.Sprintf(`SELECT id, display_name FROM users WHERE google_id = "%s" LIMIT 1`, socialID)
		checkRes, err := grpcclient.SendQuery(checkQuery)
		if err == nil && checkRes != nil && checkRes.Status == "ok" {
			dataDB := checkRes.Data.GetFields()
			if dataDB["count"].GetNumberValue() > 0 {
				userFields := dataDB["rows"].GetListValue().GetValues()[0].GetStructValue().GetFields()
				displayName := userFields["display_name"].GetStringValue()
				userID := int(userFields["id"].GetNumberValue())
				return generateSocialLoginResponse(socialID, displayName, userID, expireIn)
			}
		}

		insertQuery := fmt.Sprintf(`INSERT INTO users (google_id, display_name) VALUES ("%s", "%s")`, socialID, displayName)
		insertRes, err := grpcclient.SendQuery(insertQuery)
		if err != nil || insertRes == nil || insertRes.Status != "ok" {
			errR.Type = "social_register_failed"
			errR.Code = 1021
			return resR, errR
		}
		userID := int(insertRes.Data.GetFields()["inserted_id"].GetNumberValue())
		return generateSocialLoginResponse(socialID, displayName, userID, expireIn)

	case "discord":
		socialID, displayName, err := utils.VerifyDiscordToken(token)
		if err != nil || socialID == "" {
			errR.Type = "invalid_discord_token"
			errR.Code = 1025
			return resR, errR
		}

		checkQuery := fmt.Sprintf(`SELECT id, display_name FROM users WHERE discord_id = "%s" LIMIT 1`, socialID)
		checkRes, err := grpcclient.SendQuery(checkQuery)
		if err == nil && checkRes != nil && checkRes.Status == "ok" {
			dataDB := checkRes.Data.GetFields()
			if dataDB["count"].GetNumberValue() > 0 {
				userFields := dataDB["rows"].GetListValue().GetValues()[0].GetStructValue().GetFields()
				displayName := userFields["display_name"].GetStringValue()
				userID := int(userFields["id"].GetNumberValue())
				return generateSocialLoginResponse(socialID, displayName, userID, expireIn)
			}
		}

		insertQuery := fmt.Sprintf(`INSERT INTO users (discord_id, display_name) VALUES ("%s", "%s")`, socialID, displayName)
		insertRes, err := grpcclient.SendQuery(insertQuery)
		if err != nil || insertRes == nil || insertRes.Status != "ok" {
			errR.Type = "social_register_failed"
			errR.Code = 1021
			return resR, errR
		}
		userID := int(insertRes.Data.GetFields()["inserted_id"].GetNumberValue())
		return generateSocialLoginResponse(socialID, displayName, userID, expireIn)

	default:
		errR.Type = "unsupported_provider"
		errR.Code = 1023
		return resR, errR
	}
}

func generateSocialLoginResponse(socialID string, displayName string, userID int, expireIn int) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	duration := time.Duration(expireIn) * time.Minute
	token, err := utils.GenerateJWT(socialID, duration)
	if err != nil {
		errR.Type = "token_generation_failed"
		errR.Code = 209
		return resR, errR
	}

	memory.SetToken(token, socialID, duration)

	resR.Type = "login"
	resR.Data = map[string]interface{}{
		"display_name": displayName,
		"token":        token,
		"user_id":      userID,
	}
	return resR, errR
}
