package ws

import (
	"encoding/json"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/handlers"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type connInfo struct {
	Conn   *websocket.Conn
	UserID int64
	mu     sync.Mutex // serialize writes per-connection
}

var (
	regMu  sync.RWMutex
	byConn = make(map[*websocket.Conn]*connInfo)
	byUser = make(map[int64]map[*websocket.Conn]*connInfo)
)

// RegisterConn should be called right after upgrade
func RegisterConn(c *websocket.Conn) {
	regMu.Lock()
	defer regMu.Unlock()
	if _, ok := byConn[c]; !ok {
		byConn[c] = &connInfo{Conn: c, UserID: 0}
	}
}

// UnregisterConn should be called on close
func UnregisterConn(c *websocket.Conn) {
	regMu.Lock()
	defer regMu.Unlock()
	ci, ok := byConn[c]
	if !ok {
		return
	}
	// remove from user bucket if bound
	if ci.UserID != 0 {
		if set, ok := byUser[ci.UserID]; ok {
			delete(set, c)
			if len(set) == 0 {
				delete(byUser, ci.UserID)
			}
		}
	}
	delete(byConn, c)
}

// BindUser binds an existing connection to a userID (after login)
func BindUser(c *websocket.Conn, userID int64) {
	regMu.Lock()
	defer regMu.Unlock()

	ci, ok := byConn[c]
	if !ok {
		return // not registered
	}
	// remove from previous bucket
	if ci.UserID != 0 {
		if set, ok := byUser[ci.UserID]; ok {
			delete(set, c)
			if len(set) == 0 {
				delete(byUser, ci.UserID)
			}
		}
	}
	// add to new bucket
	ci.UserID = userID
	if userID != 0 {
		if byUser[userID] == nil {
			byUser[userID] = make(map[*websocket.Conn]*connInfo)
		}
		byUser[userID][c] = ci
	}
}

// internal helper: send payload to a list of connections
func emitToTargets(targets []*connInfo, payload any) {
	if len(targets) == 0 {
		return
	}
	// marshal once
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	// write with per-conn lock to avoid concurrent writes
	for _, ci := range targets {
		ci.mu.Lock()
		_ = ci.Conn.WriteMessage(websocket.TextMessage, b)
		ci.mu.Unlock()
	}
}

// EmitToUser sends a JSON payload to all active connections of the user
func EmitToUser(userID int64, payload any) {
	regMu.RLock()
	set := byUser[userID]
	var targets []*connInfo
	for _, ci := range set {
		targets = append(targets, ci)
	}
	regMu.RUnlock()
	emitToTargets(targets, payload)
}

// EmitToAllUsers sends to all logged-in users (with userID != 0)
func EmitToAllUsers(payload any) {
	regMu.RLock()
	var targets []*connInfo
	for _, set := range byUser {
		for _, ci := range set {
			targets = append(targets, ci)
		}
	}
	regMu.RUnlock()
	emitToTargets(targets, payload)
}

// EmitToAny sends to all connections (logged-in or guest)
func EmitToAny(payload any) {
	regMu.RLock()
	var targets []*connInfo
	for _, ci := range byConn {
		targets = append(targets, ci)
	}
	regMu.RUnlock()
	emitToTargets(targets, payload)
}

// EmitToGuests sends only to guest connections (userID == 0)
func EmitToGuests(payload any) {
	regMu.RLock()
	var targets []*connInfo
	for _, ci := range byConn {
		if ci.UserID == 0 {
			targets = append(targets, ci)
		}
	}
	regMu.RUnlock()
	emitToTargets(targets, payload)
}

// === Event helpers ===

func EmitToUserEvent(userID int64, eventType string, data any) {
	EmitToUser(userID, map[string]any{
		"type": eventType,
		"data": data,
		"at":   time.Now().UTC().Format(time.RFC3339),
	})
}

func EmitToAllUsersEvent(eventType string, data any) {
	EmitToAllUsers(map[string]any{
		"type": eventType,
		"data": data,
		"at":   time.Now().UTC().Format(time.RFC3339),
	})
}

func EmitToAnyEvent(eventType string, data any) {
	EmitToAny(map[string]any{
		"type": eventType,
		"data": data,
		"at":   time.Now().UTC().Format(time.RFC3339),
	})
}

func EmitToGuestsEvent(eventType string, data any) {
	EmitToGuests(map[string]any{
		"type": eventType,
		"data": data,
		"at":   time.Now().UTC().Format(time.RFC3339),
	})
}

func EmitServer(req map[string]interface{}, resType string, resData interface{}) {
	log.Println("EmitToAnyEvent called:", resType)
	EmitToAnyEvent("heartbeat", handlers.BattleIndex)
}
