package handlers

import (
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/grpcclient"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"log"
	"strconv"
	"strings"
)

func GetBalance(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
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

	// Sanitize and build query
	query := fmt.Sprintf(
		`SELECT balance FROM users WHERE email='%s' LIMIT 1`,
		strings.ToLower(email),
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "CREDIT_GRPC_ERROR"
		errR.Code = 1063
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
		errR.Type = "DB_DATA"
		errR.Code = 1070
		return resR, errR
	}
	// DB result rows get fields
	userFields := dataDB["rows"].GetListValue().GetValues()[0].GetStructValue().GetFields()

	// Success - Return Profile
	resR.Type = "getBalance"
	resR.Data = userFields["balance"].GetStringValue()
	return resR, errR
}

func AddRequest(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
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

	// Get User ID
	userID, OK := GetUserId(email)
	if !OK {
		errR.Type = "INVALID_CREDENTIALS"
		errR.Code = 208
		return resR, errR
	}

	// Check side
	side, exists := data["side"]
	if !exists {
		errR.Type = "SIDE_EXPECTED"
		errR.Code = 1057
		return resR, errR
	}
	if side == "" || side == 0 {
		errR.Type = "SIDE_EMPTY"
		errR.Code = 1056
		return resR, errR
	}
	if side != "deposit" && side != "withdrawal" {
		errR.Type = "SIDE_INVALID"
		errR.Code = 1058
		return resR, errR
	}

	// Check amount
	amount, exists := data["amount"]
	if !exists {
		errR.Type = "AMOUNT_EXPECTED"
		errR.Code = 1065
		return resR, errR
	}
	if amount == 0 {
		errR.Type = "AMOUNT_EMPTY"
		errR.Code = 1064
		return resR, errR
	}

	// Check ref_id
	if _, exists := data["ref_id"]; !exists {
		data["ref_id"] = ""
	}

	// Check target
	target, exists := data["target"]
	if !exists {
		data["target"] = ""
	}
	if side == "withdrawal" {
		if !exists {
			errR.Type = "TARGET_EXPECTED"
			errR.Code = 1060
			return resR, errR
		}
		if target == "" {
			errR.Type = "TARGET_EMPTY"
			errR.Code = 1059
			return resR, errR
		}
	}

	// Check method
	method, exists := data["method"]
	if !exists {
		errR.Type = "METHOD_EXPECTED"
		errR.Code = 1067
		return resR, errR
	}
	if method == "" {
		errR.Type = "METHOD_EMPTY"
		errR.Code = 1066
		return resR, errR
	}
	if method != "skin" && method != "crypto" && method != "fireblock" {
		errR.Type = "METHOD_INVALID"
		errR.Code = 1068
		return resR, errR
	}

	// Sanitize and build query
	query := fmt.Sprintf(
		`INSERT INTO credit_requests (user_id, side, method, amount, status, ref_id, target) 
				VALUES (%d, '%s', '%s', %.2f, 'pending', '%s', '%s')`,
		userID,
		data["side"].(string),
		data["method"].(string),
		data["amount"],
		data["ref_id"].(string),
		data["target"].(string),
	)

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	log.Println(res)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "CREDIT_GRPC_ERROR"
		errR.Code = 1063
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
	resR.Type = "addRequest"
	resR.Data = exist
	return resR, errR
}

func GetRequest(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
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

	// Get User ID
	userID, OK := GetUserId(email)
	if !OK {
		errR.Type = "INVALID_CREDENTIALS"
		errR.Code = 208
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
	userIDInt, _ := strconv.Atoi(fmt.Sprint(userID))
	countInt, _ := strconv.Atoi(fmt.Sprint(count))
	offsetInt, _ := strconv.Atoi(fmt.Sprint(offset))

	// Check ID
	query := ""
	id, existsID := data["id"]
	if existsID {
		idInt, _ := strconv.Atoi(fmt.Sprint(id))
		query = fmt.Sprintf(
			"SELECT * FROM credit_requests WHERE user_id=%d AND id=%d",
			userIDInt,
			idInt,
		)
	} else {
		query = fmt.Sprintf(
			`SELECT * FROM credit_requests WHERE user_id=%d ORDER BY id DESC LIMIT %d OFFSET %d`,
			userIDInt,
			countInt,
			offsetInt,
		)
	}

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "CREDIT_GRPC_ERROR"
		errR.Code = 1063
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}

	// Extract gRPC struct
	dataDB := res.Data.GetFields()

	// Success - Return Profile
	resR.Type = "getRequests"
	resR.Data = dataDB["rows"]
	return resR, errR
}

func GetTransactions(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
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

	// Get User ID
	userID, OK := GetUserId(email)
	if !OK {
		errR.Type = "INVALID_CREDENTIALS"
		errR.Code = 208
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
	userIDInt, _ := strconv.Atoi(fmt.Sprint(userID))
	countInt, _ := strconv.Atoi(fmt.Sprint(count))
	offsetInt, _ := strconv.Atoi(fmt.Sprint(offset))

	// Check ID
	query := ""
	id, existsID := data["id"]
	if existsID {
		idInt, _ := strconv.Atoi(fmt.Sprint(id))
		query = fmt.Sprintf(
			"SELECT * FROM credit_transactions WHERE user_id=%d AND `type` IN ('req_deposit', 'req_withdrawal') AND id=%d",
			userIDInt,
			idInt,
		)
	} else {
		query = fmt.Sprintf(
			"SELECT * FROM credit_transactions WHERE user_id=%d AND `type` IN ('req_deposit', 'req_withdrawal') ORDER BY id DESC LIMIT %d OFFSET %d",
			userIDInt,
			countInt,
			offsetInt,
		)
	}

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "CREDIT_GRPC_ERROR"
		errR.Code = 1063
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}

	// Extract gRPC struct
	dataDB := res.Data.GetFields()

	// Success - Return Profile
	resR.Type = "getTransactions"
	resR.Data = dataDB["rows"]
	return resR, errR
}

func GetTrades(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
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

	// Get User ID
	userID, OK := GetUserId(email)
	if !OK {
		errR.Type = "INVALID_CREDENTIALS"
		errR.Code = 208
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
	userIDInt, _ := strconv.Atoi(fmt.Sprint(userID))
	countInt, _ := strconv.Atoi(fmt.Sprint(count))
	offsetInt, _ := strconv.Atoi(fmt.Sprint(offset))

	// Check ID
	query := ""
	id, existsID := data["id"]
	if existsID {
		idInt, _ := strconv.Atoi(fmt.Sprint(id))
		query = fmt.Sprintf(
			"SELECT * FROM credit_transactions WHERE user_id=%d AND `type` IN ('market_sell', 'market_buy') AND id=%d",
			userIDInt,
			idInt,
		)
	} else {
		query = fmt.Sprintf(
			"SELECT * FROM credit_transactions WHERE user_id=%d AND `type` IN ('market_sell', 'market_buy') ORDER BY id DESC LIMIT %d OFFSET %d",
			userIDInt,
			countInt,
			offsetInt,
		)
	}

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "CREDIT_GRPC_ERROR"
		errR.Code = 1063
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}

	// Extract gRPC struct
	dataDB := res.Data.GetFields()

	// Success - Return Profile
	resR.Type = "getTrades"
	resR.Data = dataDB["rows"]
	return resR, errR
}

func GetGameHistory(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
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

	// Get User ID
	userID, OK := GetUserId(email)
	if !OK {
		errR.Type = "INVALID_CREDENTIALS"
		errR.Code = 208
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
	userIDInt, _ := strconv.Atoi(fmt.Sprint(userID))
	countInt, _ := strconv.Atoi(fmt.Sprint(count))
	offsetInt, _ := strconv.Atoi(fmt.Sprint(offset))

	// Check ID
	query := ""
	id, existsID := data["id"]
	if existsID {
		idInt, _ := strconv.Atoi(fmt.Sprint(id))
		query = fmt.Sprintf(
			"SELECT * FROM credit_transactions WHERE user_id=%d AND `type` IN ('game_win', 'game_loss') AND id=%d",
			userIDInt,
			idInt,
		)
	} else {
		query = fmt.Sprintf(
			"SELECT * FROM credit_transactions WHERE user_id=%d AND `type` IN ('game_win', 'game_loss') ORDER BY id DESC LIMIT %d OFFSET %d",
			userIDInt,
			countInt,
			offsetInt,
		)
	}

	// gRPC Call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		errR.Type = "CREDIT_GRPC_ERROR"
		errR.Code = 1063
		if res != nil {
			errR.Data = res.Error
		}
		return resR, errR
	}

	// Extract gRPC struct
	dataDB := res.Data.GetFields()

	// Success - Return Profile
	resR.Type = "getGameHistory"
	resR.Data = dataDB["rows"]
	return resR, errR
}
