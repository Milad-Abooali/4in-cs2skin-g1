package main

import (
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/handlers"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/web"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/ws"
	"log"
	"net/http"
	"os"

	"github.com/Milad-Abooali/4in-cs2skin-g1/src/configs"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/grpcclient"
	"github.com/joho/godotenv"
)

func init() {
	log.Println("▶ [init] G1 v" + configs.Version)

	// Load env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using defaults")
	}
}

func withAPIVersion(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-API-Version", configs.Version)
		h.ServeHTTP(w, r)
	}
}

func main() {
	_ = godotenv.Load()

	if os.Getenv("DEBUG") == "1" {
		configs.Debug = true
	}

	log.Println("🌐 [main] Core gRPC: ", os.Getenv("CORE_GRPC_ADDRESS"))

	grpcclient.Connect(os.Getenv("CORE_GRPC_ADDRESS"))
	grpcclient.TestConnection()

	// WebSocket
	http.HandleFunc("/ws", ws.HandleWebSocket)

	// HTTP
	http.HandleFunc("/web", withAPIVersion(web.HandleHTTP))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	handlers.FillBattleIndex()
	handlers.FillBots()
	handlers.FillCaseImpact()

	log.Println("Web server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
