package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	// Basic middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Simple HTTP endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// WebSocket endpoint
	r.Get("/ws", wsHandler)

	log.Println("HTTP  : http://localhost:8080/health")
	log.Println("WS    : ws://localhost:8080/ws")

	http.ListenAndServe(":8080", r)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Accept WebSocket connection
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// ⚠️ allow all origins for development
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		log.Println("accept error:", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "bye")

	log.Println("WebSocket client connected")

	ctx := r.Context()

	for {
		// Read message
		msgType, data, err := conn.Read(ctx)
		if err != nil {
			log.Println("read error:", err)
			break
		}

		log.Printf("received: %s\n", data)

		// Echo message back
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = conn.Write(ctx, msgType, data)
		cancel()

		if err != nil {
			log.Println("write error:", err)
			break
		}
	}
}