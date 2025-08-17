package web

import (
	"encoding/json"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/configs"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/handlers"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/ws"
	"log"
	"net/http"
	"os"
)

// dispatchWeb runs a handler and writes either error or success to the HTTP response.
func dispatchWeb(w http.ResponseWriter, fn func(map[string]interface{}) (models.HandlerOK, models.HandlerError), req map[string]interface{}) {
	res, err := fn(req)
	if err.Code > 0 {
		handlers.SendWebError(w, err.Type, err.Code, err.Data)
		return
	}
	handlers.SendWebResponse(w, res.Type, res.Data)
	ws.EmitServer(req, res.Type, res.Data)
}

// All POST routes mapped to handlers
var postRoutes = map[string]func(map[string]interface{}) (models.HandlerOK, models.HandlerError){
	// Auth
	"register":    handlers.Register,
	"socialLogin": handlers.SocialLogin,
	"login":       handlers.Login,
	"logout":      handlers.Logout,

	// Token
	"tokenRenew":    handlers.Renew,
	"tokenExtend":   handlers.Extend,
	"tokenValidate": handlers.Validate,

	// Password
	"passRecovery": handlers.PassRecovery,
	"passReset":    handlers.PassReset,
	"passChange":   handlers.PassChange,

	// Profile / Avatar / Metadata
	"getProfile":     handlers.GetProfile,
	"updateProfile":  handlers.UpdateProfile,
	"updateAvatar":   handlers.UpdateAvatar,
	"clearAvatar":    handlers.ClearAvatar,
	"getMetadata":    handlers.GetMetadata,
	"updateMetadata": handlers.UpdateMetadata,

	// User finance/history
	"getBalance":      handlers.GetBalance,
	"addRequest":      handlers.AddRequest,
	"getRequests":     handlers.GetRequest,
	"getTransactions": handlers.GetTransactions,
	"getTrades":       handlers.GetTrades,
	"getGameHistory":  handlers.GetGameHistory,

	// Admin
	"aGetUsers":        handlers.AGetUsers,
	"aGetUser":         handlers.AGetUser,
	"aLoginAsUser":     handlers.ALoginAsUser,
	"aUpdateUser":      handlers.AUpdateUser,
	"aSetPassword":     handlers.ASetPassword,
	"aDeleteUser":      handlers.ADeleteUser,
	"aGetRequests":     handlers.AGetRequests,
	"aUpdateRequest":   handlers.AUpdateRequest,
	"aAddTransaction":  handlers.AAddTransactions,
	"aGetTransactions": handlers.AGetTransactions,

	// External Layer
	"xGetJWT":         handlers.XGetJWT,
	"xGetUser":        handlers.XGetUser,
	"xAddTransaction": handlers.XAddTransactions,
}

func HandleHTTP(w http.ResponseWriter, r *http.Request) {
	// App token validation
	token := r.Header.Get("Authorization")
	if token != "Bearer "+os.Getenv("APP_TOKEN") {
		handlers.SendWebError(w, "INVALID_APP_TOKEN", 10001)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Parse envelope
	var req models.Payload
	var msg models.Request

	if err := json.NewDecoder(r.Body).Decode(&req.Payload); err != nil {
		handlers.SendWebError(w, "INVALID_JSON_BODY", 1002)
		return
	}
	if err := json.Unmarshal(req.Payload, &msg); err != nil {
		handlers.SendWebError(w, "INVALID_JSON_BODY", 1002)
		return
	}

	reqData, ok := msg.Data.(map[string]interface{})
	if !ok {
		handlers.SendWebError(w, "INVALID_DATA_FIELD_TYPE", 1003)
		return
	}

	if configs.Debug {
		log.Println("HTTP Req:", msg.Type)
	}

	switch r.Method {
	case http.MethodPost:
		if fn, ok := postRoutes[msg.Type]; ok {
			dispatchWeb(w, fn, reqData)
			return
		}
		handlers.SendWebError(w, "UNKNOWN_ROUTE", 1010)

	case http.MethodPut:
		handlers.SendWebError(w, "METHOD_NOT_ALLOWED", 1004)

	case http.MethodDelete:
		handlers.SendWebError(w, "METHOD_NOT_ALLOWED", 1004)

	default:
		handlers.SendWebError(w, "METHOD_NOT_ALLOWED", 1004)
	}
}
