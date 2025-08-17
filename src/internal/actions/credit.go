package actions

import (
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/grpcclient"
	"strconv"
	"strings"
)

// StandardError is a unified error type to map codes & keys for frontend processing.
type StandardError struct {
	Code   int
	Key    string
	Detail string
}

func (e *StandardError) Error() string {
	return fmt.Sprintf("%s (%d): %s", e.Key, e.Code, e.Detail)
}

type AddTx struct {
	UserID       int64
	Type         string
	Amount       float64
	CreatedBy    string
	RefID        string
	TxRef        string
	BalanceAfter *float64
	Description  string
}

func AddTransaction(in AddTx) (int64, *StandardError) {

	// Build query dynamically based on provided fields
	cols := []string{"user_id", "type", "amount", "created_by"}
	vals := []string{
		fmt.Sprintf("%d", in.UserID),
		q(in.Type),
		fmt.Sprintf("%.2f", in.Amount),
		q(in.CreatedBy),
	}

	if in.TxRef != "" {
		cols = append(cols, "tx_ref")
		vals = append(vals, q(in.TxRef))
	}
	if in.RefID != "" {
		cols = append(cols, "ref_id")
		vals = append(vals, q(in.RefID))
	}
	if in.BalanceAfter != nil {
		cols = append(cols, "balance_after")
		vals = append(vals, fmt.Sprintf("%.2f", *in.BalanceAfter))
	}
	if in.Description != "" {
		cols = append(cols, "description")
		vals = append(vals, q(in.Description))
	}

	query := fmt.Sprintf(
		"INSERT INTO credit_transactions (%s) VALUES (%s)",
		strings.Join(cols, ", "),
		strings.Join(vals, ", "),
	)

	// gRPC call
	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		detail := ""
		if res != nil && res.Error != "" {
			detail = res.Error
		}
		return 0, &StandardError{
			Code:   1063,
			Key:    "CREDIT_GRPC_ERROR",
			Detail: detail,
		}
	}

	data := res.Data.GetFields()
	id := int64(data["inserted_id"].GetNumberValue())
	if id == 0 {
		return 0, &StandardError{
			Code:   1070,
			Key:    "DB_DATA",
			Detail: "inserted_id == 0",
		}
	}

	tx := strings.TrimSpace(strings.ToLower(in.Type))
	if tx == "req_withdrawal" || tx == "game_loss" || tx == "market_buy" {
		if in.Amount > 0 {
			in.Amount *= -1
		}
	}
	UpdateUserBalance(in.UserID, in.Amount)

	return id, nil
}

func GetUserBalance(userID int64) (float64, *StandardError) {
	query := fmt.Sprintf(
		"SELECT balance AS balance FROM users WHERE id = %d LIMIT 1",
		userID,
	)

	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		detail := ""
		if res != nil && res.Error != "" {
			detail = res.Error
		}
		return 0, &StandardError{
			Code:   1063,
			Key:    "USER_GRPC_ERROR",
			Detail: detail,
		}
	}

	// --- safely read rows ---
	fields := res.Data.GetFields()
	rows := fields["rows"].GetListValue().GetValues()
	if len(rows) == 0 {
		// هیچ کاربری با این id پیدا نشد
		return 0, &StandardError{
			Code:   1040,
			Key:    "USER_NOT_FOUND",
			Detail: fmt.Sprintf("user id %d not found", userID),
		}
	}

	// --- first row ---
	userFields := rows[0].GetStructValue().GetFields()

	// بالانس می‌تواند string یا number باشد؛ هر دو را ساپورت می‌کنیم
	val := userFields["balance"]
	var bal float64
	switch {
	case val.GetStringValue() != "":
		parsed, perr := strconv.ParseFloat(val.GetStringValue(), 64)
		if perr != nil {
			return 0, &StandardError{Code: 1070, Key: "DB_DATA", Detail: "invalid balance format (string)"}
		}
		bal = parsed
	case val.GetNumberValue() != 0 || // اگر صفر است ممکن است واقعاً بالانس صفر باشد؛ پس حالت صفر را هم هندل کنیم
		(val.GetNumberValue() == 0 && val.GetStringValue() == ""):
		// در protobuf، اگر مقدار number باشد، از GetNumberValue می‌آید (حتی اگر صفر باشد)
		bal = val.GetNumberValue()
	default:
		return 0, &StandardError{Code: 1070, Key: "DB_DATA", Detail: "invalid balance format (unknown type)"}
	}

	return bal, nil
}

func GetRequest(requestID int64) (map[string]any, *StandardError) {
	query := fmt.Sprintf("SELECT * FROM credit_requests WHERE id = %d LIMIT 1", requestID)

	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		detail := ""
		if res != nil && res.Error != "" {
			detail = res.Error
		}
		return nil, &StandardError{Code: 1063, Key: "USER_GRPC_ERROR", Detail: detail}
	}

	// rows[0] -> struct -> fields: map[column]*structpb.Value
	row := res.Data.GetFields()["rows"].GetListValue().GetValues()[0].GetStructValue().GetFields()

	out := make(map[string]any, len(row))
	for k, v := range row {
		// v: *structpb.Value  → Go native with AsInterface()
		// number_value -> float64, string_value -> string, bool -> bool, null -> nil
		out[k] = v.AsInterface()
	}

	return out, nil
}

func UpdateUserBalance(userID int64, delta float64) (float64, *StandardError) {
	query := fmt.Sprintf(
		"UPDATE users SET balance = balance + %.2f WHERE id = %d",
		delta,
		userID,
	)

	res, err := grpcclient.SendQuery(query)
	if err != nil || res == nil || res.Status != "ok" {
		detail := ""
		if res != nil && res.Error != "" {
			detail = res.Error
		}
		return 0, &StandardError{
			Code:   1063,
			Key:    "USER_GRPC_ERROR",
			Detail: detail,
		}
	}

	newBal, serr := GetUserBalance(userID)
	if serr != nil {
		return 0, serr
	}

	return newBal, nil
}

// Basic SQL quote/escape
func q(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
