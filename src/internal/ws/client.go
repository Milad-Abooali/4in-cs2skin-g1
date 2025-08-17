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
	"sync"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
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

	// --- per-connection write lock (برای جلوگیری از concurrent write) ---
	var writeMu sync.Mutex
	safeSendOK := func(t string, data any) {
		writeMu.Lock()
		defer writeMu.Unlock()
		handlers.SendWSResponse(conn, t, data)
	}
	safeSendErr := func(typ string, code int, data any) {
		writeMu.Lock()
		defer writeMu.Unlock()
		handlers.SendWSError(conn, typ, code, data)
	}

	// --- heartbeat & deadlines ---
	conn.SetReadLimit(1 << 20) // 1MB
	conn.SetReadDeadline(time.Now().Add(75 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(75 * time.Second))
		return nil
	})
	go func(c *websocket.Conn) {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			writeMu.Lock()
			_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := c.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
				writeMu.Unlock()
				log.Println("ping error:", err)
				return
			}
			writeMu.Unlock()
		}
	}(conn)

	// --- App token check ---
	if os.Getenv("DEBUG") != "1" {
		_, token, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket Read Error:", err)
			return
		}
		if string(token) != os.Getenv("APP_TOKEN") {
			safeSendErr("INVALID_APP_TOKEN", 1001, "")
			return
		}
	}

	// --- Handshake ---
	safeSendOK("handshake", map[string]any{
		"apiVersion": configs.Version,
		"serverTime": time.Now().UTC().Format(time.RFC3339),
	})

	// --- dispatch helper برای همین کانکشن با قفل ---
	dispatch := func(fn func(map[string]interface{}) (models.HandlerOK, models.HandlerError), req map[string]interface{}) {
		res, herr := fn(req)
		if herr.Code > 0 {
			safeSendErr(herr.Type, herr.Code, herr.Data)
			return
		}
		safeSendOK(res.Type, res.Data)
		EmitServer(req, res.Type, res.Data)
	}

	// --- routes مخصوص همین کانکشن (closure) ---
	routes := map[string]func(map[string]interface{}){
		// Ping
		"ping": func(d map[string]interface{}) { dispatch(handlers.Ping, d) },

		// Store
		"getBots":  func(d map[string]interface{}) { dispatch(handlers.GetBots, d) },
		"getCases": func(d map[string]interface{}) { dispatch(handlers.GetCases, d) },

		// User Actions
		"newBattle": func(d map[string]interface{}) { dispatch(handlers.NewBattle, d) },
	}

	// --- Main loop ---
	var msg models.Request
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		if err := json.Unmarshal(data, &msg); err != nil {
			safeSendErr("INVALID_JSON_BODY", 1002, "")
			continue
		}

		reqData, ok := msg.Data.(map[string]interface{})
		if !ok {
			safeSendErr("INVALID_DATA_FIELD_TYPE", 1003, "")
			continue
		}

		if configs.Debug {
			log.Println("Web Req:", msg.Type)
		}

		// Special case: bind
		if msg.Type == "bind" {
			safeSendOK("bind.ok", map[string]any{
				"at": time.Now().UTC().Format(time.RFC3339),
			})
			continue
		}

		// Dispatch via routes
		if fn, found := routes[msg.Type]; found {
			fn(reqData)
			continue
		}

		// Unknown route
		safeSendErr("UNKNOWN_ROUTE", 1010, map[string]any{"type": msg.Type})
	}
}
