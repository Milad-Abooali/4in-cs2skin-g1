package handlers

import (
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/actions"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/grpcclient"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/validate"
	"os"
	"strconv"
	"strings"
)

func ValidateXKey(data map[string]interface{}) (string, error) {
	val, exists := data["X_KEY"]
	if !exists {
		return "", fmt.Errorf("X_KEY_EXPECTED:2001")
	}
	keyStr, ok := val.(string)
	if !ok || keyStr == "" {
		return "", fmt.Errorf("X_KEY_EMPTY:2001")
	}
	if keyStr != os.Getenv("X_KEY") {
		return "", fmt.Errorf("X_KEY_INVALID:2001")
	}
	return keyStr, nil
}

func XGetJWT(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Admin Key
	_, err := ValidateXKey(data)
	if err != nil {
		errParts := strings.Split(err.Error(), ":")
		errR.Type = errParts[0]
		errR.Code, _ = strconv.Atoi(errParts[1])
		return resR, errR
	}

	// Get Email
	email, ok := GetUserEmail(data)
	if !ok {
		errR.Type = "TOKEN_NOT_FOUND"
		errR.Code = 1032
		return resR, errR
	}

	// Get User ID
	userID, ok := GetUserId(email)
	if !ok {
		errR.Type = "TOKEN_NOT_FOUND"
		errR.Code = 1032
		return resR, errR
	}

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

	// Success - Return User ID
	resR.Type = "xGetJWT"
	resR.Data = map[string]interface{}{
		"profile":  profile,
		"metadata": metadata,
	}
	return resR, errR
}

func XGetUser(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Admin Key
	_, err := ValidateXKey(data)
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
	resR.Type = "xGetUser"
	resR.Data = map[string]interface{}{
		"profile":  profile,
		"metadata": metadata,
	}
	return resR, errR
}

func XAddTransactions(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Check Admin Key
	_, err := ValidateXKey(data)
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
		CreatedBy:    "UM::XAddTransactions",
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
	resR.Type = "xAddTransaction"
	resR.Data = map[string]interface{}{
		"transactionID": insertedID,
		"newBalance":    balanceAfter,
	}
	return resR, errR
}
