package handlers

import (
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/actions"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/grpcclient"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/memory"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/validate"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/utils"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func ValidateAdminKey(data map[string]interface{}) (string, error) {
	val, exists := data["adminKey"]
	if !exists {
		return "", fmt.Errorf("ADMIN_KEY_EXPECTED:2001")
	}
	keyStr, ok := val.(string)
	if !ok || keyStr == "" {
		return "", fmt.Errorf("ADMIN_KEY_EMPTY:2001")
	}
	if keyStr != os.Getenv("ADMIN_KEY") {
		return "", fmt.Errorf("ADMIN_KEY_INVALID:2001")
	}
	return keyStr, nil
}

func AGetUsers(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Admin Key
	_, err := ValidateAdminKey(data)
	if err != nil {
		errParts := strings.Split(err.Error(), ":")
		errR.Type = errParts[0]
		errR.Code, _ = strconv.Atoi(errParts[1])
		return resR, errR
	}

	// Check Count
	count, exists := data["count"]
	if !exists {
		count = 25
	}

	// Check Offset
	offset, exists := data["offset"]
	if !exists {
		offset = 0
	}

	// Sanitize and build query
	countInt, _ := strconv.Atoi(fmt.Sprint(count))
	offsetInt, _ := strconv.Atoi(fmt.Sprint(offset))

	// Check ID
	query := fmt.Sprintf(
		"SELECT *,'******' as `password` FROM users ORDER BY id DESC LIMIT %d OFFSET %d",
		countInt,
		offsetInt,
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "ADMIN_GRPC_ERROR"
		errR.Code = 2002
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}
	// Extract gRPC struct
	dataDB := res.Data.GetFields()
	// DB result rows count
	// exist := dataDB["count"].GetNumberValue()

	// Success - Return Profile
	resR.Type = "aGetUsers"
	resR.Data = dataDB["rows"]
	return resR, errR
}

func AGetUser(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Admin Key
	_, err := ValidateAdminKey(data)
	if err != nil {
		errParts := strings.Split(err.Error(), ":")
		errR.Type = errParts[0]
		errR.Code, _ = strconv.Atoi(errParts[1])
		return resR, errR
	}

	// Check User ID
	val, exists := data["userID"]
	if !exists {
		errR.Type = "userID_expected"
		errR.Code = 206
		return resR, errR
	}
	if val == "" {
		errR.Type = "userID_is_empty"
		errR.Code = 207
		return resR, errR
	}
	userID := int(val.(float64))

	// Query :: Profile
	query := fmt.Sprintf(
		"SELECT *,'******' as `password` FROM users WHERE id=%d LIMIT 1",
		userID,
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "ADMIN_GRPC_ERROR"
		errR.Code = 2002
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}
	// Extract gRPC struct
	dataDB := res.Data.GetFields()
	// DB result rows count
	exist := dataDB["count"].GetNumberValue()
	if exist != 0 {
	}
	profile := dataDB["rows"].GetListValue().GetValues()[0].GetStructValue().GetFields()

	// Query :: Metadata
	query = fmt.Sprintf(
		`SELECT * FROM users_meta WHERE user_id=%d `,
		userID,
	)
	// gRPC Call
	res, err = grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "ADMIN_GRPC_ERROR"
		errR.Code = 2002
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}
	// Extract gRPC struct
	dataDB = res.Data.GetFields()
	// DB result rows count
	exist = dataDB["count"].GetNumberValue()
	if exist == 0 {
	}
	metadata := dataDB["rows"].GetListValue().GetValues()

	// Success - Return Profile
	resR.Type = "aGetUser"
	resR.Data = map[string]interface{}{
		"profile":  profile,
		"metadata": metadata,
	}
	return resR, errR
}

func ALoginAsUser(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Admin Key
	_, err := ValidateAdminKey(data)
	if err != nil {
		errParts := strings.Split(err.Error(), ":")
		errR.Type = errParts[0]
		errR.Code, _ = strconv.Atoi(errParts[1])
		return resR, errR
	}

	// Check User ID
	val, exists := data["userID"]
	if !exists {
		errR.Type = "userID_expected"
		errR.Code = 206
		return resR, errR
	}
	if val == "" {
		errR.Type = "userID_is_empty"
		errR.Code = 207
		return resR, errR
	}
	userID := int(val.(float64))

	// Query :: Profile
	query := fmt.Sprintf(
		"SELECT display_name, email FROM users WHERE id=%d LIMIT 1",
		userID,
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "ADMIN_GRPC_ERROR"
		errR.Code = 2002
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}
	// Extract gRPC struct
	dataDB := res.Data.GetFields()
	userFields := dataDB["rows"].GetListValue().GetValues()[0].GetStructValue().GetFields()
	displayName := ""
	if val, ok := userFields["display_name"]; ok {
		displayName = val.GetStringValue()
	}
	email := ""
	if val, ok := userFields["email"]; ok {
		email = val.GetStringValue()
	} else {
		errR.Type = "USER_NOT_FOUND"
		errR.Code = 1040
		return resR, errR
	}

	// Generate JWT token
	expireIn := 30
	duration := time.Duration(expireIn) * time.Minute
	token, err := utils.GenerateJWT(strings.ToLower(email), duration)
	if err != nil {
		errR.Type = "TOKEN_GENERATION_FAILED"
		errR.Code = 209
		return resR, errR
	}

	// Save JWT in memory for session tracking
	memory.SetToken(token, strings.ToLower(email), duration)

	// Success
	resR.Type = "aLoginAsUser"
	resR.Data = map[string]interface{}{
		"email":        email,
		"display_name": displayName,
		"token":        token,
	}
	return resR, errR
}

func AUpdateUser(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Admin Key
	_, err := ValidateAdminKey(data)
	if err != nil {
		errParts := strings.Split(err.Error(), ":")
		errR.Type = errParts[0]
		errR.Code, _ = strconv.Atoi(errParts[1])
		return resR, errR
	}

	// Check User ID
	val, exists := data["userID"]
	if !exists {
		errR.Type = "userID_expected"
		errR.Code = 206
		return resR, errR
	}
	if val == "" {
		errR.Type = "userID_is_empty"
		errR.Code = 207
		return resR, errR
	}
	userID := int(val.(float64))

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

	// Query
	query := fmt.Sprintf(
		`UPDATE users SET 
                 display_name='%s',
                 first_name='%s',
                 last_name='%s',
                 steam_id='%s',
                 google_id='%s',
                 discord_id='%s'
             WHERE id=%d LIMIT 1`,
		data["display_name"].(string),
		data["first_name"].(string),
		data["last_name"].(string),
		data["steam_id"].(string),
		data["google_id"].(string),
		data["discord_id"].(string),
		userID,
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "ADMIN_GRPC_ERROR"
		errR.Code = 2002
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
	resR.Type = "aUpdateUser"
	resR.Data = ""
	return resR, errR
}

func ASetPassword(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Admin Key
	_, err := ValidateAdminKey(data)
	if err != nil {
		errParts := strings.Split(err.Error(), ":")
		errR.Type = errParts[0]
		errR.Code, _ = strconv.Atoi(errParts[1])
		return resR, errR
	}

	// Check User ID
	val, exists := data["userID"]
	if !exists {
		errR.Type = "userID_expected"
		errR.Code = 206
		return resR, errR
	}
	if val == "" {
		errR.Type = "userID_is_empty"
		errR.Code = 207
		return resR, errR
	}
	userID := int(val.(float64))

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

	// Sanitize and build query
	query := fmt.Sprintf(
		"UPDATE users SET `password`=MD5('%s') WHERE id=%d ",
		data["pass"].(string),
		userID,
	)
	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "login_grpc_err"
		errR.Code = 205
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
		errR.Type = "password_not_changed"
		errR.Code = 213
		return resR, errR
	}

	// Success
	resR.Type = "aSetPassword"
	resR.Data = ""
	return resR, errR
}

func ADeleteUser(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Admin Key
	_, err := ValidateAdminKey(data)
	if err != nil {
		errParts := strings.Split(err.Error(), ":")
		errR.Type = errParts[0]
		errR.Code, _ = strconv.Atoi(errParts[1])
		return resR, errR
	}

	// Check User ID
	val, exists := data["userID"]
	if !exists {
		errR.Type = "userID_expected"
		errR.Code = 206
		return resR, errR
	}
	if val == "" {
		errR.Type = "userID_is_empty"
		errR.Code = 207
		return resR, errR
	}
	userID := int(val.(float64))

	// Sanitize and build query
	query := fmt.Sprintf(
		"DELETE FROM users WHERE id=%d ",
		userID,
	)
	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "ADMIN_GRPC_ERROR"
		errR.Code = 2002
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
		errR.Code = 1040
		return resR, errR
	}

	// Success
	resR.Type = "aDeleteUser"
	resR.Data = ""
	return resR, errR
}

func AGetRequests(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Admin Key
	_, err := ValidateAdminKey(data)
	if err != nil {
		errParts := strings.Split(err.Error(), ":")
		errR.Type = errParts[0]
		errR.Code, _ = strconv.Atoi(errParts[1])
		return resR, errR
	}

	// Check Count
	count, exists := data["count"]
	if !exists {
		count = 25
	}

	// Check Offset
	offset, exists := data["offset"]
	if !exists {
		offset = 0
	}

	// Sanitize and build query
	countInt, _ := strconv.Atoi(fmt.Sprint(count))
	offsetInt, _ := strconv.Atoi(fmt.Sprint(offset))

	// Check ID
	query := fmt.Sprintf(
		"SELECT * FROM credit_requests ORDER BY id DESC LIMIT %d OFFSET %d",
		countInt,
		offsetInt,
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "ADMIN_GRPC_ERROR"
		errR.Code = 2002
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}
	// Extract gRPC struct
	dataDB := res.Data.GetFields()
	// DB result rows count
	// exist := dataDB["count"].GetNumberValue()

	// Success - Return Profile
	resR.Type = "aGetRequests"
	resR.Data = dataDB["rows"]
	return resR, errR
}

func AUpdateRequest(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Admin Key
	_, err := ValidateAdminKey(data)
	if err != nil {
		errParts := strings.Split(err.Error(), ":")
		errR.Type = errParts[0]
		errR.Code, _ = strconv.Atoi(errParts[1])
		return resR, errR
	}

	// Check Req ID
	val, exists := data["request_id"]
	if !exists {
		errR.Type = "REFERENCE_EXPECTED"
		errR.Code = 2010
		return resR, errR
	}
	if val == "" {
		errR.Type = "REFERENCE_EMPTY"
		errR.Code = 2011
		return resR, errR
	}
	requestId := int(val.(float64))

	// Check Status
	status, exists := data["status"]
	if !exists {
		errR.Type = "STATUS_EXPECTED"
		errR.Code = 2012
		return resR, errR
	}
	if status == "" || status == 0 {
		errR.Type = "STATUS_EMPTY"
		errR.Code = 2013
		return resR, errR
	}
	if status != "completed" && status != "failed" && status != "pending" {
		errR.Type = "STATUS_INVALID"
		errR.Code = 2014
		return resR, errR
	}

	// Check ref_id
	if val, exists := data["ref_id"]; exists {
		if val == "" {
			data["ref_id"] = ""
		}
	} else {
		data["ref_id"] = ""
	}

	// Query
	query := fmt.Sprintf(
		`UPDATE credit_requests SET 
                 ref_id='%s',
                 status='%s'
             WHERE id=%d LIMIT 1`,
		data["ref_id"].(string),
		status.(string),
		requestId,
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "ADMIN_GRPC_ERROR"
		errR.Code = 2002
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
		errR.Type = "REQUEST_NOT_UPDATED"
		errR.Code = 1035
		return resR, errR
	}

	if status == "completed" {
		// Get Request
		request, errA := actions.GetRequest(int64(requestId))
		if errA != nil {
			errR.Type = errA.Key
			errR.Code = errA.Code
			resR.Data = errA.Detail
			return resR, errR
		}

		userID := int64(request["user_id"].(float64))
		amount, _ := strconv.ParseFloat(request["amount"].(string), 64)
		reqType := "req_" + request["side"].(string)

		// Get Balance
		balance, errA := actions.GetUserBalance(userID)
		if errA != nil {
			errR.Type = errA.Key
			errR.Code = errA.Code
			resR.Data = errA.Detail
			return resR, errR
		}

		// New Balance
		txType := strings.TrimSpace(strings.ToLower(reqType))
		balanceAfter := balance
		if txType == "req_withdrawal" || txType == "game_loss" || txType == "market_buy" {
			if amount > 0 {
				balanceAfter -= amount
			} else {
				balanceAfter += amount
			}
		} else {
			balanceAfter += amount
		}

		// Add Transaction
		tx := actions.AddTx{
			UserID:       userID,
			Type:         reqType,
			Amount:       amount,
			CreatedBy:    "UM::AUpdateRequest",
			TxRef:        data["ref_id"].(string),
			RefID:        strconv.Itoa(requestId),
			BalanceAfter: &balanceAfter,
			Description:  "Request Completed",
		}
		insertedID, errA := actions.AddTransaction(tx)
		if errA != nil {
			errR.Type = errA.Key
			errR.Code = errA.Code
			resR.Data = errA.Detail
			return resR, errR
		}
		log.Println("Transaction inserted with ID:", insertedID)
	}

	// Success
	resR.Type = "aUpdateRequest"
	resR.Data = ""
	return resR, errR
}

func AGetTransactions(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Admin Key
	_, err := ValidateAdminKey(data)
	if err != nil {
		errParts := strings.Split(err.Error(), ":")
		errR.Type = errParts[0]
		errR.Code, _ = strconv.Atoi(errParts[1])
		return resR, errR
	}

	// Check Count
	count, exists := data["count"]
	if !exists {
		count = 25
	}

	// Check Offset
	offset, exists := data["offset"]
	if !exists {
		offset = 0
	}

	// Sanitize and build query
	countInt, _ := strconv.Atoi(fmt.Sprint(count))
	offsetInt, _ := strconv.Atoi(fmt.Sprint(offset))

	// Query
	query := fmt.Sprintf(
		"SELECT * FROM credit_transactions ORDER BY id DESC LIMIT %d OFFSET %d",
		countInt,
		offsetInt,
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "ADMIN_GRPC_ERROR"
		errR.Code = 2002
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}
	// Extract gRPC struct
	dataDB := res.Data.GetFields()
	// DB result rows count
	// exist := dataDB["count"].GetNumberValue()

	// Success - Return Profile
	resR.Type = "aGetTransactions"
	resR.Data = dataDB["rows"]
	return resR, errR
}

func AAddTransactions(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Admin Key
	_, err := ValidateAdminKey(data)
	if err != nil {
		errParts := strings.Split(err.Error(), ":")
		errR.Type = errParts[0]
		errR.Code, _ = strconv.Atoi(errParts[1])
		return resR, errR
	}

	// Check User ID
	userID, vErr, ok := validate.RequireInt(data, "userID")
	if !ok {
		return resR, vErr
	}

	// Check Type
	allowedTypes := []string{
		"req_deposit", "req_withdrawal", "game_win", "game_loss",
		"referral", "bonus", "upgrade", "market_sell", "market_buy",
	}
	txType, vErr, ok := validate.RequireStringIn(data, "type", allowedTypes)
	if !ok {
		return resR, vErr
	}

	// Check Reference (ref_id)
	refRequired := map[string]bool{
		"req_deposit":    true,
		"req_withdrawal": true,
		"game_win":       true,
		"game_loss":      true,
		"market_sell":    true,
		"market_buy":     true,
		"referral":       true,
		"bonus":          false,
		"upgrade":        false,
	}
	emptyRefID := !refRequired[txType]
	refID, vErr, ok := validate.RequireString(data, "referenceID", emptyRefID)
	if !ok {
		return resR, vErr
	}

	// Check Amount
	amount, vErr, ok := validate.RequireFloat(data, "amount")
	if !ok {
		return resR, vErr
	}

	// Check Balance After
	balance, errA := actions.GetUserBalance(userID)
	if errA != nil {
		errR.Type = "USER_NOT_FOUND"
		errR.Code = 1035
		return resR, errR
	}
	balanceAfter := balance
	if txType == "req_withdrawal" || txType == "game_loss" || txType == "market_buy" {
		if amount > 0 {
			balanceAfter -= amount
		} else {
			balanceAfter += amount
		}
	} else {
		balanceAfter += amount
	}

	// Check Transaction Number
	txRef, vErr, ok := validate.RequireString(data, "txRef", true)
	if !ok {
		return resR, vErr
	}

	// Check Description
	description, vErr, ok := validate.RequireString(data, "description", true)
	if !ok {
		return resR, vErr
	}

	// Add Transaction
	tx := actions.AddTx{
		UserID:       userID,
		Type:         txType,
		Amount:       amount,
		CreatedBy:    "UM::AAddTransactions",
		TxRef:        txRef,
		RefID:        refID,
		BalanceAfter: &balanceAfter,
		Description:  description,
	}
	insertedID, errA := actions.AddTransaction(tx)
	if errA != nil {
		errR.Type = errA.Key
		errR.Code = errA.Code
		resR.Data = errA.Detail
		return resR, errR
	}

	// Success
	resR.Type = "aAddTransaction"
	resR.Data = map[string]interface{}{
		"transactionID": insertedID,
		"newBalance":    balanceAfter,
	}
	return resR, errR
}
