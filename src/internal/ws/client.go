package ws

import (
	"encoding/json"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/configs"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/handlers"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Bind a WS connection to a user if a valid JWT is provided in the request body
func tryBindFromBody(conn *websocket.Conn, reqData map[string]interface{}) {
	var token string
	if v, ok := reqData["token"].(string); ok && v != "" {
		token = v
	} else if v, ok := reqData["jwt"].(string); ok && v != "" {
		token = v
	}
	if token == "" {
		return
	}

	email, ok := handlers.GetUserEmail(reqData)
	if !ok {
		return
	}
	userID, ok := handlers.GetUserId(email)
	if !ok {
		return
	}

	BindUser(conn, userID)
	log.Printf("WS connection bound to userID %d\n", userID)
}

// Executes a handler and sends either success or error response back to client
func dispatch(conn *websocket.Conn, fn func(map[string]interface{}) (models.HandlerOK, models.HandlerError), req map[string]interface{}) {
	res, err := fn(req)
	if err.Code > 0 {
		handlers.SendWSError(conn, err.Type, err.Code, err.Data)
		return
	}
	handlers.SendWSResponse(conn, res.Type, res.Data)
	EmitServer(req, res.Type, res.Data)
}

// All WS routes mapped to handlers
var wsRoutes = map[string]func(*websocket.Conn, map[string]interface{}){
	// Authentication
	"register":    func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.Register, d) },
	"socialLogin": func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.SocialLogin, d) },
	"login":       func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.Login, d) },
	"logout": func(c *websocket.Conn, d map[string]interface{}) {
		dispatch(c, handlers.Logout, d)
		BindUser(c, 0) // Detach connection from user
	},

	// Token
	"tokenRenew":    func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.Renew, d) },
	"tokenExtend":   func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.Extend, d) },
	"tokenValidate": func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.Validate, d) },

	// Password
	"passRecovery": func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.PassRecovery, d) },
	"passReset":    func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.PassReset, d) },
	"passChange":   func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.PassChange, d) },

	// Profile / Avatar / Metadata
	"getProfile":     func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.GetProfile, d) },
	"updateProfile":  func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.UpdateProfile, d) },
	"updateAvatar":   func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.UpdateAvatar, d) },
	"clearAvatar":    func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.ClearAvatar, d) },
	"getMetadata":    func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.GetMetadata, d) },
	"updateMetadata": func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.UpdateMetadata, d) },

	// User finance/history
	"getBalance":      func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.GetBalance, d) },
	"addRequest":      func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.AddRequest, d) },
	"getRequest":      func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.GetRequest, d) },
	"getTransactions": func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.GetTransactions, d) },
	"getTrades":       func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.GetTrades, d) },
	"getGameHistory":  func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.GetGameHistory, d) },

	// Admin Layer
	"aGetUsers":       func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.AGetUsers, d) },
	"aGetUser":        func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.AGetUser, d) },
	"aLoginAsUser":    func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.ALoginAsUser, d) },
	"aUpdateUser":     func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.AUpdateUser, d) },
	"aSetPassword":    func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.ASetPassword, d) },
	"aDeleteUser":     func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.ADeleteUser, d) },
	"aGetRequests":    func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.AGetRequests, d) },
	"aUpdateRequest":  func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.AUpdateRequest, d) },
	"aAddTransaction": func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.AAddTransactions, d) },
	"aGetTransaction": func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.AGetTransactions, d) },

	// External Layer
	"xGetJWT":         func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.XGetJWT, d) },
	"xGetUser":        func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.XGetUser, d) },
	"xAddTransaction": func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.XAddTransactions, d) },

	// Socket
	"sBind": func(c *websocket.Conn, d map[string]interface{}) { dispatch(c, handlers.SBind, d) },
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	RegisterConn(conn)
	defer func() {
		UnregisterConn(conn)
		_ = conn.Close()
	}()

	// App token check
	_, token, err := conn.ReadMessage()
	if err != nil {
		log.Println("WebSocket Read Error:", err)
		return
	}
	if string(token) != os.Getenv("APP_TOKEN") {
		handlers.SendWSError(conn, "INVALID_APP_TOKEN", 10001, "")
		return
	}

	// Handshake
	handlers.SendWSResponse(conn, "handshake", map[string]interface{}{
		"apiVersion": configs.Version,
		"serverTime": time.Now().UTC().Format(time.RFC3339),
	})

	// Main loop
	var msg models.Request
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			handlers.SendWSError(conn, "INVALID_JSON_BODY", 1002, "")
			break
		}
		reqData, ok := msg.Data.(map[string]interface{})
		if !ok {
			handlers.SendWSError(conn, "INVALID_DATA_FIELD_TYPE", 1003, "")
			return
		}
		if configs.Debug {
			log.Println("Web Req:", msg.Type)
		}

		// Auto-bind if JWT provided
		tryBindFromBody(conn, reqData)

		// Special case: bind
		if msg.Type == "bind" {
			tryBindFromBody(conn, reqData)
			handlers.SendWSResponse(conn, "bind.ok", map[string]any{
				"at": time.Now().UTC().Format(time.RFC3339),
			})
			continue
		}

		// Dispatch via map
		if fn, found := wsRoutes[msg.Type]; found {
			fn(conn, reqData)
			continue
		}

		// Unknown route
		handlers.SendWSError(conn, "UNKNOWN_ROUTE", 1010, map[string]any{"type": msg.Type})
	}
}
